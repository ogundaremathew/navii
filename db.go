// Package yuniq provides geographical navigation and state management functionality
package navii

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB handles database operations
type DB struct {
	db *sql.DB
}

// NewDB creates a new database instance
func NewDB(dbPath string) (*DB, error) {
	if dbPath == "" {
		dbPath = ".yuniq.db"
	}

	database, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{db: database}
	if err := db.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return db, nil
}

// initTables creates database tables
func (db *DB) initTables() error {
	schema := `
		CREATE TABLE IF NOT EXISTS countries (
			countryShort TEXT PRIMARY KEY,
			country TEXT NOT NULL,
			used BOOLEAN NOT NULL DEFAULT 0,
			external BOOLEAN NOT NULL DEFAULT 0,
			UNIQUE(country, countryShort)
		);
		CREATE INDEX IF NOT EXISTS idx_countries_countryShort ON countries(countryShort);

		CREATE TABLE IF NOT EXISTS states (
			stateShort TEXT NOT NULL,
			state TEXT NOT NULL,
			countryShort TEXT NOT NULL,
			used BOOLEAN NOT NULL DEFAULT 0,
			external BOOLEAN NOT NULL DEFAULT 0,
			PRIMARY KEY (stateShort, countryShort),
			FOREIGN KEY (countryShort) REFERENCES countries(countryShort) ON DELETE CASCADE,
			UNIQUE(state, stateShort, countryShort)
		);
		CREATE INDEX IF NOT EXISTS idx_states_countryShort ON states(countryShort);

		CREATE TABLE IF NOT EXISTS cities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			city TEXT NOT NULL,
			stateShort TEXT NOT NULL,
			countryShort TEXT NOT NULL,
			county TEXT,
			used BOOLEAN NOT NULL DEFAULT 0,
			external BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (stateShort, countryShort) REFERENCES states(stateShort, countryShort) ON DELETE CASCADE,
			FOREIGN KEY (countryShort) REFERENCES countries(countryShort) ON DELETE CASCADE,
			UNIQUE(city, stateShort, countryShort)
		);
		CREATE INDEX IF NOT EXISTS idx_cities_stateShort ON cities(stateShort, countryShort);
		CREATE INDEX IF NOT EXISTS idx_cities_countryShort ON cities(countryShort);

		CREATE TABLE IF NOT EXISTS zips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			zip TEXT NOT NULL,
			countryShort TEXT NOT NULL,
			used BOOLEAN NOT NULL DEFAULT 0,
			external BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (countryShort) REFERENCES countries(countryShort) ON DELETE CASCADE,
			UNIQUE(zip, countryShort)
		);
		CREATE INDEX IF NOT EXISTS idx_zips_countryShort ON zips(countryShort);

		CREATE TABLE IF NOT EXISTS queries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT NOT NULL UNIQUE,
			used BOOLEAN NOT NULL DEFAULT 0,
			external BOOLEAN NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS nav_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			format TEXT NOT NULL,
			countryShort TEXT NOT NULL,
			queryId INTEGER,
			zipId INTEGER,
			cityId INTEGER,
			stateShort TEXT,
			page TEXT,
			completed BOOLEAN NOT NULL DEFAULT 0,
			external BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (countryShort) REFERENCES countries(countryShort) ON DELETE CASCADE,
			FOREIGN KEY (queryId) REFERENCES queries(id) ON DELETE SET NULL,
			FOREIGN KEY (zipId) REFERENCES zips(id) ON DELETE SET NULL,
			FOREIGN KEY (cityId) REFERENCES cities(id) ON DELETE SET NULL,
			FOREIGN KEY (stateShort, countryShort) REFERENCES states(stateShort, countryShort) ON DELETE SET NULL
		);
	`

	_, err := db.db.Exec(schema)
	return err
}

