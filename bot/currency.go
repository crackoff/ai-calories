package bot

import (
	"fmt"
	"log"
	"strings"

	currconv "github.com/kitloong/go-currency-converter-api/v2"
)

const conversionFeeRate = 0.02

var currencyAPI *currconv.API

func InitCurrencyAPI(apiKey string) {
	if apiKey == "" {
		log.Println("CURRENCY_API_KEY not set, currency conversion disabled")
		return
	}
	currencyAPI = currconv.NewAPI(currconv.Config{
		BaseURL: "https://free.currconv.com",
		Version: "v7",
		APIKey:  apiKey,
	})
}

// ConvertToUSD converts amount from the given currency to USD.
// Adds a 2% conversion fee for non-USD currencies.
// Returns the converted amount in USD.
func ConvertToUSD(amount float64, currency string) (float64, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "USD" || currency == "" {
		return amount, nil
	}

	if currencyAPI == nil {
		return 0, fmt.Errorf("currency conversion not available (API key not configured)")
	}

	pair := fmt.Sprintf("%s_USD", currency)
	result, err := currencyAPI.ConvertCompact(currconv.ConvertRequest{
		Q: []string{pair},
	})
	if err != nil {
		return 0, fmt.Errorf("currency conversion failed for %s: %w", currency, err)
	}

	rate, ok := result[pair]
	if !ok {
		return 0, fmt.Errorf("no conversion rate found for %s to USD", currency)
	}

	converted := amount * float64(rate)
	fee := converted * conversionFeeRate
	return converted + fee, nil
}
