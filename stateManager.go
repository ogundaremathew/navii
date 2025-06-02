// Package yuniq provides geographical navigation and state management functionality
package navii

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	NavFormatZip                   NavFormat = "zip"
	NavFormatZipCountry            NavFormat = "zip-country"
	NavFormatQueryZip              NavFormat = "query-zip"
	NavFormatQueryZipCountry       NavFormat = "query-zip-country"
	NavFormatCity                  NavFormat = "city"
	NavFormatCityState             NavFormat = "city-state"
	NavFormatCityStateCountry      NavFormat = "city-state-country"
	NavFormatQueryCity             NavFormat = "query-city"
	NavFormatQueryCityState        NavFormat = "query-city-state"
	NavFormatQueryCityStateCountry NavFormat = "query-city-state-country"
	NavFormatState                 NavFormat = "state"
	NavFormatStateCountry          NavFormat = "state-country"
	NavFormatQueryState            NavFormat = "query-state"
	NavFormatQueryStateCountry     NavFormat = "query-state-country"
	NavFormatQueryCounty           NavFormat = "query-county"
	NavFormatQuery                 NavFormat = "query"
	NavFormatCounty                NavFormat = "county"
)

// ============================================================================
// STATE MANAGER (equivalent to stateManager.ts)
// ============================================================================

// StateManager manages geographical navigation state
type StateManager struct {
	db            *DB
	format        *NavFormat
	targetCountry string
	currentNav    *NavResponse
	countries     []Country
	states        []State
	cities        []City
	zips          []Zip
	queries       []Query
	currentIndex  int
	navOrder      []Nav
}

// NewStateManager creates a new state manager
func NewStateManager(dbPath string) (*StateManager, error) {
	if dbPath == "" {
		dbPath = ".yuniq.db"
	}

	db, err := NewDB(dbPath)
	if err != nil {
		return nil, err
	}

	return &StateManager{
		db:            db,
		targetCountry: "all",
		navOrder:      []Nav{},
	}, nil
}

// Init initializes the state manager with given options
func (sm *StateManager) Init(options InitOptions) error {
	sm.format = &options.Format
	sm.targetCountry = options.TargetCountry

	if err := sm.setDefault(); err != nil {
		return err
	}

	countries, err := sm.db.GetCountries(sm.targetCountry)
	if err != nil {
		return err
	}
	sm.countries = countries

	countryShorts := make([]string, len(sm.countries))
	for i, c := range sm.countries {
		countryShorts[i] = c.CountryShort
	}

	states, err := sm.db.GetStates(countryShorts)
	if err != nil {
		return err
	}
	sm.states = states

	stateShorts := make([]string, len(sm.states))
	for i, s := range sm.states {
		stateShorts[i] = s.StateShort
	}

	cities, err := sm.db.GetCities(countryShorts, stateShorts)
	if err != nil {
		return err
	}
	sm.cities = cities

	zips, err := sm.db.GetZips(countryShorts)
	if err != nil {
		return err
	}
	sm.zips = zips

	queries, err := sm.db.GetQueries()
	if err != nil {
		return err
	}
	sm.queries = queries

	sm.currentIndex = 0
	sm.generateNavOrder()
	return sm.restoreOrStartSession()
}

// setDefault populates default data if database is empty
func (sm *StateManager) setDefault() error {
	total, err := sm.db.CountTotal()
	if err != nil {
		return err
	}

	if total > 0 {
		return nil // Already populated
	}

	// Load location data - this would be populated by the postinstall process
	locationData := GetLocationData()

	var allCountries []Country
	var allStates []State
	var allCities []City
	var allZips []Zip

	// Process city data
	for key, value := range locationData.CityData {
		parts := strings.Split(key, "#")
		if len(parts) != 2 {
			continue
		}
		countryShort, countryName := parts[0], parts[1]

		allCountries = append(allCountries, Country{
			Country:      countryName,
			CountryShort: countryShort,
			External:     false,
			Used:         false,
		})

		for k, cities := range value {
			stateParts := strings.Split(k, "##")
			if len(stateParts) != 2 {
				continue
			}
			stateShort, stateName := stateParts[0], stateParts[1]

			allStates = append(allStates, State{
				State:        stateName,
				StateShort:   stateShort,
				CountryShort: countryShort,
				Used:         false,
				External:     false,
			})

			for _, city := range cities {
				allCities = append(allCities, City{
					City:         city,
					StateShort:   stateShort,
					CountryShort: countryShort,
					Used:         false,
					External:     false,
				})
			}
		}
	}

	// Process zip data
	for countryShort, zips := range locationData.ZipData {
		for _, zip := range zips {
			allZips = append(allZips, Zip{
				CountryShort: countryShort,
				Zip:          zip,
				Used:         false,
				External:     false,
			})
		}
	}

	// Insert data in transaction
	return sm.executeTransaction(func() error {
		if err := sm.db.AddCountries(allCountries, false); err != nil {
			return err
		}
		if err := sm.db.AddStates(allStates, false); err != nil {
			return err
		}
		if err := sm.db.AddCities(allCities, false); err != nil {
			return err
		}
		return sm.db.AddZips(allZips, false)
	})
}

