package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "district-friends/docs"
	"district-friends/internal/api/handlers"
	"district-friends/internal/ws"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *gorm.DB, hub *ws.Hub) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userHandler := handlers.NewUserHandler(db)
	friendHandler := handlers.NewFriendHandler(db)
	chatHandler := handlers.NewChatHandler(db, hub)
	eventHandler := handlers.NewEventHandler()
	bookingHandler := handlers.NewBookingHandler(db)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "up",
			"message": "Friends District API is running",
		})
	})

	v1 := r.Group("/api/v1")
	// Profile Routes
	profile := v1.Group("/profile")
	{
		profile.POST("", userHandler.CreateProfile)
		profile.GET("", userHandler.GetProfile)
	}

	// Friend Routes
	friends := v1.Group("/friends")
	{
		friends.POST("/request", friendHandler.SendFriendRequest)
		friends.POST("/accept", friendHandler.AcceptFriendRequest)
		friends.POST("/reject", friendHandler.RejectFriendRequest)
		friends.GET("", friendHandler.ListFriends)
		friends.GET("/status", friendHandler.GetFriendStatus)
		friends.GET("/requests/sent", friendHandler.ListSentFriendRequests)
		friends.GET("/requests/received", friendHandler.ListReceivedFriendRequests)
	}

	{
		rooms := v1.Group("/rooms")
		{
			rooms.GET("", chatHandler.ListJoinedRooms)
			rooms.GET("/invites", chatHandler.ListRoomInvites)
			rooms.POST("", chatHandler.CreateRoom)
			rooms.POST("/:id/join", chatHandler.JoinRoom)
			rooms.POST("/:id/invite", chatHandler.InviteToRoom)
			rooms.POST("/:id/accept", chatHandler.AcceptRoomInvite)
			rooms.POST("/:id/reject", chatHandler.RejectRoomInvite)
			rooms.GET("/:id/messages", chatHandler.GetMessages)
			rooms.GET("/:id/members", chatHandler.ListRoomMembers)
			rooms.GET("/:id/ws", chatHandler.ServeWS)
			rooms.POST("/:id/share", chatHandler.ShareEvent)
		}

		events := v1.Group("/events")
		{
			events.GET("/spotlight", eventHandler.GetSpotlight)
			events.GET("", eventHandler.ListEvents)
			events.GET("/:id", eventHandler.GetEvent)
		}

		bookings := v1.Group("/bookings")
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.GET("", bookingHandler.ListBookings)
		}
	}

	return r
}
