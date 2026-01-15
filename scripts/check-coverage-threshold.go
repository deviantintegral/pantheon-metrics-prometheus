// Package main provides a coverage threshold checking tool.
//
// This script reads Go coverage output and validates that coverage meets
// a minimum threshold. It's designed to work with go test -coverprofile output.
//
// Usage:
//
//	go run scripts/check-coverage-threshold.go [coverage.out] [threshold]
//
// Arguments:
//
//	coverage.out - Path to coverage profile (default: coverage.out)
//	threshold    - Minimum coverage percentage required (default: 31.0)
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultThreshold = 31.0
	epsilon          = 0.05 // Small tolerance for floating point comparison
)

func main() {
	coveragePath := "coverage.out"
	threshold := defaultThreshold

	if len(os.Args) > 1 {
		coveragePath = os.Args[1]
	}

	if len(os.Args) > 2 {
		t, err := strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: Invalid threshold value: %s\n", os.Args[2])
			os.Exit(1)
		}
		threshold = t
	}

	coverage, err := calculateCoverage(coveragePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Coverage: %.1f%%\n", coverage)

	if coverage < threshold-epsilon {
		_, _ = fmt.Fprintf(os.Stderr, "::error::Coverage %.1f%% is below the required threshold of %.1f%%\n",
			coverage, threshold)
		os.Exit(1)
	}

	fmt.Printf("Coverage meets the required threshold of %.1f%%\n", threshold)
}

func calculateCoverage(path string) (float64, error) {
	file, err := os.Open(path) //nolint:gosec // User-specified coverage file path
	if err != nil {
		return 0, fmt.Errorf("failed to open coverage file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var totalStatements int64
	var coveredStatements int64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip the mode line (e.g., "mode: atomic")
		if strings.HasPrefix(line, "mode:") {
			continue
		}

		// Parse coverage line format: name.go:line.column,line.column numStatements count
		// Example: github.com/example/pkg/file.go:10.2,12.3 5 1
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		numStatements, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		count, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		totalStatements += numStatements
		if count > 0 {
			coveredStatements += numStatements
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading coverage file: %w", err)
	}

	if totalStatements == 0 {
		return 0, fmt.Errorf("no statements found in coverage report")
	}

	coverage := (float64(coveredStatements) / float64(totalStatements)) * 100
	return coverage, nil
}