// executeTransaction executes a function within a database transaction
func (sm *StateManager) executeTransaction(fn func() error) error {
	return fn() // Simplified - individual methods handle transactions
}

// generateNavOrder generates the navigation order based on format
func (sm *StateManager) generateNavOrder() {
	sm.navOrder = []Nav{}

	for _, country := range sm.countries {
		countryStates := sm.getStatesByCountry(country.CountryShort)
		stateShorts := make([]string, len(countryStates))
		for i, s := range countryStates {
			stateShorts[i] = s.StateShort
		}

		countryCities := sm.getCitiesByCountry(country.CountryShort)
		countryZips := sm.getZipsByCountry(country.CountryShort)

		if strings.HasPrefix(string(*sm.format), "query-") {
			for _, query := range sm.queries {
				sm.addNavForQuery(&query, country, countryStates, countryCities, countryZips)
			}
		} else {
			sm.addNavForQuery(nil, country, countryStates, countryCities, countryZips)
		}
	}
}

// Helper methods for filtering data
func (sm *StateManager) getStatesByCountry(countryShort string) []State {
	var result []State
	for _, s := range sm.states {
		if s.CountryShort == countryShort {
			result = append(result, s)
		}
	}
	return result
}

func (sm *StateManager) getCitiesByCountry(countryShort string) []City {
	var result []City
	for _, c := range sm.cities {
		if c.CountryShort == countryShort {
			result = append(result, c)
		}
	}
	return result
}

func (sm *StateManager) getZipsByCountry(countryShort string) []Zip {
	var result []Zip
	for _, z := range sm.zips {
		if z.CountryShort == countryShort {
			result = append(result, z)
		}
	}
	return result
}

func (sm *StateManager) findStateByShort(stateShort string, states []State) *State {
	for _, s := range states {
		if s.StateShort == stateShort {
			return &s
		}
	}
	return nil
}

