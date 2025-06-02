// Package yuniq provides geographical navigation and state management functionality
package navii

import (
	"fmt"
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
		Format:        NavFormatCityStateCountry,
		TargetCountry: "US",
	})
	if err != nil {
		return err
	}

	// Add some sample data
	err = sm.Populate()
	if err != nil {
		return err
	}

	// Get current navigation
	nav := sm.GetNav()
	if nav != nil {
		fmt.Printf("Current navigation: %+v\n", nav)
	}

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
