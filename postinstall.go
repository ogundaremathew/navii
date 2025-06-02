// Package yuniq provides geographical navigation and state management functionality
package navii

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PostalCodeFormat represents postal code validation patterns
type PostalCodeFormat struct {
	CountryCode string
	Pattern     *regexp.Regexp
}

// PostalCode represents a postal code entry
type PostalCode struct {
	CountryCode string `json:"countryCode"`
	PostalCode  string `json:"postalCode"`
}

// CountryData represents country information from the API
type CountryData struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ISO3         string `json:"iso3"`
	ISO2         string `json:"iso2"`
	NumericCode  string `json:"numeric_code"`
	PhoneCode    string `json:"phonecode"`
	Capital      string `json:"capital"`
	Currency     string `json:"currency"`
	CurrencyName string `json:"currency_name"`
	TLD          string `json:"tld"`
	Native       string `json:"native,omitempty"`
	Region       string `json:"region"`
	Subregion    string `json:"subregion"`
	Nationality  string `json:"nationality"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Emoji        string `json:"emoji"`
	EmojiU       string `json:"emojiU"`
}

// CityDataFromAPI represents city information from the API
type CityDataFromAPI struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	StateID     int    `json:"state_id"`
	StateCode   string `json:"state_code"`
	StateName   string `json:"state_name"`
	CountryID   int    `json:"country_id"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	WikiDataID  string `json:"wikiDataId"`
}

// DataDownloader handles downloading and processing geographical data
type DataDownloader struct {
	httpClient       *http.Client
	postalCodeRegexs map[string]*regexp.Regexp
	targetCountries  []string
}

// NewDataDownloader creates a new data downloader
func NewDataDownloader() *DataDownloader {
	// Countries that heavily rely on postal codes
	targetCountries := []string{"US", "CA", "GB", "DE", "JP", "FR", "IN", "AU", "NL", "IE"}

	// Postal code format validators
	postalCodeRegexs := map[string]*regexp.Regexp{
		"US": regexp.MustCompile(`^\d{5}$`),                                                            // 5 digits
		"CA": regexp.MustCompile(`^[A-Z]\d[A-Z]\s?\d[A-Z]\d$`),                                         // 6 alphanumeric
		"GB": regexp.MustCompile(`^(?:[A-Z]{1,2}\d{1,2}[A-Z]?|[A-Z]{1,2}\d{1,2}[A-Z]?\s?\d[A-Z]{2})$`), // UK format
		"DE": regexp.MustCompile(`^\d{5}$`),                                                            // 5 digits
		"JP": regexp.MustCompile(`^\d{3}-\d{4}$`),                                                      // 7 digits with hyphen
		"FR": regexp.MustCompile(`^\d{5}$`),                                                            // 5 digits
		"IN": regexp.MustCompile(`^\d{6}$`),                                                            // 6 digits
		"AU": regexp.MustCompile(`^\d{4}$`),                                                            // 4 digits
		"NL": regexp.MustCompile(`^\d{4}[A-Z]{2}$`),                                                    // 4 digits + 2 letters
		"IE": regexp.MustCompile(`^[A-Z0-9]{3}$`),                                                      // 3 alphanumeric
	}

	return &DataDownloader{
		httpClient:       &http.Client{Timeout: 240 * time.Second},
		postalCodeRegexs: postalCodeRegexs,
		targetCountries:  targetCountries,
	}
}

// DownloadAndProcessData downloads and processes all geographical data
func (dd *DataDownloader) DownloadAndProcessData(outputPath string) error {
	fmt.Println("Starting geographical data download...")

	// Download countries and cities
	locationData, err := dd.downloadLocationData()
	if err != nil {
		return fmt.Errorf("failed to download location data: %w", err)
	}

	fmt.Println("Downloading postal codes...")
	postalCodes, err := dd.downloadPostalCodes()
	if err != nil {
		return fmt.Errorf("failed to download postal codes: %w", err)
	}

	// Convert postal codes to zip data format
	zipData := make(map[string][]string)
	for _, pc := range postalCodes {
		zipData[pc.CountryCode] = append(zipData[pc.CountryCode], pc.PostalCode)
	}

	// Create final data structure
	finalData := LocationData{
		CityData: locationData,
		ZipData:  zipData,
	}

	// Write to file
	return dd.writeLocationFile(outputPath, finalData)
}