// addNavForQuery adds navigation entries for a specific query
func (sm *StateManager) addNavForQuery(query *Query, country Country, states []State, cities []City, zips []Zip) {
	switch *sm.format {
	case NavFormatZip:
		for _, zip := range zips {
			sm.navOrder = append(sm.navOrder, Nav{
				Zip:     &zip.Zip,
				Country: &country.CountryShort,
			})
		}

	case NavFormatZipCountry:
		for _, zip := range zips {
			sm.navOrder = append(sm.navOrder, Nav{
				Zip:          &zip.Zip,
				Country:      &country.CountryShort,
				CountryShort: &country.CountryShort,
			})
		}

	case NavFormatQueryZip:
		if query != nil {
			for _, zip := range zips {
				sm.navOrder = append(sm.navOrder, Nav{
					Query:   &query.Query,
					Zip:     &zip.Zip,
					Country: &country.CountryShort,
				})
			}
		}

	case NavFormatQueryZipCountry:
		if query != nil {
			for _, zip := range zips {
				sm.navOrder = append(sm.navOrder, Nav{
					Query:        &query.Query,
					Zip:          &zip.Zip,
					Country:      &country.CountryShort,
					CountryShort: &country.CountryShort,
				})
			}
		}

	case NavFormatCity:
		for _, city := range cities {
			sm.navOrder = append(sm.navOrder, Nav{
				City:    &city.City,
				Country: &country.CountryShort,
			})
		}

	case NavFormatCityState:
		for _, city := range cities {
			if state := sm.findStateByShort(city.StateShort, states); state != nil {
				sm.navOrder = append(sm.navOrder, Nav{
					City:       &city.City,
					State:      &state.State,
					StateShort: &state.StateShort,
					Country:    &country.CountryShort,
				})
			}
		}

	case NavFormatCityStateCountry:
		for _, city := range cities {
			if state := sm.findStateByShort(city.StateShort, states); state != nil {
				sm.navOrder = append(sm.navOrder, Nav{
					City:         &city.City,
					State:        &state.State,
					StateShort:   &state.StateShort,
					Country:      &country.CountryShort,
					CountryShort: &country.CountryShort,
				})
			}
		}

	case NavFormatQueryCity:
		if query != nil {
			for _, city := range cities {
				sm.navOrder = append(sm.navOrder, Nav{
					Query:   &query.Query,
					City:    &city.City,
					Country: &country.CountryShort,
				})
			}
		}

	case NavFormatQueryCityState:
		if query != nil {
			for _, city := range cities {
				if state := sm.findStateByShort(city.StateShort, states); state != nil {
					sm.navOrder = append(sm.navOrder, Nav{
						Query:      &query.Query,
						City:       &city.City,
						State:      &state.State,
						StateShort: &state.StateShort,
						Country:    &country.CountryShort,
					})
				}
			}
		}

	case NavFormatQueryCityStateCountry:
		if query != nil {
			for _, city := range cities {
				if state := sm.findStateByShort(city.StateShort, states); state != nil {
					sm.navOrder = append(sm.navOrder, Nav{
						Query:        &query.Query,
						City:         &city.City,
						State:        &state.State,
						StateShort:   &state.StateShort,
						Country:      &country.CountryShort,
						CountryShort: &country.CountryShort,
					})
				}
			}
		}

	case NavFormatState:
		for _, state := range states {
			sm.navOrder = append(sm.navOrder, Nav{
				State:      &state.State,
				StateShort: &state.StateShort,
				Country:    &country.CountryShort,
			})
		}

	case NavFormatStateCountry:
		for _, state := range states {
			sm.navOrder = append(sm.navOrder, Nav{
				State:        &state.State,
				StateShort:   &state.StateShort,
				Country:      &country.CountryShort,
				CountryShort: &country.CountryShort,
			})
		}

	case NavFormatQueryState:
		if query != nil {
			for _, state := range states {
				sm.navOrder = append(sm.navOrder, Nav{
					Query:      &query.Query,
					State:      &state.State,
					StateShort: &state.StateShort,
					Country:    &country.CountryShort,
				})
			}
		}

	case NavFormatQueryStateCountry:
		if query != nil {
			for _, state := range states {
				sm.navOrder = append(sm.navOrder, Nav{
					Query:        &query.Query,
					State:        &state.State,
					StateShort:   &state.StateShort,
					Country:      &country.CountryShort,
					CountryShort: &country.CountryShort,
				})
			}
		}

	case NavFormatQueryCounty:
		if query != nil {
			for _, city := range cities {
				if city.County != nil {
					sm.navOrder = append(sm.navOrder, Nav{
						Query:   &query.Query,
						County:  city.County,
						Country: &country.CountryShort,
					})
				}
			}
		}

	case NavFormatQuery:
		if query != nil {
			sm.navOrder = append(sm.navOrder, Nav{
				Query:   &query.Query,
				Country: &country.CountryShort,
			})
		}

	case NavFormatCounty:
		for _, city := range cities {
			if city.County != nil {
				sm.navOrder = append(sm.navOrder, Nav{
					County:  city.County,
					Country: &country.CountryShort,
				})
			}
		}
	}
}

// restoreOrStartSession restores existing session or starts new one
func (sm *StateManager) restoreOrStartSession() error {
	session, err := sm.db.GetCurrentNavSession()
	if err != nil {
		return err
	}

	if session != nil {
		// Restore existing session
		country := sm.findCountry(session.CountryShort)
		var query *Query
		var zip *Zip
		var city *City
		var state *State

		if session.QueryID != nil {
			query = sm.findQuery(*session.QueryID)
		}
		if session.ZipID != nil {
			zip = sm.findZip(*session.ZipID)
		}
		if session.CityID != nil {
			city = sm.findCity(*session.CityID)
		}
		if session.StateShort != nil {
			state = sm.findState(*session.StateShort)
		}

		sm.currentIndex = sm.findNavIndex(*session, country, query, zip, city, state)
		sm.currentNav = sm.buildNavResponse(*session, country, query, zip, city, state)
	} else {
		// Start new session
		sm.currentNav = sm.buildNavResponseFromIndex(0)
		if sm.currentNav != nil {
			return sm.saveCurrentSession()
		}
	}

	return nil
}

