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
	database.SetupDb()
	views.SetupViews()
	accounts.SetupAccounts()

	// Public user routes
	e.POST("/register", accounts.CreateAccount)
	e.POST("/login", auth.LoginUser)

	// Access token routes
	e.POST("/create-view", views.IncrementViewCounts)

	jwtConfig := auth.GetJwtMiddleware()

	// Protected routes
	viewRoutes := e.Group("/views")
	viewRoutes.Use(jwtConfig)
	viewRoutes.GET("/count", views.GetViewCountForUrl)
	viewRoutes.GET("/counts", views.GetViewsForUrlInRange)

	// Token routes
	tokens := e.Group("/tokens")
	tokens.Use(jwtConfig)
	tokens.GET("/create", auth.GetAccessToken)
	tokens.POST("/revoke", auth.RevokeAccessToken)

	e.Logger.Fatal(e.Start(":1323"))
}
