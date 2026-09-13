// Command covergate enforces a minimum statement coverage for ./pkg/... .
//
// Usage: go run ./tools/covergate -profile coverage.out -threshold 90
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	profile := flag.String("profile", "coverage.out", "coverage profile produced by go test")
	threshold := flag.Float64("threshold", 90, "minimum coverage percent for ./pkg/...")
	prefix := flag.String("prefix", "template-golang/pkg/", "only files under this import prefix are gated")
	excludeFlag := flag.String("exclude", "", "comma-separated path fragments to skip (infra wrappers covered by integration tests)")
	flag.Parse()

	excludes := make([]string, 0)
	for _, raw := range strings.Split(*excludeFlag, ",") {
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			excludes = append(excludes, trimmed)
		}
	}

	file, err := os.Open(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open profile:", err)
		os.Exit(1)
	}
	defer func() { _ = file.Close() }()

	var covered, total int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		if !strings.HasPrefix(fields[0], *prefix) {
			continue
		}
		excluded := false
		for _, fragment := range excludes {
			if strings.Contains(fields[0], fragment) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			continue
		}
		total += statements
		if count > 0 {
			covered += statements
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "read profile:", err)
		os.Exit(1)
	}
	if total == 0 {
		fmt.Fprintln(os.Stderr, "no statements found for", *prefix)
		os.Exit(1)
	}

	percent := float64(covered) / float64(total) * 100
	fmt.Printf("pkg coverage: %.2f%% (threshold %.0f%%)\n", percent, *threshold)
	if percent+0.0001 < *threshold {
		fmt.Fprintln(os.Stderr, "coverage gate failed")
		os.Exit(1)
	}
}
