package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const conversionFeeRate = 0.02
const currencyAPIBaseURL = "https://api.freecurrencyapi.com/v1/latest"

var currencyAPIKey string
var currencyHTTPClient = &http.Client{Timeout: 10 * time.Second}

// manualRates stores manually configured rates as "1 USD = X currency"
var manualRates = map[string]float64{}

func InitCurrencyAPI(apiKey string) {
	if apiKey == "" {
		log.Println("CURRENCY_API_KEY not set, currency conversion disabled")
		return
	}
	currencyAPIKey = apiKey
}

// SetManualRate sets a manual exchange rate for a currency (expressed as 1 USD = rate units of currency)
func SetManualRate(currency string, rate float64) {
	manualRates[strings.ToUpper(strings.TrimSpace(currency))] = rate
	log.Printf("Manual rate set: 1 USD = %.2f %s", rate, currency)
}

type latestResponse struct {
	Data map[string]float64 `json:"data"`
}

// ConvertToUSD converts amount from the given currency to USD.
// Adds a 2% conversion fee for non-USD currencies.
func ConvertToUSD(amount float64, currency string) (float64, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "USD" || currency == "" {
		return amount, nil
	}

	// Check manual rates first
	if rate, ok := manualRates[currency]; ok {
		converted := amount / rate
		fee := converted * conversionFeeRate
		return converted + fee, nil
	}

	if currencyAPIKey == "" {
		return 0, fmt.Errorf("currency conversion not available (API key not configured)")
	}

	req, err := http.NewRequest("GET", currencyAPIBaseURL, nil)
	if err != nil {
		return 0, fmt.Errorf("currency conversion failed: %w", err)
	}

	q := req.URL.Query()
	q.Set("apikey", currencyAPIKey)
	q.Set("base_currency", "USD")
	q.Set("currencies", currency)
	req.URL.RawQuery = q.Encode()

	resp, err := currencyHTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("currency API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read currency API response: %w", err)
	}

	var parsed latestResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return 0, fmt.Errorf("currency conversion failed for %s: %w", currency, err)
	}

	rate, ok := parsed.Data[currency]
	if !ok {
		return 0, fmt.Errorf("no conversion rate found for %s to USD", currency)
	}

	converted := amount / rate
	fee := converted * conversionFeeRate
	return converted + fee, nil
}
