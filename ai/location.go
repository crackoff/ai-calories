package ai

func GetTimezone(ai AI, prompt string) (string, error) {
	systemPrompt := `Convert the location from the user's request into an IANA timezone. Return only timezone string without any additional text and formatting. Only a string containing a valid IANA timezone is accepted`
	return ai.QuerySimple(systemPrompt, prompt, 0)
}
