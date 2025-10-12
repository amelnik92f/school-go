package scraper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"schools-be/internal/models"

	"golang.org/x/net/html"

	"github.com/chromedp/chromedp"
)

const (
	berlinSchoolListURL = "https://www.bildung.berlin.de/Schulverzeichnis/SchulListe.aspx"
	cacheDir            = "./cache/school-details"
)

// SchoolDetailsScraper handles scraping detailed school information
type SchoolDetailsScraper struct {
	logger   *slog.Logger
	cacheDir string
	useCache bool
}

// NewSchoolDetailsScraper creates a new school details scraper
func NewSchoolDetailsScraper() *SchoolDetailsScraper {
	return &SchoolDetailsScraper{
		logger:   slog.Default(),
		cacheDir: cacheDir,
		useCache: true,
	}
}

// NewSchoolDetailsScraperWithCache creates a new school details scraper with cache control
func NewSchoolDetailsScraperWithCache(useCache bool) *SchoolDetailsScraper {
	scraper := NewSchoolDetailsScraper()
	scraper.useCache = useCache
	return scraper
}

// ensureCacheDir creates the cache directory if it doesn't exist
func (s *SchoolDetailsScraper) ensureCacheDir() error {
	return os.MkdirAll(s.cacheDir, 0755)
}

// getCacheKey generates a cache key (filename) for a given URL
func (s *SchoolDetailsScraper) getCacheKey(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// getCachePath returns the full path to the cache file for a URL
func (s *SchoolDetailsScraper) getCachePath(url string) string {
	key := s.getCacheKey(url)
	// Create subdirectories based on first 2 characters of hash to avoid too many files in one directory
	subdir := key[:2]
	return filepath.Join(s.cacheDir, subdir, key+".json")
}

// loadFromCache attempts to load cached data for a URL
func (s *SchoolDetailsScraper) loadFromCache(url string) (*models.SchoolDetailData, bool) {
	if !s.useCache {
		return nil, false
	}

	cachePath := s.getCachePath(url)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		// Cache miss or error reading
		return nil, false
	}

	var details models.SchoolDetailData
	if err := json.Unmarshal(data, &details); err != nil {
		s.logger.Warn("failed to unmarshal cached data",
			slog.String("cache_path", cachePath),
			slog.String("error", err.Error()),
		)
		return nil, false
	}

	s.logger.Info("loaded from cache", slog.String("url", url))
	return &details, true
}

// saveToCache saves scraped data to cache
func (s *SchoolDetailsScraper) saveToCache(url string, details *models.SchoolDetailData) error {
	if !s.useCache {
		return nil
	}

	cachePath := s.getCachePath(url)

	// Ensure the subdirectory exists
	cacheSubdir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheSubdir, 0755); err != nil {
		return fmt.Errorf("failed to create cache subdirectory: %w", err)
	}

	data, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// ClearCache removes all cached data
func (s *SchoolDetailsScraper) ClearCache() error {
	if err := os.RemoveAll(s.cacheDir); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	s.logger.Info("cache cleared", slog.String("cache_dir", s.cacheDir))
	return nil
}

