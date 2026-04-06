package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: curconv <amount> <currency>\nExample: curconv 200 GBP\n")
		os.Exit(1)
	}

	amount, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid amount: %s\n", os.Args[1])
		os.Exit(1)
	}
	currency := strings.ToUpper(os.Args[2])

	apiKey := os.Getenv("CURRENCY_API_KEY")
	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "CURRENCY_API_KEY env var not set\n")
		os.Exit(1)
	}

	fmt.Printf("Converting %.2f %s to USD...\n", amount, currency)

	// Check for manual rate via <CURRENCY>_RATE env var
	if manualRate := os.Getenv(currency + "_RATE"); manualRate != "" {
		rate, err := strconv.ParseFloat(manualRate, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid %s_RATE value: %s\n", currency, manualRate)
			os.Exit(1)
		}
		converted := amount / rate
		fee := converted * 0.02
		total := converted + fee
		fmt.Printf("Using manual rate: 1 USD = %.2f %s\n", rate, currency)
		fmt.Printf("Converted: %.2f USD\n", converted)
		fmt.Printf("Fee (2%%): %.2f USD\n", fee)
		fmt.Printf("Total: %.2f USD\n", total)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.freecurrencyapi.com/v1/latest", nil)
	q := req.URL.Query()
	q.Set("apikey", apiKey)
	q.Set("base_currency", "USD")
	q.Set("currencies", currency)
	req.URL.RawQuery = q.Encode()

	fmt.Printf("Request URL: %s\n", req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Raw API response: %s\n", string(body))

	var parsed struct {
		Data map[string]float64 `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	rate, ok := parsed.Data[currency]
	if !ok {
		fmt.Fprintf(os.Stderr, "No %s rate in response\n", currency)
		os.Exit(1)
	}

	// rate is: 1 USD = X <currency>, so to convert <currency> to USD: amount / rate
	usdRate := 1.0 / rate
	converted := amount * usdRate
	fee := converted * 0.02
	total := converted + fee

	fmt.Printf("Rate: 1 USD = %.6f %s (i.e. 1 %s = %.6f USD)\n", rate, currency, currency, usdRate)
	fmt.Printf("Converted: %.2f USD\n", converted)
	fmt.Printf("Fee (2%%): %.2f USD\n", fee)
	fmt.Printf("Total: %.2f USD\n", total)
}
