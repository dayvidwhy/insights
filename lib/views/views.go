package views

import (
	db "insights/db"
	"log"
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

type PageViews []struct {
	Time string `json:"time"`
}

type ViewsCountFetchByDate struct {
	Status string `json:"status"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Url    string `json:"url"`
	Views  PageViews
}

type PageViewSubmit struct {
	Url string `json:"url"`
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
func IncrementPageView(accountId int, url string) error {
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
func FetchPageViews(accountId int, url string) int {
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

func FetchPageViewsByDate(
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