// Helper methods for finding entities
func (sm *StateManager) findCountry(countryShort string) *Country {
	for _, c := range sm.countries {
		if c.CountryShort == countryShort {
			return &c
		}
	}
	return nil
}

func (sm *StateManager) findQuery(id int) *Query {
	for _, q := range sm.queries {
		if q.ID != nil && *q.ID == id {
			return &q
		}
	}
	return nil
}

func (sm *StateManager) findZip(id int) *Zip {
	for _, z := range sm.zips {
		if z.ID != nil && *z.ID == id {
			return &z
		}
	}
	return nil
}

func (sm *StateManager) findCity(id int) *City {
	for _, c := range sm.cities {
		if c.ID != nil && *c.ID == id {
			return &c
		}
	}
	return nil
}

func (sm *StateManager) findState(stateShort string) *State {
	for _, s := range sm.states {
		if s.StateShort == stateShort {
			return &s
		}
	}
	return nil
}

// findNavIndex finds the index of a navigation item
func (sm *StateManager) findNavIndex(session NavSession, country *Country, query *Query, zip *Zip, city *City, state *State) int {
	for i, nav := range sm.navOrder {
		if sm.navMatches(nav, country, query, zip, city, state) {
			return i
		}
	}
	return 0
}

// navMatches checks if a nav item matches the given entities
func (sm *StateManager) navMatches(nav Nav, country *Country, query *Query, zip *Zip, city *City, state *State) bool {
	queryMatch := (nav.Query == nil && query == nil) || (nav.Query != nil && query != nil && *nav.Query == query.Query)
	zipMatch := (nav.Zip == nil && zip == nil) || (nav.Zip != nil && zip != nil && *nav.Zip == zip.Zip)
	cityMatch := (nav.City == nil && city == nil) || (nav.City != nil && city != nil && *nav.City == city.City)
	stateMatch := (nav.State == nil && state == nil) || (nav.State != nil && state != nil && *nav.State == state.State)
	countryMatch := (nav.Country == nil && country == nil) || (nav.Country != nil && country != nil && *nav.Country == country.CountryShort)

	return queryMatch && zipMatch && cityMatch && stateMatch && countryMatch
}

// buildNavResponse builds a navigation response from session data
func (sm *StateManager) buildNavResponse(session NavSession, country *Country, query *Query, zip *Zip, city *City, state *State) *NavResponse {
	var page interface{}
	if session.Page == "completed" {
		page = "completed"
	} else if session.Page != "" {
		var pageNav PageNav
		json.Unmarshal([]byte(session.Page), &pageNav)
		page = pageNav
	}

	nav := Nav{}
	if query != nil {
		nav.Query = &query.Query
	}
	if zip != nil {
		nav.Zip = &zip.Zip
	}
	if city != nil {
		nav.City = &city.City
		nav.County = city.County
	}
	if state != nil {
		nav.State = &state.State
		nav.StateShort = &state.StateShort
	}
	if country != nil {
		nav.Country = &country.Country
		nav.CountryShort = &country.CountryShort
	}

	countryShort := ""
	if country != nil {
		countryShort = country.CountryShort
	}

	return &NavResponse{
		Format:      NavFormat(session.Format),
		Nav:         nav,
		Country:     countryShort,
		Placeholder: sm.generatePlaceholder(nav),
		Page:        page,
		HasNext:     sm.currentIndex < len(sm.navOrder)-1,
	}
}

// buildNavResponseFromIndex builds a navigation response from an index
func (sm *StateManager) buildNavResponseFromIndex(index int) *NavResponse {
	if index >= len(sm.navOrder) {
		return nil
	}

	nav := sm.navOrder[index]
	country := sm.findCountry(*nav.Country)

	countryName := ""
	if country != nil {
		countryName = country.CountryShort
	}

	return &NavResponse{
		Format:      *sm.format,
		Nav:         nav,
		Country:     countryName,
		Placeholder: sm.generatePlaceholder(nav),
		Page:        nil,
		HasNext:     index < len(sm.navOrder)-1,
	}
}

