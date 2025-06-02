package navii

type LocationData struct {
	CityData map[string]map[string][]string `json:"cityData"`
	ZipData  map[string][]string            `json:"zipData"`
}

// GetLocationData returns empty location data (to be populated by postinstall)
func GetLocationData() *LocationData {
	return &LocationData{
		CityData: make(map[string]map[string][]string),
		ZipData:  make(map[string][]string),
	}
}
