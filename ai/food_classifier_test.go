package ai

import (
	data "ai-calories/database"
	"bytes"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
)

type queryCall struct {
	system      string
	user        string
	temperature int
}

type recognizeCall struct {
	system string
	user   string
}

type fakeAI struct {
	queryResponses     []string
	queryErr           error
	queryCalls         []queryCall
	recognizeResponse  string
	recognizeErr       error
	recognizeCallCount int
	recognizeCalls     []recognizeCall
}

func (f *fakeAI) QuerySimple(system string, user string, temperature int) (string, error) {
	f.queryCalls = append(f.queryCalls, queryCall{
		system:      system,
		user:        user,
		temperature: temperature,
	})
	if f.queryErr != nil {
		return "", f.queryErr
	}
	if len(f.queryResponses) == 0 {
		return "", errors.New("no fake query response configured")
	}
	resp := f.queryResponses[0]
	f.queryResponses = f.queryResponses[1:]
	return resp, nil
}

func (f *fakeAI) RecognizeImage(img bytes.Buffer, system string, user string) (string, error) {
	f.recognizeCallCount++
	f.recognizeCalls = append(f.recognizeCalls, recognizeCall{
		system: system,
		user:   user,
	})
	if f.recognizeErr != nil {
		return "", f.recognizeErr
	}
	return f.recognizeResponse, nil
}

func requireFloatEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %.8f, want %.8f", got, want)
	}
}

func setTwoStepsEnv(t *testing.T, value string, set bool) {
	t.Helper()
	oldVal, hadOld := os.LookupEnv("TWO_STEPS_PROMPT")
	if set {
		if err := os.Setenv("TWO_STEPS_PROMPT", value); err != nil {
			t.Fatalf("failed to set env: %v", err)
		}
	} else {
		if err := os.Unsetenv("TWO_STEPS_PROMPT"); err != nil {
			t.Fatalf("failed to unset env: %v", err)
		}
	}
	t.Cleanup(func() {
		var err error
		if hadOld {
			err = os.Setenv("TWO_STEPS_PROMPT", oldVal)
		} else {
			err = os.Unsetenv("TWO_STEPS_PROMPT")
		}
		if err != nil {
			t.Fatalf("failed restoring env: %v", err)
		}
	})
}

