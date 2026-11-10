package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/helpdesk/internal/handler"
	"github.com/yourname/helpdesk/internal/middleware"
	"github.com/yourname/helpdesk/internal/service"
)

func SetupRouter(
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
	ticketHandler *handler.TicketHandler,
	commentHandler *handler.CommentHandler,
	userHandler *handler.UserHandler,
) *gin.Engine {
	engine := gin.Default()
	engine.Use(middleware.CORS())
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	protected.GET("/me", userHandler.Me)
	protected.POST("/tickets", ticketHandler.Create)
	protected.GET("/tickets", ticketHandler.List)
	protected.GET("/tickets/:id", ticketHandler.Get)
	protected.PATCH("/tickets/:id", ticketHandler.Update)
	protected.PATCH("/tickets/:id/status", ticketHandler.UpdateStatus)
	protected.GET("/tickets/:id/comments", commentHandler.List)
	protected.POST("/tickets/:id/comments", commentHandler.Create)
	protected.GET("/users", middleware.RequireRole(service.RoleAdmin, service.RoleAgent), userHandler.List)

	admin := protected.Group("")
	admin.Use(middleware.RequireRole("admin"))
	admin.PATCH("/users/:id/role", userHandler.UpdateRole)

	return engine
}
