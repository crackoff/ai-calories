package ai

import (
	data "ai-calories/database"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func GetNutritionData(description string) (data.Food, error) {
	pplxToken := os.Getenv("PPLX_TOKEN")
	systemQuery := `<CONTEXT>\nYou are nutrition facts expert and you will help user to determine a number of calories(kcal), and other nutritional content by a food name or  description.\n</CONTEXT>\n\n<OBJECTIVES>\nWhen a user types the name of a food, such as \"McDonald's cheeseburger,\" you need to determine its nutritional content using common knowledge. If exact data is unavailable, estimate using average values. Provide the answer with precise figures, using your best judgment if necessary.\n\nUser requests may include weight specifications. If the weight is specified, use it; if not, estimate using a generally accepted average value, expressed in exact grams. For instance, a user might request \"Caesar salad, 200g\". If the weight is provided in pounds or any other unit, convert it to grams. Requests may also use terms like \"portion,\" \"plate,\" or \"cup.\" In such cases, apply reasonable gram equivalents. Approximate values are acceptable in this context.\n\nUser requests may include a number of items. In this case, multiply values. For example, user can ask \"2 empanadas with meat, 120 grams each\".\n\nNow you have the total weight.\n\nFinally, calculate nutrition facts for the provided food multiplied by the total weight. We need accuracy here so think step-by-step, check your answer for every calculation. \nWhat needs to be calculated: \n1. Total fats in grams per total weight.\n2. Saturated fats in grams per total weight.\n3. Cholesterol in milligrams per total weight.\n4. Sodium in milligrams per total weight.\n5. Total carbohydrates in grams per total weight.\n6. Dietary fiber in grams per total weight.\n7. Sugars in grams per total weight.\n8. Proteins in grams per total weight.\n</OBJECTIVES>\n\n<RESPONSE> \nThis query will be used in an API call so provide an output in a json format as per the following:\nCalories should be in integer value, cholesterol and sodium in integer in mg, all others - are in float pointing values in grams.\nThe answer should contain only JSON, no other text is acceptable. Markdown is also not acceptable.\n\n{\n\t\"food_item\": \"{name of the food}\",\n\t\"total_weight\": {weight in grams},\n\t\"total_fat\": {total fat in grams},\n    \"saturated_fat\": {saturated fat in grams},\n    \"cholesterol\": {cholesterol in mg},\n    \"sodium\": {sodium in mg},\n    \"total_carbohydrates\": {total carbohydrates in grams},\n    \"dietary_fiber\": {dietary fiber in grams},\n    \"sugars\": {sugars in grams},\n    \"protein\": \"{proteing in grams}\"\n}\n\nFor example, user requests \"banana\". Your answer should be:\n\n{\n\t\"food_item\": \"banana\",\n\t\"total_weight\": 100,\n    \"calories\": 105,\n    \"total_fat\": 0.3,\n    \"saturated_fat\": 0.1,\n    \"cholesterol\": 0,\n    \"sodium\": 1,\n    \"total_carbohydrates\": 27,\n    \"dietary_fiber\": 3.1,\n    \"sugars\": 14.4,\n    \"protein\": 1.3\n}\n\n </RESPONSE>\n\n`

	client := &http.Client{}
	jsonData := fmt.Sprintf(`{"model": "llama-3-70b-instruct","messages": [{"role": "system","content": "%s"},{"role": "user","content": "%s"}],	"temperature": 0}`, systemQuery, description)

	request, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return data.Food{}, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pplxToken))

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return data.Food{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return data.Food{}, err
	}

	log.Println(string(body))

	var completion Completion
	err = json.Unmarshal(body, &completion)
	if err != nil {
		fmt.Println(err)
		return data.Food{}, err
	}

	var food data.Food
	err = json.Unmarshal([]byte(completion.Choices[0].Message.Content), &food)
	if err != nil {
		fmt.Println(err)
		return data.Food{}, err
	}

	return food, nil
}

type Completion struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Created int    `json:"created"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Object  string `json:"object"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
