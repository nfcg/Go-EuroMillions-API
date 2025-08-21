package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Result struct represents a single EuroMillions drawing result.
// It includes JSON and XML tags for serialization.
type Result struct {
	Date    string `json:"date" xml:"date"`
	Numbers []int  `json:"numbers" xml:"numbers>number"`
	Stars   []int  `json:"stars" xml:"stars>star"` // This line has been corrected
}

// AllResults is a helper struct for XML output with a root element.
type AllResults struct {
	XMLName xml.Name `xml:"results"`
	Results []Result `xml:"result"`
}

var (
	db          *sql.DB
	dbPath      string
	showHelp    bool
	versionFlag bool
	verbose     bool
	logFilePath string
)

const (
	version = "1.2"
)

// init is called before main. It sets up command-line flags with both long and short versions.
func init() {
	// Long and short flags for database path
	flag.StringVar(&dbPath, "db", "./euromillions.db", "Path to the SQLite database file")
	flag.StringVar(&dbPath, "d", "./euromillions.db", "Path to the SQLite database file (shorthand)")

	// Long and short flags for showing help
	flag.BoolVar(&showHelp, "help", false, "Show the application help message")
	flag.BoolVar(&showHelp, "h", false, "Show the application help message (shorthand)")

	// Long and short flags for showing version
	flag.BoolVar(&versionFlag, "version", false, "Show the application version")
	flag.BoolVar(&versionFlag, "v", false, "Show the application version (shorthand)")

	// New: Long and short flags for verbose logging
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging for requests")
	// The -v flag is already used for version, so we must choose a different short flag for verbose.
	// We will omit the short flag for verbose to avoid conflicts.
	
	// New: Long and short flags for log file path
	flag.StringVar(&logFilePath, "log-file", "", "Path to a file to write logs to")
	flag.StringVar(&logFilePath, "l", "", "Path to a file to write logs to (shorthand)")
}

// main is the entry point of the application.
func main() {
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	if versionFlag {
		fmt.Printf("EuroMillions API v%s\n", version)
		return
	}
	
	// New: Configure log output based on the provided flag
	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	// Initialize the database connection and apply optimizations.
	if err := initDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Configure HTTP handlers for different endpoints.
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/results", resultsHandler)
	http.HandleFunc("/results/latest", latestHandler)
	http.HandleFunc("/results/date/", dateHandler)
	http.HandleFunc("/results/year/", yearHandler)
	http.HandleFunc("/results/month/", monthYearHandler)

	log.Printf("Server started on port 8080 (Database: %s)", dbPath)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// printHelp displays a detailed help message, including usage, flags, and available endpoints.
func printHelp() {
	fmt.Println("EuroMillions API - Results Server")
	fmt.Println("---------------------------------")
	fmt.Println("\nUsage:")
	fmt.Println("  ./euromillions-api [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nAvailable Endpoints:")
	fmt.Println("  GET /                        - Returns the latest drawing result (default).")
	fmt.Println("  GET /results                 - Returns all drawing results.")
	fmt.Println("  GET /results/latest          - Returns the latest drawing result.")
	fmt.Println("  GET /results/date/{date}     - Search by a specific date (e.g., /results/date/2024-01-15).")
	fmt.Println("  GET /results/year/{year}     - Search by year (e.g., /results/year/2023).")
	fmt.Println("  GET /results/month/{month}   - Search by month and year (e.g., /results/month/2024-03).")
	fmt.Println("\nURL Query Parameters for Output Format:")
	fmt.Println("  ?format=json                 - Returns the response in JSON format (default).")
	fmt.Println("  ?format=xml                  - Returns the response in XML format.")
	fmt.Println("  ?format=plaintext            - Returns the response in plain text format.")
}

// defaultHandler redirects the root path to the latest result handler.
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if verbose {
		log.Printf("GET request for / from %s", r.RemoteAddr)
	}
	latestHandler(w, r)
}

// setPragmas applies SQLite PRAGMA settings for optimal performance.
func setPragmas() error {
	// PRAGMA journal_mode: Use WAL for better concurrency and speed.
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		return fmt.Errorf("error setting PRAGMA journal_mode: %v", err)
	}

	// PRAGMA synchronous: Set to NORMAL for a good balance of speed and safety.
	if _, err := db.Exec("PRAGMA synchronous = NORMAL;"); err != nil {
		return fmt.Errorf("error setting PRAGMA synchronous: %v", err)
	}
	return nil
}

