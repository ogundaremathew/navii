# Navii

🌍 **Navii** is a powerful Go package that provides comprehensive geographical navigation and state management functionality. It enables applications to systematically navigate through geographical data including countries, states, cities, and postal codes with intelligent session management and multiple navigation formats.

## ✨ Features

- **🗺️ Comprehensive Geographical Data**: Download and process countries, states, cities, and postal codes from reliable sources
- **📮 Postal Code Validation**: Built-in validation for postal code formats across multiple countries (US, CA, GB, DE, JP, FR, IN, AU, NL, IE)
- **🔄 State Management**: Intelligent session management with SQLite backend for persistent navigation state
- **📊 Multiple Navigation Formats**: Support for various navigation patterns including zip-based, city-based, state-based, and query-based navigation
- **🎯 Country Targeting**: Focus navigation on specific countries or navigate globally
- **💾 Efficient Storage**: SQLite-based storage with optimized queries and indexing
- **🔍 Smart Pagination**: Built-in pagination support for large datasets

## 🚀 Installation

Install Navii using `go get`:

```bash
go get github.com/ogundaremathew/navii
```

## 📋 Requirements

- Go 1.21.1 or higher
- SQLite3 (automatically handled by `github.com/mattn/go-sqlite3`)

## ⚡ Quick Start

Get up and running with Navii in just a few commands:

```bash
# 1. Install the package
go get github.com/ogundaremathew/navii

# 2. Navigate to your project directory
cd your-project

# 3. Install the CLI tool
go install github.com/ogundaremathew/navii/cmd/navii

# 4. Download geographical data
navii -download-data

# 5. Start using Navii in your code!
```

That's it! You're ready to use Navii's geographical navigation features.

## 🔧 Setup & Data Initialization

Navii requires geographical data to function. After installation, you can automatically download and process the data using one of these methods:

### Method 1: Using Make (Recommended)

```bash
# Install the CLI tool globally
make install

# Download geographical data
navii -download-data
```

### Method 2: Using Go Commands

```bash
# Install the CLI tool
go install ./cmd/navii

# Download geographical data
navii -download-data
```

### Method 3: Direct Build and Run

```bash
# Build and download data in one step
make download-data
```

### Method 4: Programmatic Download

If you prefer to integrate the download into your application:

```go
package main

import (
	"log"
	"github.com/ogundaremathew/navii"
)

func main() {
	// Download and process geographical data
	downloader := navii.NewDataDownloader()
	err := downloader.DownloadAndProcessData("location_data.json")
	if err != nil {
		log.Fatalf("Failed to download geographical data: %v", err)
	}
	
	// Initialize your application
	// (See usage examples below)
}
```

## 📖 Usage

### Basic State Manager Setup

```go
package main

import (
	"fmt"
	"log"
	"github.com/ogundaremathew/navii"
)

func main() {
	// Create a new state manager
	sm, err := navii.NewStateManager("navigation.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sm.Close()

	// Initialize with options
	err = sm.Init(navii.InitOptions{
		Format:        navii.NavFormatCityStateCountry,
		TargetCountry: "US", // Focus on US data
	})
	if err != nil {
		log.Fatal(err)
	}

	// Populate the database with geographical data
	err = sm.Populate()
	if err != nil {
		log.Fatal(err)
	}

	// Start navigation
	for {
		nav := sm.GetNav()
		if nav == nil {
			fmt.Println("Navigation completed!")
			break
		}

		fmt.Printf("Current: %+v\n", nav.Nav)
		fmt.Printf("Format: %s\n", nav.Format)
		fmt.Printf("Country: %s\n", nav.Country)
		fmt.Printf("Has Next: %t\n", nav.HasNext)

		// Move to next navigation item
		nextNav, err := sm.GetNextNav()
		if err != nil {
			log.Printf("Error getting next nav: %v", err)
			break
		}
		if nextNav == nil {
			break
		}
	}
}
```

### Navigation Formats

Navii supports multiple navigation formats to suit different use cases:

