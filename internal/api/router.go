package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
	"district-friends/internal/api/handlers"
	_ "district-friends/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler := handlers.NewUserHandler(db)
	friendHandler := handlers.NewFriendHandler(db)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
			"message": "Friends District API is running",
		})
	})

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
		}

		friends := v1.Group("/friends")
		{
			friends.POST("/request", friendHandler.SendFriendRequest)
			friends.POST("/accept", friendHandler.AcceptFriendRequest)
			friends.GET("", friendHandler.ListFriends)
		}
	}

	return r
}