// initDB initializes the database connection and performs basic validation.
func initDB() error {
	// Check if the database file exists.
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database file not found at: %s", dbPath)
	}

	// Get the absolute path for consistency.
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Errorf("error getting absolute database path: %v", err)
	}
	dbPath = absPath

	// Open the SQLite database connection.
	var errOpen error
	db, errOpen = sql.Open("sqlite3", dbPath)
	if errOpen != nil {
		return fmt.Errorf("error opening database: %v", errOpen)
	}

	// Apply PRAGMA settings for performance.
	if err := setPragmas(); err != nil {
		db.Close()
		return err
	}

	// Verify that the 'results' table exists.
	tableExists := false
	err = db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name='results'").Scan(&tableExists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking table: %v", err)
	}

	if !tableExists {
		return fmt.Errorf("table 'results' not found in database")
	}

	// Verify the table schema by running a simple query.
	_, err = db.Exec("SELECT date, number_1, number_2, number_3, number_4, number_5, star_1, star_2 FROM results LIMIT 1")
	if err != nil {
		return fmt.Errorf("table schema does not match the expected format: %v", err)
	}

	return nil
}

// resultsHandler serves all available results.
func resultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if verbose {
		log.Printf("GET request for /results from %s", r.RemoteAddr)
	}
	getAllResults(w, r)
}

// getAllResults queries the database for all results and returns them in the requested format.
func getAllResults(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT date, number_1, number_2, number_3, number_4, number_5, star_1, star_2 FROM results ORDER BY date DESC")
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		log.Printf("Error fetching results: %v", err)
		return
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var res Result
		var n1, n2, n3, n4, n5, s1, s2 int
		err := rows.Scan(&res.Date, &n1, &n2, &n3, &n4, &n5, &s1, &s2)
		if err != nil {
			http.Error(w, "Error processing results", http.StatusInternalServerError)
			log.Printf("Error reading database row: %v", err)
			return
		}
		res.Numbers = []int{n1, n2, n3, n4, n5}
		res.Stars = []int{s1, s2}
		results = append(results, res)
	}

	if len(results) == 0 {
		http.Error(w, "No results found", http.StatusNotFound)
		return
	}

	sendResponse(w, r, results)
}

// latestHandler serves the latest result.
func latestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if verbose {
		log.Printf("GET request for /results/latest from %s", r.RemoteAddr)
	}

	var result Result
	var n1, n2, n3, n4, n5, s1, s2 int
	err := db.QueryRow("SELECT date, number_1, number_2, number_3, number_4, number_5, star_1, star_2 FROM results ORDER BY date DESC LIMIT 1").
		Scan(&result.Date, &n1, &n2, &n3, &n4, &n5, &s1, &s2)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No results found", http.StatusNotFound)
		} else {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			log.Printf("Error fetching latest result: %v", err)
		}
		return
	}

	result.Numbers = []int{n1, n2, n3, n4, n5}
	result.Stars = []int{s1, s2}

	sendResponse(w, r, []Result{result})
}

// dateHandler serves the result for a specific date.
func dateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if verbose {
		log.Printf("GET request for /results/date/ from %s", r.RemoteAddr)
	}

	date := r.URL.Path[len("/results/date/"):]
	if date == "" {
		http.Error(w, "Date parameter is required (format YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006-01-02", date); err != nil {
		http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	var result Result
	var n1, n2, n3, n4, n5, s1, s2 int
	err := db.QueryRow("SELECT date, number_1, number_2, number_3, number_4, number_5, star_1, star_2 FROM results WHERE date = ?", date).
		Scan(&result.Date, &n1, &n2, &n3, &n4, &n5, &s1, &s2)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No results found for the specified date", http.StatusNotFound)
		} else {
			http.Error(w, "Error querying database", http.StatusInternalServerError)
			log.Printf("Error fetching result by date (%s): %v", date, err)
		}
		return
	}

	result.Numbers = []int{n1, n2, n3, n4, n5}
	result.Stars = []int{s1, s2}

	sendResponse(w, r, []Result{result})
}