// generatePlaceholder generates a placeholder string from navigation data
func (sm *StateManager) generatePlaceholder(nav Nav) string {
	var parts []string

	if nav.Query != nil {
		parts = append(parts, *nav.Query)
	}
	if nav.City != nil {
		parts = append(parts, *nav.City)
	} else if nav.Zip != nil {
		parts = append(parts, *nav.Zip)
	} else if nav.State != nil {
		parts = append(parts, *nav.State)
	} else if nav.County != nil {
		parts = append(parts, *nav.County)
	}

	if len(parts) == 0 {
		return "Unknown"
	}

	return strings.Join(parts, "##")
}

// saveCurrentSession saves the current navigation session
func (sm *StateManager) saveCurrentSession() error {
	if sm.currentNav == nil {
		return nil
	}

	country := sm.findCountry(sm.currentNav.Country)
	var query *Query
	var zip *Zip
	var city *City
	var state *State

	if sm.currentNav.Nav.Query != nil {
		query = sm.findQueryByText(*sm.currentNav.Nav.Query)
	}
	if sm.currentNav.Nav.Zip != nil {
		zip = sm.findZipByText(*sm.currentNav.Nav.Zip)
	}
	if sm.currentNav.Nav.City != nil {
		city = sm.findCityByText(*sm.currentNav.Nav.City)
	}
	if sm.currentNav.Nav.StateShort != nil {
		state = sm.findState(*sm.currentNav.Nav.StateShort)
	}

	pageJSON := ""
	if sm.currentNav.Page != nil {
		pageBytes, _ := json.Marshal(sm.currentNav.Page)
		pageJSON = string(pageBytes)
	}

	session := NavSession{
		Format:       string(sm.currentNav.Format),
		CountryShort: country.CountryShort,
		Page:         pageJSON,
		Completed:    false,
		External:     true,
	}

	if query != nil && query.ID != nil {
		session.QueryID = query.ID
	}
	if zip != nil && zip.ID != nil {
		session.ZipID = zip.ID
	}
	if city != nil && city.ID != nil {
		session.CityID = city.ID
	}
	if state != nil {
		session.StateShort = &state.StateShort
	}

	if err := sm.db.SaveNavSession(session); err != nil {
		return err
	}

	// Mark entities as used
	return sm.markEntitiesAsUsed(country, query, zip, city, state)
}

// Helper methods for finding entities by text
func (sm *StateManager) findQueryByText(queryText string) *Query {
	for _, q := range sm.queries {
		if q.Query == queryText {
			return &q
		}
	}
	return nil
}

func (sm *StateManager) findZipByText(zipText string) *Zip {
	for _, z := range sm.zips {
		if z.Zip == zipText {
			return &z
		}
	}
	return nil
}

func (sm *StateManager) findCityByText(cityText string) *City {
	for _, c := range sm.cities {
		if c.City == cityText {
			return &c
		}
	}
	return nil
}

