package main

import (
	"fmt"
	"strings"
)

func printHeader(title string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 72))
}

func printSubHeader(title string) {
	fmt.Println()
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", 72))
}

func printResult(key, value string) {
	fmt.Printf("%-30s: %s\n", key, value)
}

func printError(err error) {
	fmt.Printf("Error: %v\n", err)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