// yearHandler serves all results for a specific year.
func yearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if verbose {
		log.Printf("GET request for /results/year/ from %s", r.RemoteAddr)
	}

	year := r.URL.Path[len("/results/year/"):]
	if year == "" {
		http.Error(w, "Year parameter is required (format YYYY)", http.StatusBadRequest)
		return
	}

	if _, err := time.Parse("2006", year); err != nil {
		http.Error(w, "Invalid year format (use YYYY)", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT date, number_1, number_2, number_3, number_4, number_5, star_1, star_2 FROM results WHERE strftime('%Y', date) = ? ORDER BY date DESC", year)
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		log.Printf("Error fetching results by year (%s): %v", year, err)
		return
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var res Result
		var n1, n2, n3, n4, n5, s1, s2 int
		err := rows.Scan(&res.Date, &n1, &n2, &n3, &n4, &n5, &s1, &s2)
		if err != nil {
			http.Error(w, "Error processing results", http.StatusInternalServerError)
			log.Printf("Error reading database row: %v", err)
			return
		}
		res.Numbers = []int{n1, n2, n3, n4, n5}
		res.Stars = []int{s1, s2}
		results = append(results, res)
	}

	if len(results) == 0 {
		http.Error(w, fmt.Sprintf("No results found for the year %s", year), http.StatusNotFound)
		return
	}

	sendResponse(w, r, results)
}

// monthYearHandler serves all results for a specific month and year.
func monthYearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if verbose {
		log.Printf("GET request for /results/month/ from %s", r.RemoteAddr)
	}

	monthYear := r.URL.Path[len("/results/month/"):]
	if monthYear == "" {
		http.Error(w, "Month/Year parameter is required (format YYYY-MM)", http.StatusBadRequest)
		return
	}

	parts := strings.Split(monthYear, "-")
	if len(parts) != 2 {
		http.Error(w, "Invalid format (use YYYY-MM)", http.StatusBadRequest)
		return
	}

	year := parts[0]
	month := parts[1]

	if _, err := time.Parse("2006-01", monthYear); err != nil {
		http.Error(w, "Invalid month/year format (use YYYY-MM)", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT date, number_1, number_2, number_3, number_4, number_5, star_1, star_2 FROM results WHERE strftime('%Y', date) = ? AND strftime('%m', date) = ? ORDER BY date DESC", year, month)
	if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		log.Printf("Error fetching results by month/year (%s): %v", monthYear, err)
		return
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var res Result
		var n1, n2, n3, n4, n5, s1, s2 int
		err := rows.Scan(&res.Date, &n1, &n2, &n3, &n4, &n5, &s1, &s2)
		if err != nil {
			http.Error(w, "Error processing results", http.StatusInternalServerError)
			log.Printf("Error reading database row: %v", err)
			return
		}
		res.Numbers = []int{n1, n2, n3, n4, n5}
		res.Stars = []int{s1, s2}
		results = append(results, res)
	}

	if len(results) == 0 {
		http.Error(w, fmt.Sprintf("No results found for %s", monthYear), http.StatusNotFound)
		return
	}

	sendResponse(w, r, results)
}

// sendResponse writes the response in the correct format (XML, Plain Text, or JSON).
// It prioritizes the 'format' URL query parameter.
func sendResponse(w http.ResponseWriter, r *http.Request, results []Result) {
	format := r.URL.Query().Get("format")

	switch strings.ToLower(format) {
	case "xml":
		w.Header().Set("Content-Type", "application/xml")
		if len(results) == 1 {
			if err := xml.NewEncoder(w).Encode(results[0]); err != nil {
				log.Printf("Error encoding XML response: %v", err)
			}
		} else {
			allResults := AllResults{Results: results}
			if err := xml.NewEncoder(w).Encode(allResults); err != nil {
				log.Printf("Error encoding XML response: %v", err)
			}
		}
		return
	case "plaintext":
		w.Header().Set("Content-Type", "text/plain")
		for _, result := range results {
			numbers := fmt.Sprintf("%d,%d,%d,%d,%d", result.Numbers[0], result.Numbers[1], result.Numbers[2], result.Numbers[3], result.Numbers[4])
			stars := fmt.Sprintf("%d,%d", result.Stars[0], result.Stars[1])
			fmt.Fprintf(w, "Date: %s, Numbers: %s, Stars: %s\n", result.Date, numbers, stars)
		}
		return
	default: // Fallback to JSON
		w.Header().Set("Content-Type", "application/json")
		if len(results) == 1 {
			if err := json.NewEncoder(w).Encode(results[0]); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}
		} else {
			if err := json.NewEncoder(w).Encode(results); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
			}
		}
		return
	}
}
