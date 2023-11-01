package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/caleberi/simple-bank/db/sqlc"
	"github.com/caleberi/simple-bank/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	CurrencyCode  string `json:"currency_code" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var request transferRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, request.FromAccountID, request.CurrencyCode)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_, valid = server.validAccount(ctx, request.ToAccountID, request.CurrencyCode)
	if !valid {
		return
	}

	arg := db.TransferTxnParams{
		FromAccountID: request.FromAccountID,
		ToAccountID:   request.ToAccountID,
		Amount:        request.Amount,
	}

	result, err := server.store.PerformTransactionTrxn(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("transaction initiated successfully", result))
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currencyCode string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return account, false
	}

	if account.CurrencyCode != currencyCode {
		err := fmt.Errorf("account [%d] currency mismatch: %v vs %s", account.ID, account.CurrencyCode, currencyCode)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
