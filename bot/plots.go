package bot

import (
	"bytes"
	"fmt"
	"github.com/matheusoliveira/go-ordered-map/omap"
	"github.com/wcharczuk/go-chart"
	"regexp"
	"strings"
)

func DrawPieChart(values omap.OMap[string, float64]) (bytes.Buffer, error) {

	var img bytes.Buffer

	var chartValues []chart.Value
	for it := values.Iterator(); it.Next(); {
		if strings.Contains(it.Key(), "Total") {
			continue
		}
		chartValues = append(chartValues, chart.Value{Label: removeEmojis(it.Key()), Value: it.Value()})
	}

	pie := chart.PieChart{
		Width:  1024,
		Height: 1024,
		Values: chartValues,
	}

	err := pie.Render(chart.PNG, &img)
	if err != nil {
		return img, err
	}

	return img, nil
}

func DrawBarChart(values omap.OMap[string, float64]) (bytes.Buffer, error) {

	var img bytes.Buffer

	var chartValues []chart.Value
	for it := values.Iterator(); it.Next(); {
		label := fmt.Sprintf("%s\n$%.0f", removeEmojis(it.Key()), it.Value())
		chartValues = append(chartValues, chart.Value{Label: label, Value: it.Value()})
	}

	bar := chart.BarChart{
		Width:  1024,
		Height: 768,
		Bars:   chartValues,
	}

	bar.XAxis.Show = true
	bar.BarWidth = 60
	bar.TitleStyle.Show = true

	err := bar.Render(chart.PNG, &img)
	if err != nil {
		return img, err
	}

	return img, nil
}

func removeEmojis(str string) string {
	reg := regexp.MustCompile(`[\x{1F600}-\x{1F7BF}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}]+`)
	return reg.ReplaceAllString(str, "")
}
