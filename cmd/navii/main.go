package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ogundaremathew/navii"
)

func main() {
	var (
		downloadData = flag.Bool("download-data", false, "Download and process geographical data")
		outputPath   = flag.String("output", "location_data.json", "Output path for geographical data")
		help         = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *downloadData {
		fmt.Println("🌍 Navii Data Downloader")
		fmt.Println("========================")
		fmt.Printf("Ensuring geographical data availability at: %s\n\n", *outputPath)

		err := navii.SmartDownloadData("", *outputPath)
		if err != nil {
			log.Fatalf("❌ Failed to ensure data availability: %v", err)
		}

		fmt.Println("✅ Geographical data is ready!")
		fmt.Println("You can now use Navii in your applications.")
		return
	}

	// If no flags provided, show help
	showHelp()
}

func showHelp() {
	fmt.Println("🌍 Navii - Geographical Navigation Package")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  navii [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -download-data    Download and process geographical data")
	fmt.Println("  -output string    Output path for geographical data (default \"location_data.json\")")
	fmt.Println("  -help            Show this help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  navii -download-data")
	fmt.Println("  navii -download-data -output /path/to/data.json")
	fmt.Println("  navii -help")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/ogundaremathew/navii")
}
