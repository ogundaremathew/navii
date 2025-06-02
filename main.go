// Package yuniq provides geographical navigation and state management functionality
package navii

import (
	"fmt"
	"log"
)

// Example usage function
func ExampleUsage() error {
	// Create state manager
	sm, err := NewStateManager("example.db")
	if err != nil {
		return err
	}
	defer sm.Close()

	// Initialize with options
	err = sm.Init(InitOptions{
		Format:        NavFormatCityState,
		TargetCountry: "US",
	})
	if err != nil {
		return err
	}

	// Get current navigation
	nav := sm.GetNav()
	if nav != nil {
		fmt.Printf("Current navigation: %+v\n", nav)
	}

	fmt.Println("Navigating to next level...")
	log.Println(nav)
	// Get next navigation
	nextNav, err := sm.GetNextNav()
	if err != nil {
		return err
	}

	if nextNav != nil {
		fmt.Printf("Next navigation: %+v\n", nextNav)
	}

	// Debug information
	sm.Debug()

	return nil
}

func main() {
	ExampleUsage()
}