// ScrapeSchoolDetails scrapes detailed information for all schools
func (s *SchoolDetailsScraper) ScrapeSchoolDetails(ctx context.Context) ([]models.SchoolDetailData, error) {
	s.logger.Info("starting school details scrape", slog.String("url", berlinSchoolListURL))

	// Ensure cache directory exists
	if err := s.ensureCacheDir(); err != nil {
		s.logger.Warn("failed to create cache directory", slog.String("error", err.Error()))
	}

	// Create chrome context
	allocCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Get list of school URLs
	schoolLinks, err := s.getSchoolLinks(allocCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get school links: %w", err)
	}

	s.logger.Info("found schools", slog.Int("count", len(schoolLinks)))

	// Scrape each school's details
	var allDetails []models.SchoolDetailData
	cachedCount := 0
	scrapedCount := 0

	for i, link := range schoolLinks {
		s.logger.Info("processing school",
			slog.Int("index", i+1),
			slog.Int("total", len(schoolLinks)),
			slog.String("url", link),
		)

		// Try to load from cache first
		if cachedDetails, found := s.loadFromCache(link); found {
			allDetails = append(allDetails, *cachedDetails)
			cachedCount++
			continue
		}

		// Not in cache, scrape it
		details, err := s.ScrapeSchoolDetail(ctx, link)
		if err != nil {
			s.logger.Error("failed to scrape school",
				slog.String("url", link),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Save to cache
		if err := s.saveToCache(link, details); err != nil {
			s.logger.Warn("failed to save to cache",
				slog.String("url", link),
				slog.String("error", err.Error()),
			)
		}

		allDetails = append(allDetails, *details)
		scrapedCount++

		// Be respectful to the server (only when scraping, not when using cache)
		time.Sleep(2 * time.Second)
	}

	s.logger.Info("scraping complete",
		slog.Int("total", len(allDetails)),
		slog.Int("from_cache", cachedCount),
		slog.Int("newly_scraped", scrapedCount),
	)
	return allDetails, nil
}

// getSchoolLinks gets all school detail page URLs from the main list
func (s *SchoolDetailsScraper) getSchoolLinks(ctx context.Context) ([]string, error) {
	var links []string

	err := chromedp.Run(ctx,
		chromedp.Navigate(berlinSchoolListURL),
		chromedp.WaitVisible(`#DataListSchulen`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for page to fully load
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('#DataListSchulen > tbody > tr a')).map(a => a.href)
		`, &links),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to extract school links: %w", err)
	}

	return links, nil
}

// ScrapeSchoolDetail scrapes detailed information for a single school
func (s *SchoolDetailsScraper) ScrapeSchoolDetail(ctx context.Context, schoolURL string) (*models.SchoolDetailData, error) {
	// Create new context for this school
	allocCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Add timeout for this school
	timeoutCtx, timeoutCancel := context.WithTimeout(allocCtx, 2*time.Minute)
	defer timeoutCancel()

	details := &models.SchoolDetailData{
		SchoolURL: schoolURL,
		ScrapedAt: time.Now(),
	}

	// Variable to hold the full school name with number
	var schoolNameWithNumber string

	// Navigate to school page and extract basic info
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(schoolURL),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),

		// Extract school name with number (e.g., "Georg-Friedrich-Händel-Gymnasium - 02Y04")
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblSchulname');
				return el ? el.textContent.trim() : '';
			})()
		`, &schoolNameWithNumber),

		// Extract languages
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblSprachen');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.Languages),

		// Extract courses
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblLeistungskurse');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.Courses),

		// Extract offerings
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblAngebote');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.Offerings),

		// Extract additional info
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblBemerkungenSchulzweig');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.AdditionalInfo),

		// Extract equipment
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblAusstattung');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.Equipment),

		// Extract working groups (AGs)
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblAGs');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.WorkingGroups),

		// Extract partners
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblPartner');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.Partners),

		// Extract differentiation
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblDiff');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.Differentiation),

		// Extract lunch info
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblMittag');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.LunchInfo),

		// Extract dual learning
		chromedp.Evaluate(`
			(function() {
				var el = document.getElementById('ContentPlaceHolderMenuListe_lblDualesLernen');
				return el ? el.textContent.trim() : '';
			})()
		`, &details.DualLearning),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to extract basic info: %w", err)
	}

	// Parse school name and number from the format "School Name - SchoolNumber"
	details.SchoolName, details.SchoolNumber = s.parseSchoolNameAndNumber(schoolNameWithNumber)

	// Check if available after 4th grade
	details.AvailableAfter4thGrade = strings.Contains(details.Offerings, "ab Jahrgangsstufe 5 beginnende") ||
		strings.Contains(details.AdditionalInfo, "ab Jahrgangsstufe 5 beginnende")

	// Try to scrape statistics (might not always be available)
	s.scrapeStatistics(timeoutCtx, details)

	return details, nil
}

// scrapeStatistics attempts to scrape student statistics
func (s *SchoolDetailsScraper) scrapeStatistics(ctx context.Context, details *models.SchoolDetailData) {
	s.logger.Info("attempting to scrape statistics", slog.String("school", details.SchoolName))

	// First, check if there's a Schülerschaft tab/link
	var navInfo struct {
		HasNaviSchuelerschaft bool     `json:"hasNaviSchuelerschaft"`
		AvailableTabs         []string `json:"availableTabs"`
		AllIDs                []string `json:"allIds"`
	}

	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				// Look for NaviSchuelerschaft
				var navi = document.getElementById('NaviSchuelerschaft');
				var hasNavi = navi !== null;
				
				// Find all elements with title attribute (potential tabs)
				var tabElements = document.querySelectorAll('[title]');
				var tabs = [];
				for (var i = 0; i < tabElements.length; i++) {
					tabs.push(tabElements[i].getAttribute('title'));
				}
				
				// Get all IDs on the page
				var allElements = document.querySelectorAll('[id]');
				var ids = [];
				for (var i = 0; i < allElements.length; i++) {
					var id = allElements[i].id;
					if (id.includes('Navi') || id.includes('Schuel') || id.includes('Tab')) {
						ids.push(id);
					}
				}
				
				return {
					hasNaviSchuelerschaft: hasNavi,
					availableTabs: tabs,
					allIds: ids
				};
			})()
		`, &navInfo),
	)

	if err != nil {
		s.logger.Warn("error checking for statistics section",
			slog.String("school", details.SchoolName),
			slog.String("error", err.Error()),
		)
		return
	}

	s.logger.Info("statistics navigation check",
		slog.Bool("has_navi", navInfo.HasNaviSchuelerschaft),
		slog.Int("tabs_found", len(navInfo.AvailableTabs)),
		slog.Int("relevant_ids", len(navInfo.AllIDs)),
	)

	// If NaviSchuelerschaft exists, try to click it
	if navInfo.HasNaviSchuelerschaft {
		err = chromedp.Run(ctx,
			chromedp.Click(`#NaviSchuelerschaft`, chromedp.ByQuery),
			chromedp.Sleep(2*time.Second), // Wait longer for ASP.NET postback
		)

		if err != nil {
			s.logger.Warn("failed to click NaviSchuelerschaft",
				slog.String("school", details.SchoolName),
				slog.String("error", err.Error()),
			)
		} else {
			s.logger.Info("clicked NaviSchuelerschaft successfully")
		}
	}

	// Try to scrape each statistic type
	s.logger.Info("attempting to scrape citizenship data")
	details.CitizenshipTable = s.scrapeStatisticTable(ctx, "Staatsangehörigkeit")

	s.logger.Info("attempting to scrape language data")
	details.LanguageTable = s.scrapeStatisticTable(ctx, "Nichtdeutsche Herkunftssprache")

	s.logger.Info("attempting to scrape residence data")
	details.ResidenceTable = s.scrapeStatisticTable(ctx, "Wohnorte")

	s.logger.Info("attempting to scrape absence data")
	details.AbsenceTable = s.scrapeStatisticTable(ctx, "Fehlzeiten")

	// Log summary
	var foundCount int
	if details.CitizenshipTable != nil {
		foundCount++
	}
	if details.LanguageTable != nil {
		foundCount++
	}
	if details.ResidenceTable != nil {
		foundCount++
	}
	if details.AbsenceTable != nil {
		foundCount++
	}

	s.logger.Info("statistics scraping complete",
		slog.Int("tables_found", foundCount),
		slog.Int("tables_expected", 4),
	)
}