// AddCountries adds countries to the database
func (db *DB) AddCountries(countries []Country, external bool) error {
	for _, country := range countries {
		if country.CountryShort == "" || country.Country == "" {
			return fmt.Errorf("all countries must have countryShort and country")
		}
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO countries (countryShort, country, used, external)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, country := range countries {
		_, err := stmt.Exec(country.CountryShort, country.Country, country.Used, external)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddStates adds states to the database
func (db *DB) AddStates(states []State, external bool) error {
	for _, state := range states {
		if state.StateShort == "" || state.State == "" || state.CountryShort == "" {
			return fmt.Errorf("all states must have stateShort, state, and countryShort")
		}
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO states (stateShort, state, countryShort, used, external)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, state := range states {
		_, err := stmt.Exec(state.StateShort, state.State, state.CountryShort, state.Used, external)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddCities adds cities to the database
func (db *DB) AddCities(cities []City, external bool) error {
	for _, city := range cities {
		if city.City == "" || city.StateShort == "" || city.CountryShort == "" {
			return fmt.Errorf("all cities must have city, stateShort, and countryShort")
		}
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO cities (city, stateShort, countryShort, county, used, external)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, city := range cities {
		_, err := stmt.Exec(city.City, city.StateShort, city.CountryShort, city.County, city.Used, external)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddZips adds zip codes to the database
func (db *DB) AddZips(zips []Zip, external bool) error {
	for _, zip := range zips {
		if zip.Zip == "" || zip.CountryShort == "" {
			return fmt.Errorf("all zips must have zip and countryShort")
		}
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO zips (zip, countryShort, used, external)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, zip := range zips {
		_, err := stmt.Exec(zip.Zip, zip.CountryShort, zip.Used, external)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AddQueries adds queries to the database
func (db *DB) AddQueries(queries []string, external bool) error {
	for _, query := range queries {
		if query == "" {
			return fmt.Errorf("all queries must be non-empty strings")
		}
	}

	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO queries (query, used, external)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, query := range queries {
		_, err := stmt.Exec(query, false, external)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ClearQueries removes external queries
func (db *DB) ClearQueries() error {
	_, err := db.db.Exec(`DELETE FROM queries WHERE external = 1`)
	return err
}

// GetQueries retrieves all queries
func (db *DB) GetQueries() ([]Query, error) {
	rows, err := db.db.Query(`SELECT id, query, used, external FROM queries`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []Query
	for rows.Next() {
		var q Query
		err := rows.Scan(&q.ID, &q.Query, &q.Used, &q.External)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}

	return queries, rows.Err()
}

// GetCountries retrieves countries based on target
func (db *DB) GetCountries(targetCountry string) ([]Country, error) {
	var query string
	var args []interface{}

	if targetCountry == "all" {
		query = `SELECT countryShort, country, used, external FROM countries`
	} else {
		query = `SELECT countryShort, country, used, external FROM countries WHERE countryShort = ?`
		args = []interface{}{targetCountry}
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		var c Country
		err := rows.Scan(&c.CountryShort, &c.Country, &c.Used, &c.External)
		if err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}

	return countries, rows.Err()
}

// GetStates retrieves states for given countries
func (db *DB) GetStates(countryShorts []string) ([]State, error) {
	if len(countryShorts) == 0 {
		return []State{}, nil
	}

	placeholders := strings.Repeat("?,", len(countryShorts))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf(`SELECT stateShort, state, countryShort, used, external FROM states WHERE countryShort IN (%s)`, placeholders)

	args := make([]interface{}, len(countryShorts))
	for i, cs := range countryShorts {
		args[i] = cs
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []State
	for rows.Next() {
		var s State
		err := rows.Scan(&s.StateShort, &s.State, &s.CountryShort, &s.Used, &s.External)
		if err != nil {
			return nil, err
		}
		states = append(states, s)
	}

	return states, rows.Err()
}

// GetCities retrieves cities for given countries and states
func (db *DB) GetCities(countryShorts []string, stateShorts []string) ([]City, error) {
	if len(countryShorts) == 0 && len(stateShorts) == 0 {
		return []City{}, nil
	}

	var query string
	var args []interface{}

	if len(stateShorts) > 0 {
		// Build query for state-country combinations
		var conditions []string
		for _, stateShort := range stateShorts {
			for _, countryShort := range countryShorts {
				conditions = append(conditions, "(stateShort = ? AND countryShort = ?)")
				args = append(args, stateShort, countryShort)
			}
		}
		query = fmt.Sprintf(`SELECT id, city, stateShort, countryShort, county, used, external FROM cities WHERE %s`, strings.Join(conditions, " OR "))
	} else if len(countryShorts) > 0 {
		placeholders := strings.Repeat("?,", len(countryShorts))
		placeholders = placeholders[:len(placeholders)-1]
		query = fmt.Sprintf(`SELECT id, city, stateShort, countryShort, county, used, external FROM cities WHERE countryShort IN (%s)`, placeholders)
		for _, cs := range countryShorts {
			args = append(args, cs)
		}
	} else {
		query = `SELECT id, city, stateShort, countryShort, county, used, external FROM cities`
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var c City
		err := rows.Scan(&c.ID, &c.City, &c.StateShort, &c.CountryShort, &c.County, &c.Used, &c.External)
		if err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}

	return cities, rows.Err()
}

// GetZips retrieves zips for given countries
func (db *DB) GetZips(countryShorts []string) ([]Zip, error) {
	if len(countryShorts) == 0 {
		return []Zip{}, nil
	}

	placeholders := strings.Repeat("?,", len(countryShorts))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`SELECT id, zip, countryShort, used, external FROM zips WHERE countryShort IN (%s)`, placeholders)

	args := make([]interface{}, len(countryShorts))
	for i, cs := range countryShorts {
		args[i] = cs
	}

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zips []Zip
	for rows.Next() {
		var z Zip
		err := rows.Scan(&z.ID, &z.Zip, &z.CountryShort, &z.Used, &z.External)
		if err != nil {
			return nil, err
		}
		zips = append(zips, z)
	}

	return zips, rows.Err()
}

// SaveNavSession saves a navigation session
func (db *DB) SaveNavSession(session NavSession) error {
	_, err := db.db.Exec(`
		INSERT INTO nav_sessions (format, countryShort, queryId, zipId, cityId, stateShort, page, completed, external)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.Format, session.CountryShort, session.QueryID, session.ZipID, session.CityID, session.StateShort, session.Page, session.Completed, session.External)
	return err
}

// UpdateNavSession updates a navigation session
func (db *DB) UpdateNavSession(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	var setParts []string
	var args []interface{}

	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}
	args = append(args, id)

	query := fmt.Sprintf("UPDATE nav_sessions SET %s WHERE id = ?", strings.Join(setParts, ", "))
	_, err := db.db.Exec(query, args...)
	return err
}

// GetCurrentNavSession retrieves the current navigation session
func (db *DB) GetCurrentNavSession() (*NavSession, error) {
	var session NavSession
	err := db.db.QueryRow(`SELECT id, format, countryShort, queryId, zipId, cityId, stateShort, page, completed, external FROM nav_sessions WHERE completed = 0 LIMIT 1`).Scan(
		&session.ID, &session.Format, &session.CountryShort, &session.QueryID, &session.ZipID, &session.CityID, &session.StateShort, &session.Page, &session.Completed, &session.External)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetAllNavSessions retrieves all navigation sessions
func (db *DB) GetAllNavSessions() ([]NavSession, error) {
	rows, err := db.db.Query(`SELECT id, format, countryShort, queryId, zipId, cityId, stateShort, page, completed, external FROM nav_sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []NavSession
	for rows.Next() {
		var s NavSession
		err := rows.Scan(&s.ID, &s.Format, &s.CountryShort, &s.QueryID, &s.ZipID, &s.CityID, &s.StateShort, &s.Page, &s.Completed, &s.External)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

// ResetNavSessions deletes all navigation sessions
func (db *DB) ResetNavSessions() error {
	_, err := db.db.Exec(`DELETE FROM nav_sessions`)
	return err
}

// ResetDatabase resets all usage flags and sessions
func (db *DB) ResetDatabase() error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := []string{
		`UPDATE countries SET used = 0`,
		`UPDATE states SET used = 0`,
		`UPDATE cities SET used = 0`,
		`UPDATE zips SET used = 0`,
		`UPDATE queries SET used = 0`,
		`DELETE FROM nav_sessions`,
	}

	for _, query := range queries {
		_, err := tx.Exec(query)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CountTotal returns the total number of countries
func (db *DB) CountTotal() (int, error) {
	var total int
	err := db.db.QueryRow("SELECT COUNT(*) FROM countries").Scan(&total)
	return total, err
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.db.Close()
}
