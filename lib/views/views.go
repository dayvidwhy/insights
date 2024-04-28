package views

import (
	"insights/lib/auth"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type ViewsHandler struct {
	store        *ViewsStore
	tokenHandler *auth.TokenHandler
}

// Response type for counting views
type ViewCountResponse struct {
	Message string `json:"message"`
	Url     string `json:"url"`
}

type ViewsCountFetch struct {
	Views int    `json:"views"`
	Url   string `json:"url"`
}

type PageView struct {
	Time string `json:"time"`
}

type PageViews []PageView

type ViewsCountFetchByDate struct {
	Start string    `json:"start"`
	End   string    `json:"end"`
	Url   string    `json:"url"`
	Views PageViews `json:"views"`
}

type PageViewSubmit struct {
	Url string `json:"url"`
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

func NewViews(store *ViewsStore, tokenHandler *auth.TokenHandler) *ViewsHandler {
	return &ViewsHandler{
		store:        store,
		tokenHandler: tokenHandler,
	}
}

// Receive page views from clients
func (vh *ViewsHandler) IncrementViewCounts(c echo.Context) error {
	token, err := vh.tokenHandler.TokenAuth(c)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	// Get the account ID from the token
	accountId, err := vh.tokenHandler.GetAccountId(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	u := new(PageViewSubmit)
	if err := c.Bind(u); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	if u.Url == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "URL is empty.")
	}

	log.Println("Tracking URL: " + u.Url)

	if err := vh.store.incrementPageView(accountId, u.Url); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error recording page view.")
	}

	return c.JSON(http.StatusOK, &ViewCountResponse{
		Message: "URL has been tracked.",
		Url:     u.Url,
	})
}

// Returns the total number of views for a given URL
func (vh *ViewsHandler) GetViewCountForUrl(c echo.Context) error {
	accountId, err := auth.GetAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	url := c.QueryParam("url")
	pageViews := vh.store.fetchPageViews(accountId, url)

	return c.JSON(http.StatusOK, &ViewsCountFetch{
		Views: pageViews,
		Url:   url,
	})
}

// Returns a list of views between two dates
func (vh *ViewsHandler) GetViewsForUrlInRange(c echo.Context) error {
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
	pageViews, err := vh.store.fetchPageViewsByDate(accountId, url, start, end)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching page views by date.")
	}

	return c.JSON(http.StatusOK, &ViewsCountFetchByDate{
		Start: start,
		End:   end,
		Url:   url,
		Views: pageViews,
	})
}

// Fetch all urls tracked and associated view counts
func (vh *ViewsHandler) GetAllViews(c echo.Context) error {
	accountId, err := auth.GetAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	allPageViews, err := vh.store.fetchAllViews(accountId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Issue fetching all views")
	}

	return c.JSON(http.StatusOK, &AllPageViewCountsResponse{
		Status: "success",
		Views:  allPageViews,
	})
}
