package views

import (
	"database/sql"
	"errors"
	"log"
)

type ViewsStore struct {
	db *sql.DB
}

// Setup table to store overall pageviews
func SetupViews(db *sql.DB) *ViewsStore {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		url TEXT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		UNIQUE(accountId, url)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS page_views_accountId_url
		ON page_views (accountId, url)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Setup table to store individual pageviews
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views_individual (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		url TEXT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	return &ViewsStore{db: db}
}

// Increment the page view count for a URL
func (vs *ViewsStore) incrementPageView(accountId int, url string) error {
	_, err := vs.db.Exec(`
		INSERT INTO page_views (accountId, url, count)
		VALUES ($1, $2, 1) ON CONFLICT (accountId, url)
		DO UPDATE SET count = page_views.count + 1`, accountId, url)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = vs.db.Exec(`
		INSERT INTO page_views_individual (accountId, url)
		VALUES ($1, $2)
	`, accountId, url)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Retrieve page views for a given URL, return 0 count if not found
func (vs *ViewsStore) fetchPageViews(accountId int, url string) int {
	log.Println("Fetch pageviews for: " + url)
	var count int
	err := vs.db.QueryRow(`
		SELECT count
		FROM page_views
		WHERE url = $1
		AND accountId = $2`,
		url, accountId).Scan(&count)

	// If we don't find a record, return 0
	if err != nil {
		return 0
	}

	return count
}

func (vs *ViewsStore) fetchAllViews(accountId int) (PageViewCounts, error) {
	var pageViewCounts PageViewCounts
	rows, err := vs.db.Query(`
		SELECT url, count
		FROM page_views
		WHERE accountId = $1
	`, accountId)
	if err != nil {
		log.Println(err)
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

func (vs *ViewsStore) fetchPageViewsByDate(
	accountId int,
	url string,
	start string,
	end string,
) (
	PageViews,
	error,
) {
	var pageViews PageViews
	rows, err := vs.db.Query(`
		SELECT createdAt
		FROM page_views_individual
		WHERE url = $1
		AND createdAt BETWEEN $2 AND $3
		AND accountId = $4
	`, url, start, end, accountId)

	if err != nil {
		log.Println(err)
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