// downloadLocationData downloads countries and cities data
func (dd *DataDownloader) downloadLocationData() (map[string]map[string][]string, error) {
	baseURL := "https://raw.githubusercontent.com/dr5hn/countries-states-cities-database/refs/heads/master/json"

	// Download countries
	fmt.Println("Downloading countries...")
	countriesData, err := dd.downloadJSON(fmt.Sprintf("%s/countries.json", baseURL))
	if err != nil {
		return nil, err
	}

	var countries []CountryData
	if err := json.Unmarshal(countriesData, &countries); err != nil {
		return nil, err
	}

	// Initialize location data structure
	locationData := make(map[string]map[string][]string)
	for _, country := range countries {
		key := fmt.Sprintf("%s#%s", strings.ToUpper(country.ISO2), country.Name)
		locationData[key] = make(map[string][]string)
	}

	// Download cities
	fmt.Println("Downloading cities...")
	citiesData, err := dd.downloadJSON(fmt.Sprintf("%s/cities.json", baseURL))
	if err != nil {
		return nil, err
	}

	var cities []CityDataFromAPI
	if err := json.Unmarshal(citiesData, &cities); err != nil {
		return nil, err
	}

	// Process cities data
	dd.processCities(cities, locationData)

	fmt.Println("Location data download completed")
	return locationData, nil
}

// processCities processes cities and adds them to location data
func (dd *DataDownloader) processCities(cities []CityDataFromAPI, locationData map[string]map[string][]string) {
	for _, city := range cities {
		countryCode := strings.ToUpper(strings.TrimSpace(city.CountryCode))
		stateCode := strings.ToUpper(city.StateCode)

		if stateCode == "" {
			continue
		}

		// Find country in location data
		var countryKey string
		for key := range locationData {
			if strings.HasPrefix(key, countryCode+"#") {
				countryKey = key
				break
			}
		}

		if countryKey == "" {
			continue
		}

		// Find or create state key
		stateKey := fmt.Sprintf("%s##%s", stateCode, city.StateName)
		var foundStateKey string
		for key := range locationData[countryKey] {
			if strings.HasPrefix(key, stateCode+"##") {
				foundStateKey = key
				break
			}
		}

		if foundStateKey == "" {
			foundStateKey = stateKey
			locationData[countryKey][foundStateKey] = []string{}
		}

		// Add city
		locationData[countryKey][foundStateKey] = append(locationData[countryKey][foundStateKey], city.Name)
	}
}

// downloadPostalCodes downloads postal codes for target countries
func (dd *DataDownloader) downloadPostalCodes() ([]PostalCode, error) {
	var allPostalCodes []PostalCode

	for _, countryCode := range dd.targetCountries {
		fmt.Printf("Downloading postal codes for %s...\n", countryCode)

		postalCodes, err := dd.downloadCountryPostalCodes(countryCode)
		if err != nil {
			return nil, fmt.Errorf("failed to download postal codes for %s: %w", countryCode, err)
		}

		allPostalCodes = append(allPostalCodes, postalCodes...)
		fmt.Printf("Downloaded %d postal codes for %s\n", len(postalCodes), countryCode)
	}

	return allPostalCodes, nil
}

// downloadCountryPostalCodes downloads postal codes for a specific country
func (dd *DataDownloader) downloadCountryPostalCodes(countryCode string) ([]PostalCode, error) {
	isFullFormatCountry := contains([]string{"NL", "CA", "GB"}, countryCode)
	suffix := ""
	targetFileSuffix := ""
	if isFullFormatCountry {
		suffix = "_full.csv"
		targetFileSuffix = "_full"
	}

	url := fmt.Sprintf("https://download.geonames.org/export/zip/%s%s.zip", countryCode, suffix)
	targetFile := fmt.Sprintf("%s%s.txt", countryCode, targetFileSuffix)

	// Download ZIP file
	zipData, err := dd.downloadFile(url)
	if err != nil {
		return nil, err
	}

	// Extract target file from ZIP
	extractedData, err := dd.extractZipFile(zipData, targetFile)
	if err != nil {
		return nil, err
	}

	// Parse postal codes
	return dd.parsePostalCodes(extractedData, countryCode), nil
}

