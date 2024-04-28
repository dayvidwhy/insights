package main

import (
	database "insights/db"
	accounts "insights/lib/accounts"
	auth "insights/lib/auth"
	views "insights/lib/views"

	_ "github.com/joho/godotenv/autoload"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// Instantiate DB handler
	db := database.SetupDb()

	// Setup stores that interact with DB
	authStore := auth.SetupAuth(db)
	tokenStore := auth.SetupTokens(db)
	viewsStore := views.SetupViews(db)
	accountsStore := accounts.SetupAccounts(db)

	// Pass stores to endpoint handlers
	authHandler := auth.NewAuth(authStore)
	tokenHandler := auth.NewTokens(tokenStore)
	viewsHandler := views.NewViews(viewsStore, tokenHandler)
	accountsHandler := accounts.NewAccounts(accountsStore)

	// Public user routes
	e.POST("/register", accountsHandler.CreateAccount)
	e.POST("/login", authHandler.LoginUser)

	// Access token routes
	e.POST("/create-view", viewsHandler.IncrementViewCounts)

	jwtMiddleware := auth.GetJwtMiddleware()

	// Protected routes
	viewRoutes := e.Group("/views")
	viewRoutes.Use(jwtMiddleware)
	viewRoutes.GET("/count", viewsHandler.GetViewCountForUrl)
	viewRoutes.GET("/counts", viewsHandler.GetViewsForUrlInRange)
	viewRoutes.GET("/all", viewsHandler.GetAllViews)

	// Token routes
	tokens := e.Group("/tokens")
	tokens.Use(jwtMiddleware)
	tokens.GET("/create", tokenHandler.GetAccessToken)
	tokens.DELETE("/revoke", tokenHandler.RevokeAccessToken)

	e.Logger.Fatal(e.Start(":1323"))
}
