package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/caleberi/simple-bank/db/sqlc"
	"github.com/caleberi/simple-bank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

var (
	ErrForeginKeyViolation = "foreign_key_violation"
	ErrUniqueViolation     = "unique_violation"
)

type createAccountRequest struct {
	CurrencyCode string `json:"currency_code" binding:"required,oneof=USD EUR GBP NGN AUD CAD CDF"`
}

func (server *Server) createAccountHandler(ctx *gin.Context) {
	var request createAccountRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountParams{
		Owner:        authPayload.Username,
		CurrencyCode: request.CurrencyCode,
		Balance:      0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if perr, ok := err.(*pq.Error); ok {
			switch perr.Code.Name() {
			case ErrForeginKeyViolation, ErrUniqueViolation:
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, successResponse("account created successfully", account))
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccountHandler(ctx *gin.Context) {
	var request getAccountRequest

	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, request.ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, fmt.Errorf("account with ID [%d] does not exist", request.ID))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("retrieved account successfully", account))
}

type listAccountsRequest struct {
	PageID   int32 `json:"page_id" binding:"required,min=1"`
	PageSize int32 `json:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccountHandler(ctx *gin.Context) {
	var request listAccountsRequest

	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	offset := (request.PageID - 1) * request.PageSize
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Offset: offset,
		Limit:  request.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, successResponse(
		fmt.Sprintf("retrieved accounts from offset %d with size %d",
			offset, request.PageSize),
		accounts))
}

type deleteAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAccount(ctx *gin.Context) {
	var request deleteAccountRequest

	if err := ctx.ShouldBindUri(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteAccount(ctx, request.ID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse(fmt.Sprintf("deleted account with id (%d) successfully", request.ID), nil))
}
