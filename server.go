package main

import (
	database "insights/db"
	accounts "insights/lib/accounts"
	auth "insights/lib/auth"
	views "insights/lib/views"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// Extract the account ID from the JWT token
func getAccountIdFromJwt(c echo.Context) (int, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	accountId := int(claims["accountId"].(float64))
	return accountId, nil
}

// Receive page views from clients
func createViews(c echo.Context) error {
	token, err := auth.TokenAuth(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	u := new(views.PageViewSubmit)
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	if u.Url == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "URL is empty.")
	}

	// Get the account ID from the token
	accountId, err := accounts.GetAccountIdFromToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	log.Println("Tracking URL: " + u.Url)
	views.IncrementPageView(accountId, u.Url)

	return c.JSON(http.StatusOK, &views.ViewCountResponse{
		Status:  "success",
		Message: "URL has been tracked.",
		Url:     u.Url,
	})
}

// Returns the total number of views for a given URL
func countViews(c echo.Context) error {
	accountId, err := getAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	url := c.QueryParam("url")
	pageViews := views.FetchPageViews(accountId, url)

	return c.JSON(http.StatusOK, &views.ViewsCountFetch{
		Status: "success",
		Views:  pageViews,
		Url:    url,
	})
}

// Returns a list of views between two dates
func viewCounts(c echo.Context) error {
	accountId, err := getAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	url := c.QueryParam("url")
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	_, err = time.Parse("2006-01-02 15:04:05.000", start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid start date:"+start+". Please use UTC in the format: yyyy-mm-dd hh:mm:ss.fff")
	}
	_, err = time.Parse("2006-01-02 15:04:05.000", end)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid start date:"+end+". Please use UTC in the format: yyyy-mm-dd hh:mm:ss.fff")
	}
	pageViews, err := views.FetchPageViewsByDate(accountId, url, start, end)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching page views by date.")
	}

	return c.JSON(http.StatusOK, &views.ViewsCountFetchByDate{
		Status: "success",
		Start:  start,
		End:    end,
		Url:    url,
		Views:  pageViews,
	})
}

// Create a new user account
func createAccount(c echo.Context) error {
	u := new(accounts.User)
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	if u.Email == "" || u.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email or password is empty.")
	}

	err := accounts.CreateUserAccount(u.Email, u.Password)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating account.")
	}

	return c.JSON(http.StatusOK, &accounts.AccountResponse{
		Status:  "success",
		Message: "Account has been created.",
	})
}

// Validate user credentials and return a JWT token
func loginUser(c echo.Context) error {
	u := new(accounts.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	accountId, err := accounts.LogInUser(u.Email, u.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials.")
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = u.Email
	claims["accountId"] = accountId
	claims["exp"] = time.Now().Add(time.Second * 72).Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": t,
	})
}

// Generate a client side access token
func getAccessToken(c echo.Context) error {
	accountId, err := getAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	token, err := accounts.CreateAccessToken(accountId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating token.")
	}

	return c.JSON(http.StatusOK, &accounts.AccountTokenResponse{
		Status:  "success",
		Message: "Authorized",
		Token:   token,
	})
}

func main() {
	e := echo.New()
	database.SetupDb()
	views.SetupViews()
	accounts.SetupAccounts()

	// Public user routes
	e.POST("/register", createAccount)
	e.POST("/login", loginUser)

	// Access token routes
	e.POST("/create-view", createViews)

	// Protected routes
	r := e.Group("/api")
	r.Use(echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	r.GET("/views/count", countViews)
	r.GET("/views/counts", viewCounts)
	r.GET("/accounts/token", getAccessToken)

	e.Logger.Fatal(e.Start(":1323"))
}
