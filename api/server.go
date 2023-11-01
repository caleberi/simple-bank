package api

import (
	"fmt"

	db "github.com/caleberi/simple-bank/db/sqlc"
	"github.com/caleberi/simple-bank/pkg/utils"
	"github.com/caleberi/simple-bank/token"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config         utils.Config
	store          db.Store
	tokenGenerator token.Maker
	router         *gin.Engine
}

func NewServer(config utils.Config, store db.Store) (*Server, error) {
	tokenGenerator, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:         config,
		store:          store,
		tokenGenerator: tokenGenerator,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.registerRoutes()

	return server, nil
}

func (server *Server) registerRoutes() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenGenerator))

	authRoutes.POST("/accounts", server.createAccountHandler)
	authRoutes.GET("/accounts/:id", server.getAccountHandler)
	authRoutes.GET("/accounts", server.listAccountHandler)
	authRoutes.DELETE("/accounts/:id", server.deleteAccount)
	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router

}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	res := gin.H{}
	res["success"] = false
	res["error"] = err.Error()
	return res
}

func successResponse(message string, data interface{}) gin.H {
	res := gin.H{}
	res["success"] = true
	res["message"] = message
	res["data"] = data
	return res
}
