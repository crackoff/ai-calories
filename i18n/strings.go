package i18n

import "fmt"

func FormatNutrition(totalCalories int, totalFat, totalCarbs, totalProtein float64, lang string) string {
	return fmt.Sprintf(GetString("format", lang), totalCalories, totalProtein, totalFat, totalCarbs)
}

func GetString(key, lang string) string {
	en := map[string]string{
		"welcome":     "Welcome to the daily calorie count bot! Just enter the name or description of the dish, and its nutritional value will be recorded. You can also specify how much you ate in grams, cups, or plates. The bot can track your daily statistics. Keep track of how much you're eating :)",
		"calories":    "Calories",
		"format":      "CAL: *%d*\n*Protein:* %.2fg.\n*Fats:* %.2fg.\n*Carbs:* %.2fg.\n",
		"deleted":     "Deleted: *%s*",
		"added":       "Added: *%s* (%dг.)\n%s",
		"total_today": "Total today:\n%s\n%s",
		"morning":     "Morning",
		"afternoon":   "Afternoon",
		"evening":     "Evening",
	}

	es := map[string]string{
		"welcome":     "¡Bienvenido al bot de conteo de calorías diarias! Simplemente ingresa el nombre o la descripción del plato, y su valor nutricional se registrará. También puedes especificar cuánto comiste en gramos, tazas o platos. El bot puede hacer un seguimiento de tus estadísticas diarias. ¡Mantente al tanto de cuánto estás comiendo! :)",
		"calories":    "Calorías",
		"format":      "CAL: *%d*\n*Proteínas:* %.2fg.\n*Grasas:* %.2fg.\n*Carbohidratos:* %.2fg.\n",
		"deleted":     "Eliminado: *%s*",
		"added":       "Añadido: *%s* (%dг.)\n%s",
		"total_today": "Total hoy:\n%s\n%s",
		"morning":     "Mañana",
		"afternoon":   "Tarde",
		"evening":     "Noche",
	}

	ru := map[string]string{
		"welcome":     "Добро пожаловать в бот расчета дневных калорий! Просто введи название или описание блюда, и его пищевая ценность будет записана. Можешь также указать, сколько ты съел в грамах, стаканах, тарелках. Бот умеет считать статистику за сегодня. Следи за тем, сколько ты ешь :)",
		"calories":    "Калории",
		"format":      "ККАЛ: *%d*\n*Белки:* %.2fг.\n*Жиры:* %.2fг.\n*Углеводы:* %.2fг.\n",
		"deleted":     "Удалено: *%s*",
		"added":       "Добавлено: *%s* (%dг.)\n%s",
		"total_today": "Всего сегодня:\n%s\n%s",
		"morning":     "Утро",
		"afternoon":   "День",
		"evening":     "Вечер",
	}

	ua := map[string]string{
		"welcome":     "Ласкаво просимо до боту підрахунку денних калорій! Просто введіть назву або опис страви, і її харчова цінність буде записана. Ви також можете вказати, скільки ви з'їли в грамах, чашках або тарілках. Бот може відстежувати вашу щоденну статистику. Слідкуйте за тим, скільки ви їсте :)",
		"calories":    "Калорії",
		"format":      "ККАЛ: *%d*\n*Білки:* %.2fг.\n*Жири:* %.2fг.\n*Вуглеводи:* %.2fг.\n",
		"deleted":     "Видалено: *%s*",
		"added":       "Додано: *%s* (%dг.)\n%s",
		"total_today": "Всього сьогодні:\n%s\n%s",
		"morning":     "Ранок",
		"afternoon":   "День",
		"evening":     "Вечір",
	}

	switch lang {
	case "en":
		return en[key]
	case "es":
		return es[key]
	case "ru":
		return ru[key]
	case "ua":
		return ua[key]
	default:
		return en[key]
	}
}
