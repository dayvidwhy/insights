package views

import (
	"errors"
	db "insights/db"
	"insights/lib/auth"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Response type for counting views
type ViewCountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

type ViewsCountFetch struct {
	Status string `json:"status"`
	Views  int    `json:"views"`
	Url    string `json:"url"`
}

type PageView struct {
	Time string `json:"time"`
}

type PageViews []PageView

type ViewsCountFetchByDate struct {
	Status string    `json:"status"`
	Start  string    `json:"start"`
	End    string    `json:"end"`
	Url    string    `json:"url"`
	Views  PageViews `json:"views"`
}

type PageViewSubmit struct {
	Url string `json:"url"`
}

// Receive page views from clients
func IncrementViewCounts(c echo.Context) error {
	token, err := auth.TokenAuth(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	u := new(PageViewSubmit)
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	if u.Url == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "URL is empty.")
	}

	// Get the account ID from the token
	accountId, err := auth.GetAccountIdFromToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	log.Println("Tracking URL: " + u.Url)
	incrementPageView(accountId, u.Url)

	return c.JSON(http.StatusOK, &ViewCountResponse{
		Status:  "success",
		Message: "URL has been tracked.",
		Url:     u.Url,
	})
}

// Returns the total number of views for a given URL
func GetViewCountForUrl(c echo.Context) error {
	accountId, err := auth.GetAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	url := c.QueryParam("url")
	pageViews := fetchPageViews(accountId, url)

	return c.JSON(http.StatusOK, &ViewsCountFetch{
		Status: "success",
		Views:  pageViews,
		Url:    url,
	})
}

// Returns a list of views between two dates
func GetViewsForUrlInRange(c echo.Context) error {
	accountId, err := auth.GetAccountIdFromJwt(c)
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
	pageViews, err := fetchPageViewsByDate(accountId, url, start, end)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching page views by date.")
	}

	return c.JSON(http.StatusOK, &ViewsCountFetchByDate{
		Status: "success",
		Start:  start,
		End:    end,
		Url:    url,
		Views:  pageViews,
	})
}

// Setup table to store overall pageviews
func SetupViews() {
	_, err := db.Database.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		url TEXT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		UNIQUE(accountId, url)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Database.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS page_views_accountId_url
		ON page_views (accountId, url)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Setup table to store individual pageviews
	_, err = db.Database.Exec(`CREATE TABLE IF NOT EXISTS page_views_individual (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		url TEXT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

// Increment the page view count for a URL
func incrementPageView(accountId int, url string) error {
	_, err := db.Database.Exec(`
		INSERT INTO page_views (accountId, url, count)
		VALUES ($1, $2, 1) ON CONFLICT (accountId, url)
		DO UPDATE SET count = page_views.count + 1`, accountId, url)
	if err != nil {
		return err
	}

	_, err = db.Database.Exec(`
		INSERT INTO page_views_individual (accountId, url)
		VALUES ($1, $2)
	`, accountId, url)

	return err
}

// Retrieve page views for a given URL, return 0 count if not found
func fetchPageViews(accountId int, url string) int {
	log.Println("Fetch pageviews for: " + url)
	row := db.Database.QueryRow(`
		SELECT count
		FROM page_views
		WHERE url = $1
		AND accountId = $2`, url, accountId)

	var count int
	err := row.Scan(&count)

	// If we don't find a record, return 0
	if err != nil {
		return 0
	}

	return count
}

type PageViewCount struct {
	Url   string `json:"url"`
	Count int    `json:"count"`
}

type PageViewCounts []PageViewCount

type AllPageViewCountsResponse struct {
	Status string         `json:"status"`
	Views  PageViewCounts `json:"views"`
}

func fetchAllViews(accountId int) (PageViewCounts, error) {
	var pageViewCounts PageViewCounts
	rows, err := db.Database.Query(`
		SELECT url, count
		FROM page_views
		WHERE accountId = $1
	`, accountId)
	if err != nil {
		return pageViewCounts, errors.New("failed to retrieve page views")
	}

	defer rows.Close()

	for rows.Next() {
		var pageViewCount PageViewCount
		err := rows.Scan(&pageViewCount.Url, &pageViewCount.Count)
		if err != nil {
			log.Println("Error fetching all views: ", err)
			return nil, err
		}
		pageViewCounts = append(pageViewCounts, pageViewCount)
	}

	return pageViewCounts, nil
}

// Fetch all urls tracked and associated view counts
func GetAllViews(c echo.Context) error {
	accountId, err := auth.GetAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	allPageViews, err := fetchAllViews(accountId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Issue fetching all views")
	}

	return c.JSON(http.StatusOK, &AllPageViewCountsResponse{
		Status: "success",
		Views:  allPageViews,
	})
}

func fetchPageViewsByDate(
	accountId int,
	url string,
	start string,
	end string,
) (
	PageViews,
	error,
) {
	var pageViews PageViews
	rows, err := db.Database.Query(`
		SELECT createdAt
		FROM page_views_individual
		WHERE url = $1
		AND createdAt BETWEEN $2 AND $3
		AND accountId = $4
	`, url, start, end, accountId)

	if err != nil {
		return pageViews, err
	}

	// defer closing until we're done with the rows
	defer rows.Close()

	for rows.Next() {
		var pageView PageView
		err := rows.Scan(&pageView.Time)
		if err != nil {
			log.Println("Error fetching page views by date: ", err)
			return nil, err
		}
		pageViews = append(pageViews, pageView)
	}
	return pageViews, nil
}