// downloadFile downloads a file and returns its content
func (dd *DataDownloader) downloadFile(url string) ([]byte, error) {
	resp, err := dd.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// downloadJSON downloads and returns JSON data
func (dd *DataDownloader) downloadJSON(url string) ([]byte, error) {
	return dd.downloadFile(url)
}

// extractZipFile extracts a specific file from ZIP data
func (dd *DataDownloader) extractZipFile(zipData []byte, targetFile string) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", err
	}

	for _, file := range reader.File {
		if file.Name == targetFile {
			rc, err := file.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}

			return string(content), nil
		}
	}

	return "", fmt.Errorf("target file %s not found in ZIP archive", targetFile)
}

// parsePostalCodes parses postal code data and validates formats
func (dd *DataDownloader) parsePostalCodes(data, countryCode string) []PostalCode {
	formatRegex := dd.postalCodeRegexs[countryCode]
	if formatRegex == nil {
		fmt.Printf("Warning: No postal code format defined for %s\n", countryCode)
		return []PostalCode{}
	}

	postalCodesSet := make(map[string]bool)
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		postalCode := strings.TrimSpace(fields[1])
		postalCode = strings.ReplaceAll(postalCode, " ", "")

		// Standardize formats
		postalCode = dd.standardizePostalCode(postalCode, countryCode)

		// Validate format
		if !formatRegex.MatchString(postalCode) {
			continue
		}

		postalCodesSet[postalCode] = true
	}

	// Convert set to slice
	var result []PostalCode
	for postalCode := range postalCodesSet {
		result = append(result, PostalCode{
			CountryCode: countryCode,
			PostalCode:  postalCode,
		})
	}

	return result
}

// standardizePostalCode standardizes postal code format for specific countries
func (dd *DataDownloader) standardizePostalCode(postalCode, countryCode string) string {
	switch countryCode {
	case "JP":
		if len(postalCode) == 7 && !strings.Contains(postalCode, "-") {
			return fmt.Sprintf("%s-%s", postalCode[:3], postalCode[3:])
		}
	case "CA":
		if matched, _ := regexp.MatchString(`^[A-Z]\d[A-Z]\d[A-Z]\d$`, postalCode); matched {
			return fmt.Sprintf("%s %s", postalCode[:3], postalCode[3:])
		}
	case "GB":
		if matched, _ := regexp.MatchString(`^[A-Z]{1,2}\d{1,2}[A-Z]?\d[A-Z]{2}$`, postalCode); matched {
			re := regexp.MustCompile(`^([A-Z]{1,2}\d{1,2}[A-Z]?)(\d[A-Z]{2})$`)
			return re.ReplaceAllString(postalCode, "$1 $2")
		}
	case "NL":
		postalCode = strings.ReplaceAll(postalCode, " ", "")
	}

	return postalCode
}

// writeLocationFile writes the location data to a JSON file
func (dd *DataDownloader) writeLocationFile(outputPath string, data LocationData) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Set the absolute path in location.go for consistency
	SetDataFilePath(absPath)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(absPath, jsonData, 0644)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============================================================================
// MAIN COMMAND LINE INTERFACE
// ============================================================================

// RunPostInstall runs the post-installation data download process
func RunPostInstall() error {
	fmt.Println("Starting navii geographical data download...")

	downloader := NewDataDownloader()
	outputPath := "location_data.json"

	if err := downloader.DownloadAndProcessData(outputPath); err != nil {
		return fmt.Errorf("post-install failed: %w", err)
	}

	fmt.Printf("✓ Geographical data successfully downloaded and saved to %s\n", outputPath)
	return nil
}
