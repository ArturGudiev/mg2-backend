package main

import (
	"log"
	"os"

	"arturgudiev/memoryguard/app"
	"arturgudiev/memoryguard/auth"
	"arturgudiev/memoryguard/docs"
	"arturgudiev/memoryguard/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           MemoryGuard API
// @version         1.0
// @description     Memory nodes, cards, and card items API server.
// @description     **Authentication:** Click **Authorize**, enter your login or email as username and your password, then Authorize. Swagger will call `POST /users/token` and attach the access token as a Bearer header.

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3033
// @BasePath  /

// @schemes   http https

// @securitydefinitions.oauth2.password Login
// @tokenUrl /users/token
// @scope.api Access the MemoryGuard API

// @security Login[api]

func main() {
	_ = godotenv.Load()

	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer application.Close()

	// Swagger UI uses docs.SwaggerInfo.Host for "Try it out" requests.
	// Default empty host = same origin as the Swagger page (works with any published port).
	// Override with SWAGGER_HOST, e.g. "158.160.36.7:3033".
	if host := os.Getenv("SWAGGER_HOST"); host != "" {
		docs.SwaggerInfo.Host = host
	} else {
		docs.SwaggerInfo.Host = ""
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOriginFunc = func(origin string) bool {
		return true
	}
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"Accept",
		"X-Requested-With",
		"userId",
	}
	config.MaxAge = 86400
	router.Use(cors.New(config))

	router.Use(auth.AuthMiddleware())

	router.OPTIONS("/*path", func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, userId")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Status(204)
	})

	// Swagger UI route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(c *ginSwagger.Config) {
		c.PersistAuthorization = true
		c.Oauth2DefaultClientID = "swagger"
	}))

	h := handlers.NewHandler(application)

	router.GET("/", h.Root)

	// User routes
	router.GET("/users/me", h.GetMe)
	router.GET("/users/:id", h.GetUser)
	router.POST("/users", h.AddUser)
	router.POST("/users/login", h.LoginUser)
	router.POST("/users/token", h.IssueToken)
	router.POST("/users/refresh", h.RefreshUserToken)
	router.POST("/users/logout", h.LogoutUser)

	// Memory nodes
	router.GET("/memory-node/:id", h.GetMemoryNodeByID)
	router.GET("/memory-node/:id/parents-path", h.GetMemoryNodeParentsPath)
	router.GET("/memory-nodes/roots", h.ListRootMemoryNodes)
	router.GET("/memory-nodes", h.ListMemoryNodes)
	router.POST("/get-memory-nodes", h.GetMemoryNodesByIDs)
	router.GET("/memory-node-by-alias/:alias", h.GetMemoryNodeByAlias)
	router.GET("/node-by-alias/:alias", h.GetMemoryNodeByAlias)
	router.POST("/new-memory-node", h.NewMemoryNode)
	router.PUT("/update-memory-node", h.UpdateMemoryNode)
	router.DELETE("/memory-node/:id", h.DeleteMemoryNode)
	router.POST("/memory-node/:id/cards", h.BulkNewTextCards)
	router.POST("/memory-node/:id/users", h.GrantMemoryNodeAccess)
	router.POST("/admin/memory-node/:id/move-to-user", h.MoveSharedNodeToUser)
	router.POST("/admin/memory-node/:id/remove-from-user", h.RemoveSharedNodeFromUser)

	// Cards
	router.GET("/card/:id", h.GetCardByID)
	router.GET("/cards", h.ListCards)
	router.POST("/get-cards", h.GetCardsByIDs)
	router.POST("/new-card", h.NewCard)
	router.PUT("/update-card", h.UpdateCard)
	router.POST("/update-cards-field", h.UpdateCardsField)
	router.POST("/delete-cards", h.DeleteCards)
	router.POST("/cards-by-query", h.CardsByQuery)
	router.PUT("/increase-card-count/:id", h.IncreaseCardCount)
	router.PUT("/decrease-card-count/:id", h.DecreaseCardCount)

	// Card items
	router.GET("/card-item/:id", h.GetCardItemByID)
	router.GET("/card-items", h.ListCardItems)
	router.POST("/get-card-items", h.GetCardItemsByIDs)
	router.POST("/new-card-item", h.NewCardItem)
	router.PUT("/update-card-item", h.UpdateCardItem)
	router.DELETE("/card-item/:id", h.DeleteCardItem)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3033"
	}
	log.Printf("MemoryGuard server listening on :%s", port)
	log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
