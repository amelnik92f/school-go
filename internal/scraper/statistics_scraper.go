package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"schools-be/internal/models"

	"github.com/gocolly/colly/v2"
)

const (
	berlinStatisticsURL = "https://www.bildungsstatistik.berlin.de/statistik/ListGen/SVZ_Fakt5.aspx"
)

// StatisticsScraper handles scraping education statistics
type StatisticsScraper struct {
	collector  *colly.Collector
	statistics []models.StatisticData
	logger     *slog.Logger
}

// NewStatisticsScraper creates a new statistics scraper
func NewStatisticsScraper() *StatisticsScraper {

	// Create Colly collector with best practices
	c := colly.NewCollector(
		// Visit only the target domain
		colly.AllowedDomains("www.bildungsstatistik.berlin.de", "bildungsstatistik.berlin.de"),

		// Set User-Agent
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),

		// Cache responses to avoid re-scraping
		colly.CacheDir("./cache/statistics"),
	)

	// Set timeouts
	c.SetRequestTimeout(30 * time.Second)

	// Rate limiting - be respectful to government servers
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*.bildungsstatistik.berlin.de",
		Parallelism: 1,               // One request at a time
		Delay:       2 * time.Second, // 2 seconds between requests
		RandomDelay: 1 * time.Second, // Random delay up to 1 second
	})
	if err != nil {
		slog.Error("failed to set rate limit", slog.String("error", err.Error()))
	}

	scraper := &StatisticsScraper{
		collector:  c,
		statistics: make([]models.StatisticData, 0),
		logger:     slog.Default(),
	}

	// Set up callbacks
	scraper.setupCallbacks()

	return scraper
}

func (s *StatisticsScraper) setupCallbacks() {
	// Before making a request
	s.collector.OnRequest(func(r *colly.Request) {
		s.logger.Info("visiting statistics page", slog.String("url", r.URL.String()))

		// Add headers to mimic real browser
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		r.Headers.Set("Referer", "https://www.bildungsstatistik.berlin.de/")
		r.Headers.Set("Connection", "keep-alive")
	})

	// After receiving response
	s.collector.OnResponse(func(r *colly.Response) {
		s.logger.Info("received response",
			slog.String("url", r.Request.URL.String()),
			slog.Int("status", r.StatusCode),
			slog.Int("size", len(r.Body)),
		)
	})

	// Handle errors
	s.collector.OnError(func(r *colly.Response, err error) {
		s.logger.Error("error scraping statistics",
			slog.String("url", r.Request.URL.String()),
			slog.Int("status", r.StatusCode),
			slog.String("error", err.Error()),
		)
	})

	// Parse the specific data grid table using the correct selector
	s.collector.OnHTML("#myDatagrid", func(e *colly.HTMLElement) {
		s.logger.Info("found myDatagrid table")

		// Extract headers from the first row (tr) with orange background
		var headers []string
		var headerRowFound bool
		e.ForEach("tr", func(i int, row *colly.HTMLElement) {
			// First row has the headers (bgcolor="#F39300")
			if i == 0 || row.Attr("bgcolor") == "#F39300" {
				row.ForEach("td", func(_ int, cell *colly.HTMLElement) {
					header := strings.TrimSpace(cell.Text)
					headers = append(headers, header)
				})
				headerRowFound = true
				return
			}
		})

		if !headerRowFound || len(headers) == 0 {
			s.logger.Warn("no headers found in myDatagrid table")
			return
		}

		s.logger.Info("table headers", slog.Int("count", len(headers)), slog.Any("headers", headers))

		// Parse all data rows (skip the first header row)
		rowCount := 0
		e.ForEach("tr", func(i int, row *colly.HTMLElement) {
			// Skip header row (first row or row with orange background)
			if i == 0 || row.Attr("bgcolor") == "#F39300" {
				return
			}

			stat := models.StatisticData{
				Metadata:  make(map[string]string),
				ScrapedAt: time.Now(),
			}

			cells := []string{}
			row.ForEach("td", func(cellIndex int, cell *colly.HTMLElement) {
				value := strings.TrimSpace(cell.Text)
				cells = append(cells, value)

				// Map to metadata
				if cellIndex < len(headers) && headers[cellIndex] != "" {
					stat.Metadata[headers[cellIndex]] = value
				}

				// Map to known fields based on actual column names from the website
				if cellIndex < len(headers) {
					header := strings.ToLower(headers[cellIndex])
					switch {
					case header == "bsn":
						stat.SchoolNumber = value
					case header == "name":
						stat.SchoolName = value
					case header == "schuljahr":
						stat.SchoolYear = value
					case strings.Contains(header, "schüler (m/w/d)") || strings.Contains(header, "schueler (m/w/d)"):
						stat.Students = value
					case strings.Contains(header, "schüler (w)") || strings.Contains(header, "schueler (w)"):
						stat.StudentsFemale = value
					case strings.Contains(header, "schüler (m)") || strings.Contains(header, "schueler (m)"):
						stat.StudentsMale = value
					case strings.Contains(header, "lehrkräfte (m,w,d)") || strings.Contains(header, "lehrkraefte (m,w,d)"):
						stat.Teachers = value
					case strings.Contains(header, "lehrkräfte (w)") || strings.Contains(header, "lehrkraefte (w)"):
						stat.TeachersFemale = value
					case strings.Contains(header, "lehrkräfte (m)") || strings.Contains(header, "lehrkraefte (m)"):
						stat.TeachersMale = value
					case header == "bezirk", header == "district":
						stat.District = value
					case header == "schulart", header == "school type":
						stat.SchoolType = value
					case header == "klassen", header == "classes":
						stat.Classes = value
					}
				}
			})

			// Only add if we got meaningful data
			if len(cells) > 0 && stat.SchoolNumber != "" {
				s.statistics = append(s.statistics, stat)
				rowCount++
			}
		})

		s.logger.Info("parsed table rows", slog.Int("count", rowCount))
	})

	// When scraping is complete
	s.collector.OnScraped(func(r *colly.Response) {
		s.logger.Info("finished scraping",
			slog.String("url", r.Request.URL.String()),
			slog.Int("statistics_count", len(s.statistics)),
		)
	})
}

// ScrapeStatistics scrapes statistics from the website and returns them
func (s *StatisticsScraper) ScrapeStatistics(ctx context.Context) ([]models.StatisticData, error) {
	s.logger.Info("starting statistics scrape", slog.String("url", berlinStatisticsURL))

	// Reset statistics
	s.statistics = make([]models.StatisticData, 0)

	// Visit the page
	if err := s.collector.Visit(berlinStatisticsURL); err != nil {
		return nil, fmt.Errorf("failed to visit URL: %w", err)
	}

	// Wait for async operations
	s.collector.Wait()

	if len(s.statistics) == 0 {
		s.logger.Warn("no statistics found")
		return nil, fmt.Errorf("no statistics found")
	}

	s.logger.Info("scraping complete", slog.Int("statistics", len(s.statistics)))

	return s.statistics, nil
}
