package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// List of common User-Agents to use randomly
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.4; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.2420.81",
}

var (
	verboseFlag  bool
	outputFile   string
	databasePath string
	siteIDStr    string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&databasePath, "database", "", "Path to the SQLite database file.")
	flag.StringVar(&databasePath, "d", "", "Path to the SQLite database file. (shorthand)")
	flag.StringVar(&siteIDStr, "site", "", "The site ID to update (1, 2, 3, 4, 5) or 'all' to run all.")
	flag.StringVar(&siteIDStr, "s", "", "The site ID to update (1, 2, 3, 4, 5) or 'all' to run all. (shorthand)")
	flag.BoolVar(&verboseFlag, "verbose", false, "Enable verbose logging.")
	flag.BoolVar(&verboseFlag, "v", false, "Enable verbose logging. (shorthand)")
	flag.StringVar(&outputFile, "output", "", "Path to a log file. Output is to console by default.")
	flag.StringVar(&outputFile, "o", "", "Path to a log file. Output is to console by default. (shorthand)")
}

func getBetween(s, start, end string) string {
	initialPos := strings.Index(s, start)
	if initialPos == -1 {
		return ""
	}
	initialPos += len(start)
	endPos := strings.Index(s[initialPos:], end)
	if endPos == -1 {
		return ""
	}
	return s[initialPos : initialPos+endPos]
}