func TestParseLabelResponse_ComputesTotalsFromServingAndPackageWeight(t *testing.T) {
	classifier := FoodClassifier{}
	food, err := classifier.ParseLabelResponse(
		"LABEL: Greek Yogurt|serving:150|calories:95|protein:15|fat:0.5|carbs:8|total_weight:450",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if food.FoodItem != "Greek Yogurt" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
	requireFloatEqual(t, food.Weight, 450)
	requireFloatEqual(t, food.Calories, 285)
	requireFloatEqual(t, food.Protein, 45)
	requireFloatEqual(t, food.Fat, 1.5)
	requireFloatEqual(t, food.Carbohydrates, 24)
}

func TestParseLabelResponse_ScalesToUserAmountFromDescription(t *testing.T) {
	classifier := FoodClassifier{}
	// Label base: per 100ml; user asked for 300 ml via caption.
	food, err := classifier.ParseLabelResponse(
		"LABEL: orange juice|serving:100|calories:41|protein:0.7|fat:0|carbs:9.5|total_weight:200",
		"300 мл",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if food.FoodItem != "orange juice" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
	requireFloatEqual(t, food.Weight, 300)
	requireFloatEqual(t, food.Calories, 123)
	requireFloatEqual(t, food.Protein, 2.1)
	requireFloatEqual(t, food.Fat, 0)
	requireFloatEqual(t, food.Carbohydrates, 28.5)
}

func TestParseLabelResponse_ReplacesQuestionMarkName(t *testing.T) {
	classifier := FoodClassifier{}
	food, err := classifier.ParseLabelResponse(
		"LABEL: ?|serving:100|calories:41|protein:0.7|fat:0|carbs:9.5|total_weight:100",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if food.FoodItem != "food" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
}

func TestParseUserAmountGrams(t *testing.T) {
	cases := []struct {
		in     string
		want   float64
		wantOK bool
	}{
		{"300 мл", 300, true},
		{"300 ml", 300, true},
		{"220g", 220, true},
		{"1.5 kg", 1500, true},
		{"no amount here", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseUserAmountGrams(tc.in)
		if ok != tc.wantOK {
			t.Fatalf("%q: ok=%v, want %v", tc.in, ok, tc.wantOK)
		}
		if ok {
			requireFloatEqual(t, got, tc.want)
		}
	}
}

func TestParseLabelResponse_DefaultsTotalWeightToServingWhenMissing(t *testing.T) {
	classifier := FoodClassifier{}
	food, err := classifier.ParseLabelResponse(
		"LABEL: Protein Bar|serving:60|calories:210|protein:20|fat:7|carbs:22",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if food.FoodItem != "Protein Bar" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
	requireFloatEqual(t, food.Weight, 60)
	requireFloatEqual(t, food.Calories, 210)
	requireFloatEqual(t, food.Protein, 20)
	requireFloatEqual(t, food.Fat, 7)
	requireFloatEqual(t, food.Carbohydrates, 22)
}

func TestParseLabelResponse_IgnoresUnknownOrQuestionMarkValues(t *testing.T) {
	classifier := FoodClassifier{}
	food, err := classifier.ParseLabelResponse(
		"LABEL: Mystery Product|serving:?|calories:100|protein:?|fat:2|carbs:10|total_weight:200|unknown:abc",
		"",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if food.FoodItem != "Mystery Product" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
	requireFloatEqual(t, food.Weight, 200)
	requireFloatEqual(t, food.Calories, 200)
	requireFloatEqual(t, food.Protein, 0)
	requireFloatEqual(t, food.Fat, 4)
	requireFloatEqual(t, food.Carbohydrates, 20)
}

func TestGetGetNutritionDataByImage_LabelResponseBypassesTextNutritionFlow(t *testing.T) {
	setTwoStepsEnv(t, "", false)
	fake := &fakeAI{
		recognizeResponse: "LABEL: Greek Yogurt|serving:150|calories:95|protein:15|fat:0.5|carbs:8|total_weight:450",
	}
	classifier := FoodClassifier{ai: fake}
	var img bytes.Buffer

	food, err := classifier.GetGetNutritionDataByImage(&img, "plain greek yogurt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fake.recognizeCallCount != 1 {
		t.Fatalf("recognize call count: got %d, want 1", fake.recognizeCallCount)
	}
	if len(fake.queryCalls) != 0 {
		t.Fatalf("query call count: got %d, want 0", len(fake.queryCalls))
	}
	if !strings.Contains(fake.recognizeCalls[0].user, "plain greek yogurt") {
		t.Fatalf("recognize user prompt missing caption: %q", fake.recognizeCalls[0].user)
	}
	requireFloatEqual(t, food.Weight, 450)
	requireFloatEqual(t, food.Calories, 285)
}

func TestGetGetNutritionDataByImage_LabelScalesCaptionAmount(t *testing.T) {
	setTwoStepsEnv(t, "", false)
	fake := &fakeAI{
		recognizeResponse: "LABEL: juice|serving:100|calories:41|protein:0.7|fat:0|carbs:9.5|total_weight:100",
	}
	classifier := FoodClassifier{ai: fake}
	var img bytes.Buffer

	food, err := classifier.GetGetNutritionDataByImage(&img, "300 мл")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	requireFloatEqual(t, food.Weight, 300)
	requireFloatEqual(t, food.Calories, 123)
}

func TestGetGetNutritionDataByImage_FoodResponseAppendsDescription(t *testing.T) {
	setTwoStepsEnv(t, "", false)
	fake := &fakeAI{
		recognizeResponse: "FOOD: Caesar salad with chicken, 280 grams",
		queryResponses: []string{
			`{"food_item":"Caesar salad with chicken","weight":280,"calories":500,"fat":20,"carbohydrates":30,"protein":40}`,
		},
	}
	classifier := FoodClassifier{ai: fake}
	var img bytes.Buffer

	food, err := classifier.GetGetNutritionDataByImage(&img, "with dressing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.queryCalls) != 1 {
		t.Fatalf("query call count: got %d, want 1", len(fake.queryCalls))
	}
	gotPrompt := fake.queryCalls[0].user
	wantPrompt := "Caesar salad with chicken, 280 grams\nwith dressing"
	if gotPrompt != wantPrompt {
		t.Fatalf("query user prompt mismatch:\n got: %q\nwant: %q", gotPrompt, wantPrompt)
	}
	if food.FoodItem != "Caesar salad with chicken" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
}

func TestGetNutritionData_DefaultsToOneStepWhenEnvNotSet(t *testing.T) {
	setTwoStepsEnv(t, "", false)
	fake := &fakeAI{
		queryResponses: []string{
			`{"food_item":"banana","weight":100,"calories":105,"fat":0.3,"carbohydrates":27,"protein":1.3}`,
		},
	}
	classifier := FoodClassifier{ai: fake}

	food, err := classifier.GetNutritionData("banana")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.queryCalls) != 1 {
		t.Fatalf("query call count: got %d, want 1", len(fake.queryCalls))
	}
	if fake.queryCalls[0].user != "banana" {
		t.Fatalf("unexpected query user: %q", fake.queryCalls[0].user)
	}
	if food.FoodItem != "banana" {
		t.Fatalf("unexpected food item: %q", food.FoodItem)
	}
}

func TestGetNutritionData_UsesTwoStepsWhenEnvSet(t *testing.T) {
	setTwoStepsEnv(t, "1", true)
	fake := &fakeAI{
		queryResponses: []string{
			"step-by-step reasoning",
			`{"food_item":"banana","weight":100,"calories":105,"fat":0.3,"carbohydrates":27,"protein":1.3}`,
		},
	}
	classifier := FoodClassifier{ai: fake}

	food, err := classifier.GetNutritionData("banana")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.queryCalls) != 2 {
		t.Fatalf("query call count: got %d, want 2", len(fake.queryCalls))
	}
	if fake.queryCalls[0].user != "banana" {
		t.Fatalf("first query user mismatch: %q", fake.queryCalls[0].user)
	}
	if fake.queryCalls[1].user != "step-by-step reasoning" {
		t.Fatalf("second query user mismatch: %q", fake.queryCalls[1].user)
	}
	if food != (data.Food{
		FoodItem:      "banana",
		Weight:        100,
		Calories:      105,
		Fat:           0.3,
		Carbohydrates: 27,
		Protein:       1.3,
	}) {
		t.Fatalf("unexpected food payload: %+v", food)
	}
}