// scrapeStatisticTable clicks on a tab and scrapes the table data
func (s *SchoolDetailsScraper) scrapeStatisticTable(ctx context.Context, tabTitle string) *models.StatisticTable {
	var tableHTML string
	var clickSuccess bool

	// First, try to find and click the tab
	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				console.log('Looking for tab with title: %s');
				var elements = document.querySelectorAll('[title="%s"]');
				console.log('Found elements:', elements.length);
				if (elements.length > 0) {
					elements[0].click();
					console.log('Clicked on tab');
					return true;
				}
				return false;
			})()
		`, tabTitle, tabTitle), &clickSuccess),
	)

	if err != nil {
		s.logger.Warn("error finding tab element",
			slog.String("title", tabTitle),
			slog.String("error", err.Error()),
		)
		return nil
	}

	if !clickSuccess {
		s.logger.Info("tab element not found",
			slog.String("title", tabTitle),
		)
		return nil
	}

	// Wait for the table to appear after clicking
	err = chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second), // Give more time for ASP.NET to load content
	)

	if err != nil {
		s.logger.Warn("error waiting after tab click",
			slog.String("title", tabTitle),
			slog.String("error", err.Error()),
		)
		return nil
	}

	// Extract table HTML with multiple strategies
	var tableInfo struct {
		HTML          string   `json:"html"`
		TotalTables   int      `json:"totalTables"`
		VisibleTables int      `json:"visibleTables"`
		TableInfo     []string `json:"tableInfo"`
	}

	err = chromedp.Run(ctx,
		chromedp.Evaluate(`
			(function() {
				console.log('Looking for visible tables');
				var tables = document.querySelectorAll('table');
				console.log('Total tables found:', tables.length);
				
				var result = {
					html: '',
					totalTables: tables.length,
					visibleTables: 0,
					tableInfo: []
				};
				
				// Collect info about all tables
				for (var i = 0; i < tables.length; i++) {
					var table = tables[i];
					var isVisible = table.offsetParent !== null;
					var display = window.getComputedStyle(table).display;
					var visibility = window.getComputedStyle(table).visibility;
					var hasContent = table.rows && table.rows.length > 0;
					var rowCount = table.rows ? table.rows.length : 0;
					var id = table.id || 'no-id';
					var classes = table.className || 'no-class';
					
					var info = 'Table ' + i + ': id=' + id + ', class=' + classes + 
					           ', visible=' + isVisible + ', display=' + display + 
					           ', visibility=' + visibility + ', rows=' + rowCount;
					result.tableInfo.push(info);
					console.log(info);
					
					// Strategy 1: offsetParent check
					if (isVisible && hasContent && display !== 'none' && visibility !== 'hidden') {
						result.visibleTables++;
						if (result.html === '') {
							console.log('Found visible table with offsetParent:', i);
							result.html = table.outerHTML;
						}
					}
				}
				
				// Strategy 2: If no table found with offsetParent, try finding by display style
				if (result.html === '') {
					for (var i = 0; i < tables.length; i++) {
						var table = tables[i];
						var display = window.getComputedStyle(table).display;
						var visibility = window.getComputedStyle(table).visibility;
						var hasContent = table.rows && table.rows.length > 0;
						
						if (display !== 'none' && visibility !== 'hidden' && hasContent) {
							console.log('Found table with display check:', i);
							result.html = table.outerHTML;
							break;
						}
					}
				}
				
				// Strategy 3: If still no table, get the first table with content
				if (result.html === '' && tables.length > 0) {
					for (var i = 0; i < tables.length; i++) {
						var table = tables[i];
						if (table.rows && table.rows.length > 0) {
							console.log('Using first table with content:', i);
							result.html = table.outerHTML;
							break;
						}
					}
				}
				
				console.log('Result: totalTables=' + result.totalTables + 
				           ', visibleTables=' + result.visibleTables + 
				           ', htmlLength=' + result.html.length);
				return result;
			})()
		`, &tableInfo),
	)

	tableHTML = tableInfo.HTML

	if err != nil {
		s.logger.Warn("error extracting table HTML",
			slog.String("title", tabTitle),
			slog.String("error", err.Error()),
		)
		return nil
	}

	// Log detailed table information
	s.logger.Info("table detection results",
		slog.String("title", tabTitle),
		slog.Int("total_tables", tableInfo.TotalTables),
		slog.Int("visible_tables", tableInfo.VisibleTables),
		slog.Int("html_length", len(tableHTML)),
	)

	// Log each table's information
	for i, info := range tableInfo.TableInfo {
		if i < 5 { // Limit to first 5 tables to avoid log spam
			s.logger.Info("table details", slog.String("info", info))
		}
	}

	if tableHTML == "" {
		s.logger.Warn("no table HTML found for statistic",
			slog.String("title", tabTitle),
			slog.Int("total_tables_on_page", tableInfo.TotalTables),
		)
		return nil
	}

	s.logger.Info("successfully extracted table",
		slog.String("title", tabTitle),
		slog.Int("html_length", len(tableHTML)),
	)

	// Parse the table HTML
	table := s.parseTableHTML(tableHTML)
	return table
}

// parseTableHTML parses HTML table into structured data using Go's html parser
func (s *SchoolDetailsScraper) parseTableHTML(tableHTML string) *models.StatisticTable {
	if tableHTML == "" {
		return nil
	}

	// Parse HTML using Go's standard library
	doc, err := html.Parse(strings.NewReader(tableHTML))
	if err != nil {
		s.logger.Warn("failed to parse HTML", slog.String("error", err.Error()))
		return nil
	}

	table := &models.StatisticTable{
		Headers: []string{},
		Rows:    [][]string{},
	}

	// Find the table element
	var findTable func(*html.Node) *html.Node
	findTable = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && n.Data == "table" {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := findTable(c); result != nil {
				return result
			}
		}
		return nil
	}

	tableNode := findTable(doc)
	if tableNode == nil {
		s.logger.Warn("no table element found in HTML")
		return nil
	}

	// Extract headers and rows
	var extractText func(*html.Node) string
	extractText = func(n *html.Node) string {
		if n.Type == html.TextNode {
			return strings.TrimSpace(n.Data)
		}
		var text string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			text += extractText(c)
		}
		return strings.TrimSpace(text)
	}

	// Parse all rows
	var parseNode func(*html.Node, bool)
	var isFirstRow = true
	parseNode = func(n *html.Node, inThead bool) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "thead":
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					parseNode(c, true)
				}
				return
			case "tbody":
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					parseNode(c, false)
				}
				return
			case "tr":
				var cells []string
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && (c.Data == "th" || c.Data == "td") {
						cellText := extractText(c)
						cells = append(cells, cellText)
					}
				}

				if len(cells) > 0 {
					// First row in thead or first row overall becomes headers
					if inThead || isFirstRow {
						table.Headers = cells
						isFirstRow = false
					} else {
						table.Rows = append(table.Rows, cells)
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if n.Type == html.ElementNode && n.Data != "thead" && n.Data != "tbody" {
				parseNode(c, inThead)
			}
		}
	}

	parseNode(tableNode, false)

	s.logger.Info("parsed table",
		slog.Int("headers", len(table.Headers)),
		slog.Int("rows", len(table.Rows)),
	)

	return table
}

// parseSchoolNameAndNumber parses the school name and number from the format
// "School Name - SchoolNumber" (e.g., "Georg-Friedrich-Händel-Gymnasium - 02Y04")
// Returns the school name and school number separately
func (s *SchoolDetailsScraper) parseSchoolNameAndNumber(fullName string) (string, string) {
	if fullName == "" {
		return "", ""
	}

	// Split by " - " to separate name and number
	parts := strings.Split(fullName, " - ")
	if len(parts) >= 2 {
		// The school name is everything except the last part
		schoolName := strings.Join(parts[:len(parts)-1], " - ")
		// The school number is the last part
		schoolNumber := strings.TrimSpace(parts[len(parts)-1])
		return strings.TrimSpace(schoolName), schoolNumber
	}

	// If no " - " separator found, return the full name as school name
	return strings.TrimSpace(fullName), ""
}
