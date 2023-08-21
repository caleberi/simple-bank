package api

import (
	db "github.com/caleberi/simple-bank/db/sqlc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	store  *db.SQLStore
	router *gin.Engine
}

func NewServer(store *db.SQLStore) *Server {

	server := &Server{store: store}
	router := gin.Default()
	server.router = router

	router.POST("/accounts", server.createAccountHandler)
	router.GET("/accounts/:id", server.getAccountHandler)
	router.GET("/accounts", server.listAccountHandler)
	router.DELETE("/accounts/:id", server.deleteAccount)

	return server
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
