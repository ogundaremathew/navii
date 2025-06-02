package navii

// ============================================================================
// TYPE DEFINITIONS (equivalent to db.types.ts and core.types.ts)
// ============================================================================

// Country represents a country entity
type Country struct {
	ID           *int   `json:"id,omitempty" db:"id"`
	Country      string `json:"country" db:"country"`
	CountryShort string `json:"countryShort" db:"countryShort"`
	Used         bool   `json:"used" db:"used"`
	External     bool   `json:"external" db:"external"`
}

// State represents a state/province entity
type State struct {
	ID           *int   `json:"id,omitempty" db:"id"`
	State        string `json:"state" db:"state"`
	StateShort   string `json:"stateShort" db:"stateShort"`
	CountryShort string `json:"countryShort" db:"countryShort"`
	Used         bool   `json:"used" db:"used"`
	External     bool   `json:"external" db:"external"`
}

// City represents a city entity
type City struct {
	ID           *int    `json:"id,omitempty" db:"id"`
	City         string  `json:"city" db:"city"`
	StateShort   string  `json:"stateShort" db:"stateShort"`
	CountryShort string  `json:"countryShort" db:"countryShort"`
	County       *string `json:"county,omitempty" db:"county"`
	Used         bool    `json:"used" db:"used"`
	External     bool    `json:"external" db:"external"`
}

// Zip represents a postal code entity
type Zip struct {
	ID           *int   `json:"id,omitempty" db:"id"`
	Zip          string `json:"zip" db:"zip"`
	CountryShort string `json:"countryShort" db:"countryShort"`
	Used         bool   `json:"used" db:"used"`
	External     bool   `json:"external" db:"external"`
}

// Query represents a search query entity
type Query struct {
	ID       *int   `json:"id,omitempty" db:"id"`
	Query    string `json:"query" db:"query"`
	Used     bool   `json:"used" db:"used"`
	External bool   `json:"external" db:"external"`
}

// NavSession represents a navigation session
type NavSession struct {
	ID           int     `json:"id" db:"id"`
	Format       string  `json:"format" db:"format"`
	CountryShort string  `json:"countryShort" db:"countryShort"`
	QueryID      *int    `json:"queryId,omitempty" db:"queryId"`
	ZipID        *int    `json:"zipId,omitempty" db:"zipId"`
	CityID       *int    `json:"cityId,omitempty" db:"cityId"`
	StateShort   *string `json:"stateShort,omitempty" db:"stateShort"`
	Page         string  `json:"page" db:"page"`
	Completed    bool    `json:"completed" db:"completed"`
	External     bool    `json:"external" db:"external"`
}

// NavFormat represents different navigation format types
type NavFormat string

// Nav represents navigation data
type Nav struct {
	Query        *string `json:"query,omitempty"`
	Zip          *string `json:"zip,omitempty"`
	City         *string `json:"city,omitempty"`
	State        *string `json:"state,omitempty"`
	StateShort   *string `json:"stateShort,omitempty"`
	Country      *string `json:"country,omitempty"`
	CountryShort *string `json:"countryShort,omitempty"`
	County       *string `json:"county,omitempty"`
}

// PageNav represents pagination information
type PageNav struct {
	Pages []int `json:"pages"`
	Total int   `json:"total"`
}

// NavResponse represents a navigation response
type NavResponse struct {
	Format      NavFormat   `json:"format"`
	Nav         Nav         `json:"nav"`
	Country     string      `json:"country"`
	Placeholder string      `json:"placeholder"`
	Page        interface{} `json:"page"` // Can be PageNav or "completed" or nil
	HasNext     bool        `json:"hasNext"`
}

// InitOptions represents initialization options
type InitOptions struct {
	Format        NavFormat `json:"format"`
	TargetCountry string    `json:"targetCountry"` // ISO2 code or "all"
}

// ICountryShort represents valid ISO2 country codes
var ValidCountryCodes = []string{
	"AD", "AE", "AF", "AG", "AI", "AL", "AM", "AO", "AQ", "AR", "AS", "AT", "AU", "AW", "AX", "AZ",
	"BA", "BB", "BD", "BE", "BF", "BG", "BH", "BI", "BJ", "BL", "BM", "BN", "BO", "BQ", "BR", "BS",
	"BT", "BV", "BW", "BY", "BZ", "CA", "CC", "CD", "CF", "CG", "CH", "CI", "CK", "CL", "CM", "CN",
	"CO", "CR", "CU", "CV", "CW", "CX", "CY", "CZ", "DE", "DJ", "DK", "DM", "DO", "DZ", "EC", "EE",
	"EG", "EH", "ER", "ES", "ET", "FI", "FJ", "FK", "FM", "FO", "FR", "GA", "GB", "GD", "GE", "GF",
	"GG", "GH", "GI", "GL", "GM", "GN", "GP", "GQ", "GR", "GS", "GT", "GU", "GW", "GY", "HK", "HM",
	"HN", "HR", "HT", "HU", "ID", "IE", "IL", "IM", "IN", "IO", "IQ", "IR", "IS", "IT", "JE", "JM",
	"JO", "JP", "KE", "KG", "KH", "KI", "KM", "KN", "KP", "KR", "KW", "KY", "KZ", "LA", "LB", "LC",
	"LI", "LK", "LR", "LS", "LT", "LU", "LV", "LY", "MA", "MC", "MD", "ME", "MF", "MG", "MH", "MK",
	"ML", "MM", "MN", "MO", "MP", "MQ", "MR", "MS", "MT", "MU", "MV", "MW", "MX", "MY", "MZ", "NA",
	"NC", "NE", "NF", "NG", "NI", "NL", "NO", "NP", "NR", "NU", "NZ", "OM", "PA", "PE", "PF", "PG",
	"PH", "PK", "PL", "PM", "PN", "PR", "PS", "PT", "PW", "PY", "QA", "RE", "RO", "RS", "RU", "RW",
	"SA", "SB", "SC", "SD", "SE", "SG", "SH", "SI", "SJ", "SK", "SL", "SM", "SN", "SO", "SR", "SS",
	"ST", "SV", "SX", "SY", "SZ", "TC", "TD", "TF", "TG", "TH", "TJ", "TK", "TL", "TM", "TN", "TO",
	"TR", "TT", "TV", "TW", "TZ", "UA", "UG", "UM", "US", "UY", "UZ", "VA", "VC", "VE", "VG", "VI",
	"VN", "VU", "WF", "WS", "XK", "YE", "YT", "ZA", "ZM", "ZW",
}
