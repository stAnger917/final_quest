package server

import (
	"final_quest/internal/usecase/users"
	"final_quest/pkg/authmw"
	"final_quest/pkg/logging"
	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	userService *users.Users
	logger      *logging.Logger
}

func InitAppHandler(usersUseCase *users.Users, logger *logging.Logger) *AppHandler {
	return &AppHandler{
		userService: usersUseCase,
		logger:      logger,
	}
}

func (h *AppHandler) Init() *gin.Engine {
	router := gin.Default()
	router.Use(gin.Recovery())
	userRoutes := router.Group("/api/user")
	{
		userRoutes.POST("/register", h.UserRegistration)
		userRoutes.POST("/login", h.UserLogin)
		userRoutes.POST("/orders", authmw.TokenMW(), h.PostOrders)
		userRoutes.GET("/orders", authmw.TokenMW(), h.GetOrders)
		userRoutes.GET("/balance", authmw.TokenMW(), h.GetBalance)
		userRoutes.POST("/balance/withdraw", authmw.TokenMW(), h.PostWithdraw)
		userRoutes.GET("/balance/withdrawals", authmw.TokenMW(), h.GetWithdrawals)
	}
	return router
}
