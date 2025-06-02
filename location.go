package navii

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type LocationData struct {
	CityData map[string]map[string][]string `json:"cityData"`
	ZipData  map[string][]string            `json:"zipData"`
}

// cachedLocationData holds the loaded data to avoid repeated file reads
var cachedLocationData *LocationData
var dataFilePath string

// SetDataFilePath sets the absolute path to the location data JSON file
func SetDataFilePath(absolutePath string) {
	dataFilePath = absolutePath
	// Clear cache when path changes
	cachedLocationData = nil
}

// GetDataFilePath returns the current data file path
func GetDataFilePath() string {
	if dataFilePath != "" {
		return dataFilePath
	}
	// Default fallback: location_data.json in the same directory as this Go file
	return getDefaultDataFilePath()
}

// getDefaultDataFilePath returns the default path for the data file
func getDefaultDataFilePath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "location_data.json" // fallback to current directory
	}
	return filepath.Join(filepath.Dir(filename), "location_data.json")
}

// GetLocationData returns the populated location data from JSON file if available,
// otherwise returns empty location data structure
func GetLocationData() *LocationData {
	// Return cached data if already loaded
	if cachedLocationData != nil {
		return cachedLocationData
	}

	// Try to load data from JSON file
	data, err := loadLocationDataFromJSON()
	if err == nil {
		cachedLocationData = data
		return cachedLocationData
	}

	// Return empty structure if no data file exists or loading failed
	return &LocationData{
		CityData: make(map[string]map[string][]string),
		ZipData:  make(map[string][]string),
	}
}

// GetLocationDataFromPath loads location data from a specific absolute path
func GetLocationDataFromPath(absolutePath string) (*LocationData, error) {
	return loadLocationDataFromPath(absolutePath)
}

// loadLocationDataFromJSON loads location data from the configured JSON file path
func loadLocationDataFromJSON() (*LocationData, error) {
	jsonPath := GetDataFilePath()
	return loadLocationDataFromPath(jsonPath)
}

// loadLocationDataFromPath loads location data from a specific file path
func loadLocationDataFromPath(filePath string) (*LocationData, error) {
	// Convert to absolute path if not already
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Read the JSON file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// Parse JSON data
	var locationData LocationData
	if err := json.Unmarshal(data, &locationData); err != nil {
		return nil, err
	}

	return &locationData, nil
}

// IsDataPopulated checks if geographical data has been downloaded and populated
func IsDataPopulated() bool {
	data := GetLocationData()
	return len(data.CityData) > 0 || len(data.ZipData) > 0
}

// GetCitiesForCountryState returns cities for a specific country and state
func GetCitiesForCountryState(countryCode, stateCode string) []string {
	data := GetLocationData()

	// Find country key
	for countryKey, states := range data.CityData {
		if len(countryKey) >= 2 && countryKey[:2] == countryCode {
			// Find state key
			for stateKey, cities := range states {
				if len(stateKey) >= len(stateCode) && stateKey[:len(stateCode)] == stateCode {
					return cities
				}
			}
		}
	}

	return []string{}
}

// GetPostalCodesForCountry returns postal codes for a specific country
func GetPostalCodesForCountry(countryCode string) []string {
	data := GetLocationData()
	return data.ZipData[countryCode]
}

// GetAvailableCountries returns a list of available country codes
func GetAvailableCountries() []string {
	data := GetLocationData()
	countries := make([]string, 0, len(data.CityData))

	for countryKey := range data.CityData {
		if len(countryKey) >= 2 {
			countryCode := countryKey[:2]
			// Check if already added
			found := false
			for _, existing := range countries {
				if existing == countryCode {
					found = true
					break
				}
			}
			if !found {
				countries = append(countries, countryCode)
			}
		}
	}

	return countries
}
