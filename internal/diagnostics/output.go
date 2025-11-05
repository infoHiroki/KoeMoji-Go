package diagnostics

import (
	"fmt"
)

var (
	totalChecks  int
	passedChecks int
	warnings     int
	errors       int
)

func printSummary() {
	// Calculate totals
	totalChecks = 0
	passedChecks = 0
	warnings = 0
	errors = 0

	// System checks (always OK if we got here)
	totalChecks++
	passedChecks++

	// Audio checks
	totalChecks++
	if audioOK {
		passedChecks++
	} else {
		errors++
	}

	// Dual recording checks
	totalChecks++
	if dualRecordingOK {
		passedChecks++
	} else if len(dualRecordingWarnings) > 0 {
		warnings++
	} else {
		errors++
	}

	// Config checks
	totalChecks++
	if configOK {
		passedChecks++
	} else if len(configWarnings) > 0 {
		warnings++
	} else {
		errors++
	}

	// Print summary
	if passedChecks > 0 {
		fmt.Printf("✓ %d check(s) passed\n", passedChecks)
	}
	if warnings > 0 {
		fmt.Printf("⚠ %d warning(s)\n", warnings)
	}
	if errors > 0 {
		fmt.Printf("✗ %d error(s)\n", errors)
	}

	fmt.Println()

	// Final message
	if errors == 0 {
		if warnings == 0 {
			fmt.Println("Your system is ready for recording!")
		} else {
			fmt.Println("Your system is mostly ready. Please review the warnings above.")
		}
	} else {
		fmt.Println("Please resolve the errors above before using KoeMoji-Go.")
	}
}