// markEntitiesAsUsed marks entities as used in the database
func (sm *StateManager) markEntitiesAsUsed(country *Country, query *Query, zip *Zip, city *City, state *State) error {
	if country != nil {
		_, err := sm.db.db.Exec(`UPDATE countries SET used = 1 WHERE countryShort = ?`, country.CountryShort)
		if err != nil {
			return err
		}
	}

	if query != nil && query.ID != nil {
		_, err := sm.db.db.Exec(`UPDATE queries SET used = 1 WHERE id = ?`, *query.ID)
		if err != nil {
			return err
		}
	}

	if zip != nil && zip.ID != nil {
		_, err := sm.db.db.Exec(`UPDATE zips SET used = 1 WHERE id = ?`, *zip.ID)
		if err != nil {
			return err
		}
	}

	if city != nil && city.ID != nil {
		_, err := sm.db.db.Exec(`UPDATE cities SET used = 1 WHERE id = ?`, *city.ID)
		if err != nil {
			return err
		}
	}

	if state != nil {
		_, err := sm.db.db.Exec(`UPDATE states SET used = 1 WHERE stateShort = ? AND countryShort = ?`, state.StateShort, state.CountryShort)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetNav returns the current navigation response
func (sm *StateManager) GetNav() *NavResponse {
	return sm.currentNav
}

// GetNextNav gets the next navigation item
func (sm *StateManager) GetNextNav() (*NavResponse, error) {
	session, err := sm.db.GetCurrentNavSession()
	if err != nil {
		return nil, err
	}

	if session != nil && !session.Completed {
		return sm.currentNav, nil
	}

	sm.currentIndex++
	sm.currentNav = sm.buildNavResponseFromIndex(sm.currentIndex)

	if sm.currentNav != nil {
		return sm.currentNav, sm.saveCurrentSession()
	}

	return sm.currentNav, nil
}

// GetCurrentNav returns the current navigation response
func (sm *StateManager) GetCurrentNav() *NavResponse {
	return sm.currentNav
}

// SetPageNav sets pagination information
func (sm *StateManager) SetPageNav(totalPages int, pages []int) error {
	if sm.currentNav == nil {
		return nil
	}

	pageNav := PageNav{
		Pages: pages,
		Total: totalPages,
	}

	sm.currentNav.Page = pageNav

	session, err := sm.db.GetCurrentNavSession()
	if err != nil {
		return err
	}

	if session != nil {
		pageJSON, _ := json.Marshal(pageNav)
		return sm.db.UpdateNavSession(session.ID, map[string]interface{}{
			"page": string(pageJSON),
		})
	}

	return nil
}

// MarkPageAsDone marks a page as completed
func (sm *StateManager) MarkPageAsDone(page int) error {
	if sm.currentNav == nil || sm.currentNav.Page == "completed" {
		return nil
	}

	pageNav, ok := sm.currentNav.Page.(PageNav)
	if !ok {
		return nil
	}

	// Check if page is already marked
	for _, p := range pageNav.Pages {
		if p == page {
			return nil
		}
	}

	pageNav.Pages = append(pageNav.Pages, page)
	sort.Ints(pageNav.Pages)

	session, err := sm.db.GetCurrentNavSession()
	if err != nil {
		return err
	}

	if session != nil {
		pageJSON, _ := json.Marshal(pageNav)
		err = sm.db.UpdateNavSession(session.ID, map[string]interface{}{
			"page": string(pageJSON),
		})
		if err != nil {
			return err
		}

		if len(pageNav.Pages) == pageNav.Total {
			return sm.MarkComplete()
		}
	}

	return nil
}

// MarkComplete marks the current navigation as complete
func (sm *StateManager) MarkComplete() error {
	session, err := sm.db.GetCurrentNavSession()
	if err != nil {
		return err
	}

	if session != nil {
		err = sm.db.UpdateNavSession(session.ID, map[string]interface{}{
			"completed": true,
		})
		if err != nil {
			return err
		}

		sm.currentNav.Page = "completed"
	}

	return nil
}

// AddSearchQueries adds search queries
func (sm *StateManager) AddSearchQueries(queries []string) error {
	if len(queries) == 0 {
		return nil
	}

	if err := sm.db.AddQueries(queries, true); err != nil {
		return err
	}

	updatedQueries, err := sm.db.GetQueries()
	if err != nil {
		return err
	}
	sm.queries = updatedQueries
	sm.generateNavOrder()
	return nil
}

// ClearSearchQueries clears all search queries
func (sm *StateManager) ClearSearchQueries() error {
	if err := sm.db.ClearQueries(); err != nil {
		return err
	}

	sm.queries = []Query{}
	sm.generateNavOrder()
	return nil
}

// ResetNav resets navigation sessions
func (sm *StateManager) ResetNav() error {
	if err := sm.db.ResetNavSessions(); err != nil {
		return err
	}

	sm.currentIndex = 0
	sm.currentNav = nil
	return sm.restoreOrStartSession()
}

// AddSearchQuery adds a single search query
func (sm *StateManager) AddSearchQuery(query string) error {
	if query == "" {
		return nil
	}

	return sm.AddSearchQueries([]string{query})
}

// AddCities adds cities to the database
func (sm *StateManager) AddCities(cities []struct {
	City         string `json:"city"`
	State        string `json:"state"`
	StateShort   string `json:"stateShort"`
	CountryShort string `json:"countryShort"`
}) error {
	if len(cities) == 0 {
		return nil
	}

	for _, city := range cities {
		if city.City == "" || city.State == "" || city.StateShort == "" || city.CountryShort == "" {
			return fmt.Errorf("all cities must have city, state, stateShort, and countryShort")
		}
	}

	var dbCities []City
	for _, city := range cities {
		dbCities = append(dbCities, City{
			City:         city.City,
			StateShort:   city.StateShort,
			CountryShort: city.CountryShort,
			Used:         false,
			External:     true,
		})
	}

	if err := sm.db.AddCities(dbCities, true); err != nil {
		return err
	}

	return sm.refreshData()
}

// AddStates adds states to the database
func (sm *StateManager) AddStates(states []struct {
	State        string  `json:"state"`
	StateShort   string  `json:"stateShort"`
	County       *string `json:"county,omitempty"`
	CountryShort string  `json:"countryShort"`
}) error {
	if len(states) == 0 {
		return nil
	}

	for _, state := range states {
		if state.State == "" || state.StateShort == "" || state.CountryShort == "" {
			return fmt.Errorf("all states must have state, stateShort, and countryShort")
		}
	}

	var dbStates []State
	for _, state := range states {
		dbStates = append(dbStates, State{
			State:        state.State,
			StateShort:   state.StateShort,
			CountryShort: state.CountryShort,
			Used:         false,
			External:     true,
		})
	}

	if err := sm.db.AddStates(dbStates, true); err != nil {
		return err
	}

	return sm.refreshData()
}

// AddCountries adds countries to the database
func (sm *StateManager) AddCountries(countries []struct {
	Country      string `json:"country"`
	CountryShort string `json:"countryShort"`
}) error {
	if len(countries) == 0 {
		return nil
	}

	for _, country := range countries {
		if country.CountryShort == "" || country.Country == "" {
			return fmt.Errorf("all countries must have countryShort and country name")
		}
	}

	var dbCountries []Country
	for _, country := range countries {
		dbCountries = append(dbCountries, Country{
			Country:      country.Country,
			CountryShort: country.CountryShort,
			Used:         false,
			External:     true,
		})
	}

	if err := sm.db.AddCountries(dbCountries, true); err != nil {
		return err
	}

	return sm.refreshData()
}

// refreshData refreshes all data from database
func (sm *StateManager) refreshData() error {
	countries, err := sm.db.GetCountries(sm.targetCountry)
	if err != nil {
		return err
	}
	sm.countries = countries

	countryShorts := make([]string, len(sm.countries))
	for i, c := range sm.countries {
		countryShorts[i] = c.CountryShort
	}

	states, err := sm.db.GetStates(countryShorts)
	if err != nil {
		return err
	}
	sm.states = states

	stateShorts := make([]string, len(sm.states))
	for i, s := range sm.states {
		stateShorts[i] = s.StateShort
	}

	cities, err := sm.db.GetCities(countryShorts, stateShorts)
	if err != nil {
		return err
	}
	sm.cities = cities

	sm.generateNavOrder()
	return nil
}

// Debug prints debug information
func (sm *StateManager) Debug() {
	fmt.Printf("StateManager Debug Info:\n")
	fmt.Printf("Format: %v\n", sm.format)
	fmt.Printf("TargetCountry: %s\n", sm.targetCountry)
	fmt.Printf("CurrentNav: %+v\n", sm.currentNav)
	fmt.Printf("NavOrderLength: %d\n", len(sm.navOrder))
	fmt.Printf("CurrentIndex: %d\n", sm.currentIndex)
	fmt.Printf("Queries: %d\n", len(sm.queries))
	fmt.Printf("Countries: %d\n", len(sm.countries))
	fmt.Printf("States: %d\n", len(sm.states))
	fmt.Printf("Cities: %d\n", len(sm.cities))
	fmt.Printf("Zips: %d\n", len(sm.zips))
}

// Populate populates the database with sample data
func (sm *StateManager) Populate() error {
	countries := []Country{
		{
			Country:      "United States",
			CountryShort: "US",
			Used:         false,
			External:     true,
		},
	}

	states := []State{
		{
			State:        "California",
			StateShort:   "CA",
			CountryShort: "US",
			Used:         false,
			External:     true,
		},
	}

	zips := []Zip{
		{
			Zip:          "90001",
			CountryShort: "US",
			Used:         false,
			External:     true,
		},
	}

	queries := []string{"Realtor", "Restaurant"}

	return sm.executeTransaction(func() error {
		if err := sm.db.AddCountries(countries, true); err != nil {
			return err
		}
		if err := sm.db.AddStates(states, true); err != nil {
			return err
		}
		if err := sm.db.AddZips(zips, true); err != nil {
			return err
		}
		if err := sm.db.AddQueries(queries, true); err != nil {
			return err
		}

		sm.generateNavOrder()
		return nil
	})
}

// ResetDatabase resets the database
func (sm *StateManager) ResetDatabase() error {
	if err := sm.db.ResetDatabase(); err != nil {
		return err
	}

	return sm.refreshData()
}

// Close closes the state manager and database connection
func (sm *StateManager) Close() error {
	return sm.db.Close()
}