```go
// Postal code navigation
sm.Init(navii.InitOptions{
	Format:        navii.NavFormatZip,
	TargetCountry: "US",
})

// City, state, and country navigation
sm.Init(navii.InitOptions{
	Format:        navii.NavFormatCityStateCountry,
	TargetCountry: "all", // All countries
})

// Query-based navigation
sm.Init(navii.InitOptions{
	Format:        navii.NavFormatQuery,
	TargetCountry: "CA",
})
```

### Available Navigation Formats

| Format | Description |
|--------|-------------|
| `NavFormatZip` | Navigate through postal codes |
| `NavFormatZipCountry` | Postal codes with country context |
| `NavFormatCity` | Navigate through cities |
| `NavFormatCityState` | Cities with state context |
| `NavFormatCityStateCountry` | Cities with state and country context |
| `NavFormatState` | Navigate through states/provinces |
| `NavFormatStateCountry` | States with country context |
| `NavFormatQuery` | Custom query-based navigation |
| `NavFormatQueryZip` | Query with postal code context |
| `NavFormatQueryCity` | Query with city context |
| `NavFormatQueryState` | Query with state context |

### Working with Navigation Data

```go
// Get current navigation item
nav := sm.GetNav()
if nav != nil {
	// Access navigation data
	if nav.Nav.City != nil {
		fmt.Printf("City: %s\n", *nav.Nav.City)
	}
	if nav.Nav.State != nil {
		fmt.Printf("State: %s\n", *nav.Nav.State)
	}
	if nav.Nav.Country != nil {
		fmt.Printf("Country: %s\n", *nav.Nav.Country)
	}
	if nav.Nav.Zip != nil {
		fmt.Printf("Postal Code: %s\n", *nav.Nav.Zip)
	}
}

// Check pagination information
if pageNav, ok := nav.Page.(navii.PageNav); ok {
	fmt.Printf("Total pages: %d\n", pageNav.Total)
	fmt.Printf("Available pages: %v\n", pageNav.Pages)
}
```

### Session Management

```go
// Sessions are automatically managed
// Navigation state persists across application restarts

// Reset session to start over
sm.ResetSession()

// Get session information
session := sm.GetCurrentSession()
if session != nil {
	fmt.Printf("Session format: %s\n", session.Format)
	fmt.Printf("Target country: %s\n", session.CountryShort)
	fmt.Printf("Completed: %t\n", session.Completed)
}
```

## 🌍 Supported Countries

Navii includes postal code validation for:
- 🇺🇸 **United States** (US) - 5-digit ZIP codes
- 🇨🇦 **Canada** (CA) - Alphanumeric postal codes
- 🇬🇧 **United Kingdom** (GB) - UK postal code format
- 🇩🇪 **Germany** (DE) - 5-digit postal codes
- 🇯🇵 **Japan** (JP) - 7-digit postal codes with hyphen
- 🇫🇷 **France** (FR) - 5-digit postal codes
- 🇮🇳 **India** (IN) - 6-digit PIN codes
- 🇦🇺 **Australia** (AU) - 4-digit postal codes
- 🇳🇱 **Netherlands** (NL) - 4 digits + 2 letters
- 🇮🇪 **Ireland** (IE) - 3 alphanumeric characters

## 🔧 Advanced Configuration

### Custom Database Path

```go
// Use custom database location
sm, err := navii.NewStateManager("/path/to/custom/navigation.db")
```

### Debug Information

```go
// Enable debug output
sm.Debug()
```

## 📊 Data Sources

Navii fetches geographical data from:
- **Countries & Cities**: [Countries States Cities Database](https://github.com/dr5hn/countries-states-cities-database)
- **Postal Codes**: Various reliable postal code databases

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/ogundaremathew/navii.git
cd navii

# Install dependencies
go mod tidy

# Build the CLI tool
make build

# Download geographical data
make download-data

# Run tests
go test ./...

# Install CLI globally for development
make install
```

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the Navii CLI tool |
| `make install` | Install the CLI tool globally |
| `make download-data` | Build and run data download |
| `make clean` | Clean build artifacts |
| `make help` | Show available commands |

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Thanks to the [Countries States Cities Database](https://github.com/dr5hn/countries-states-cities-database) project for providing comprehensive geographical data
- SQLite for providing a reliable embedded database solution

---

**Made with ❤️ for the Go community**