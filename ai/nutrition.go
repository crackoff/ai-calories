package ai

import (
	data "ai-calories/database"
	"encoding/json"
	"fmt"
	"log"
)

func GetNutritionData(description string) (data.Food, error) {
	// Execute the first query to step-by-step calculations with thinking out loud
	systemQuery1 := `<CONTEXT>\nYou are nutrition facts expert and you will help user to determine a number of calories (kcal), and other nutritional content by a food name or  description.\n</CONTEXT>\n\n<OBJECTIVES>\nWhen a user types the name of a food, such as \"McDonald's cheeseburger,\" you need to determine its nutritional content using common knowledge. If exact data is unavailable, estimate using average values. Provide the answer with precise figures, using your best judgment if necessary.\n\nUser requests may include weight specifications. If the weight is specified, use it; if not, estimate using a generally accepted average value, expressed in exact grams. For instance, a user might request \"Caesar salad, 200g\". If the weight is provided in pounds or any other unit, convert it to grams. Requests may also use terms like \"portion,\" \"plate,\" or \"cup.\" In such cases, apply reasonable gram equivalents. Approximate values are acceptable in this context.\n\nUser requests may include a number of items. In this case, multiply values. For example, user can ask \"2 empanadas with meat, 120 grams each\".\n\nNow you have the total weight.\n\nFinally, calculate nutrition facts for the provided food multiplied by the total weight. We need accuracy here so think step-by-step, check your answer for every calculation. \nWhat needs to be calculated: \n1. Total fats in grams per total weight.\n2. Saturated fats in grams per total weight.\n3. Cholesterol in milligrams per total weight.\n4. Sodium in milligrams per total weight.\n5. Total carbohydrates in grams per total weight.\n6. Dietary fiber in grams per total weight.\n7. Sugars in grams per total weight.\n8. Proteins in grams per total weight.\n</OBJECTIVES>\n\n<RESPONSE>\nWe need a high accuracy here. Think step-by-step and explain all your calculations.\nIf you know nutrition data per 100g, re-calculate it for the total product weight.\nIf you know nutrition data separately, calculate it separately for all products. For example, if user says \"a cup black coffee with milk\", most probably it means 210 grams of black coffee and 30 grams of milk. Calculate separately and then summarize.\nRepeat calculations as many times as you need. Show all steps.\nDon't afraid to be wrong firs time and then correct your mistakes. Better to get a very accurate result.\n</RESPONSE>`
	result1, err := ExecutePplxQuerySimple(systemQuery1, description, 0)
	if err != nil {
		log.Println(err)
		return data.Food{}, err
	}

	// Execute the second query to format into JSON
	systemQuery2 := `<CONTEXT>\nYou are the API resolver and you will generate a JSON API call output.\n</CONTEXT>\n\n<OBJECTIVES>\nUser will provide a calculations of nutrition facts. Your goal is to get a final result of these calculations and generate a structured output which will be used for further calculations.\n</OBJECTIVES>\n\n<RESPONSE>\nProvide the final results from user input in a json format as per the following structure.\nCalories should be in integer value, cholesterol and sodium  are integers in mg, all others - are float pointing values in grams.\nThe answer should contain only JSON, no other text is acceptable. Markdown is also not acceptable.\n\n{\n    \"food_item\": \"{name of the food}\",\n    \"total_weight\": {weight in grams},\n    \"total_fat\": {total fat in grams},\n    \"saturated_fat\": {saturated fat in grams},\n    \"cholesterol\": {cholesterol in mg},\n    \"sodium\": {sodium in mg},\n    \"total_carbohydrates\": {total carbohydrates in grams},\n    \"dietary_fiber\": {dietary fiber in grams},\n    \"sugars\": {sugars in grams},\n    \"protein\": {protein in grams}\n}\n\nFor example, user requests \"banana\". Your answer should be:\n\n{\n    \"food_item\": \"banana\",\n    \"total_weight\": 100,\n    \"calories\": 105,\n    \"total_fat\": 0.3,\n    \"saturated_fat\": 0.1,\n    \"cholesterol\": 0,\n    \"sodium\": 1,\n    \"total_carbohydrates\": 27,\n    \"dietary_fiber\": 3.1,\n    \"sugars\": 14.4,\n    \"protein\": 1.3\n}\n</RESPONSE>`
	result2, err := ExecutePplxQuerySimple(systemQuery2, result1, 0)
	if err != nil {
		log.Println(err)
		return data.Food{}, err
	}

	// Deserialize JSON to an object
	var food data.Food
	err = json.Unmarshal([]byte(result2), &food)
	if err != nil {
		fmt.Println(err)
		return data.Food{}, err
	}

	return food, nil
}