func getWebPage(url string) (string, error) {
	if verboseFlag {
		log.Printf("Fetching URL: %s", url)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	randomUserAgent := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", randomUserAgent)
	req.Header.Set("Referer", "https://www.bing.com/?cc=pt")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getCSV(url string) (string, error) {
	if verboseFlag {
		log.Printf("Fetching CSV from URL: %s", url)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	randomUserAgent := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", randomUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func runUpdate(db *sql.DB, siteID int) error {
	var (
		url     string
		newDate string
		numbers []string
		err     error
	)

	log.Printf("Executing option for Site ID: %d", siteID)
	
	var oldDate string
	err = db.QueryRow("SELECT date FROM results ORDER BY date DESC LIMIT 1").Scan(&oldDate)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("database query error: %v", err)
	}

	if verboseFlag {
		log.Printf("Last date in database for this run: %s", oldDate)
	}
	
	switch siteID {
	case 1:
		url = "https://www.euromilhoes.com/"
		var response string
		response, err = getWebPage(url)
		if err != nil {
			return fmt.Errorf("failed to fetch page: %v", err)
		}
		full := getBetween(response, "last-results-container", "selector-wrapper")
		dataStr := getBetween(full, "<span>", "</span>")
		var t time.Time
		t, err = time.Parse("02.01.2006", dataStr)
		if err != nil {
			return fmt.Errorf("date parsing error: %v", err)
		}
		newDate = t.Format("2006-01-02")
		numFull := getBetween(full, `<ul class="results">`, `</ul>`)
		re := regexp.MustCompile(`>(\d+)<`)
		matches := re.FindAllStringSubmatch(numFull, -1)
		for _, match := range matches {
			numbers = append(numbers, match[1])
		}
	case 2:
		url = "https://www.euro-millions.com/results"
		var response string
		response, err = getWebPage(url)
		if err != nil {
			return fmt.Errorf("failed to fetch page: %v", err)
		}
		full := getBetween(response, `<ul class="balls">`, `</ul>`)
		dataStr := getBetween(response, `<li><a href="/results/`, `"`)
		var t time.Time
		t, err = time.Parse("02-01-2006", dataStr)
		if err != nil {
			return fmt.Errorf("date parsing error: %v", err)
		}
		newDate = t.Format("2006-01-02")
		re := regexp.MustCompile(`>(\d+)<`)
		matches := re.FindAllStringSubmatch(full, -1)
		for _, match := range matches {
			numbers = append(numbers, match[1])
		}
	case 3:
		url = "https://www.jogossantacasa.pt/web/SCCartazResult/"
		response, err := getWebPage(url)
		if err != nil {
			return fmt.Errorf("failed to fetch page: %v", err)
		}

		dateRegex := regexp.MustCompile(`Data do Sorteio - (\d{2}\/\d{2}\/\d{4})`)
		dateMatches := dateRegex.FindStringSubmatch(response)
		if len(dateMatches) < 2 {
			return fmt.Errorf("could not find the date in the page content")
		}
		dataStr := dateMatches[1]
		
		var t time.Time
		t, err = time.Parse("02/01/2006", dataStr)
		if err != nil {
			return fmt.Errorf("error parsing date from website: %v", err)
		}
		newDate = t.Format("2006-01-02")

		numRegex := regexp.MustCompile(`<li>(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})\s+(\d{1,2})\s+\+\s+(\d{1,2})\s+(\d{1,2})`)
		numMatches := numRegex.FindAllStringSubmatch(response, -1)

		if len(numMatches) < 1 || len(numMatches[0]) != 8 {
			return fmt.Errorf("expected 7 numbers, found %d", len(numMatches))
		}

		for i := 1; i <= 7; i++ {
			numbers = append(numbers, numMatches[0][i])
		}
		
	case 4:
		url = "https://www.euromilhoes.com/"
		response, err := getWebPage(url)
		if err != nil {
			return fmt.Errorf("failed to fetch page: %v", err)
		}

		dateSection := getBetween(response, `<section class="last-results">`, `</section>`)
		if verboseFlag {
			log.Printf("Raw HTML snippet for date search: %s", dateSection)
		}
		dateRegex := regexp.MustCompile(`<span>(\d{2}\.\d{2}\.\d{4})</span>`)
		dateMatches := dateRegex.FindStringSubmatch(dateSection)
		
		if len(dateMatches) < 2 {
			return fmt.Errorf("could not find the date in the page content")
		}
		dataStr := dateMatches[1]
		var t time.Time
		t, err = time.Parse("02.01.2006", dataStr)
		if err != nil {
			return fmt.Errorf("date parsing error: %v", err)
		}
		newDate = t.Format("2006-01-02")

		numSection := getBetween(response, `<ul class="results">`, `</ul>`)
		if numSection == "" {
			return fmt.Errorf("could not find the numbers section")
		}

		if verboseFlag {
			log.Printf("Raw HTML snippet for numbers search: %s", numSection)
		}

		numRegex := regexp.MustCompile(`>(\d+)<`)
		matches := numRegex.FindAllStringSubmatch(numSection, -1)
		
		if verboseFlag {
			log.Printf("Numbers found by regex: %v", matches)
		}

		if len(matches) < 7 {
			return fmt.Errorf("invalid number of results for insertion. Expected 7, got: %d", len(matches))
		}
		for _, match := range matches {
			numbers = append(numbers, match[1])
		}

	case 5:
		url = "https://www.national-lottery.co.uk/results/euromillions/draw-history/csv"
		csvData, err := getCSV(url)
		if err != nil {
			return fmt.Errorf("failed to fetch CSV: %v", err)
		}

		r := csv.NewReader(strings.NewReader(csvData))
		
		_, err = r.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV header: %v", err)
		}

		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("no data found in CSV")
			}
			return fmt.Errorf("failed to read CSV record: %v", err)
		}

		if len(record) < 8 {
			return fmt.Errorf("invalid CSV format. Expected at least 8 columns, got %d", len(record))
		}

		var t time.Time
		t, err = time.Parse("02-Jan-2006", record[0])
		if err != nil {
			return fmt.Errorf("date parsing error: %v", err)
		}
		newDate = t.Format("2006-01-02")

		numbers = []string{
			record[1], // Ball 1
			record[2], // Ball 2
			record[3], // Ball 3
			record[4], // Ball 4
			record[5], // Ball 5
			record[6], // Lucky Star 1
			record[7], // Lucky Star 2
		}

		for i, num := range numbers {
			if _, err := strconv.Atoi(num); err != nil {
				return fmt.Errorf("invalid number at position %d: %s", i+1, num)
			}
		}

	default:
		return fmt.Errorf("unsupported site ID: %d", siteID)
	}

	if newDate == oldDate {
		log.Printf("Exiting. The date is the same: %s", newDate)
		return nil
	}
	if newDate > oldDate {
		log.Printf("OK. New date: %s", newDate)
		log.Printf("Numbers: %s", strings.Join(numbers, ", "))

		if len(numbers) != 7 {
			return fmt.Errorf("invalid number of results for insertion. Expected 7, got: %d", len(numbers))
		}

		stmt, err := db.Prepare("INSERT INTO results (date, number_1, number_2, number_3, number_4, number_5, star_1, star_2) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("failed to prepare SQL statement: %v", err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(newDate, numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5], numbers[6])
		if err != nil {
			return fmt.Errorf("failed to execute SQL statement: %v", err)
		}
		log.Println("Data inserted successfully.")
	} else {
		log.Println("Exiting. The old date is more recent than the new one.")
	}
	
	return nil
}

func main() {
	flag.Parse()

	if databasePath == "" || siteIDStr == "" {
		flag.Usage()
		os.Exit(1)
	}

	if outputFile != "" {
		logFile, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	if siteIDStr == "all" {
		sitesToUpdate := []int{1, 2, 3, 4, 5}
		for _, id := range sitesToUpdate {
			if err := runUpdate(db, id); err != nil {
				log.Printf("Error processing site %d: %v", id, err)
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		siteID, err := strconv.Atoi(siteIDStr)
		if err != nil {
			log.Fatalf("Invalid site ID: %v", err)
		}
		if err := runUpdate(db, siteID); err != nil {
			log.Fatal(err)
		}
	}
}
