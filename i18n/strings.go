package i18n

import "fmt"

func FormatNutrition(totalCalories, todayCalories, totalFat, totalCarbs, totalProtein float64, lang string) string {
	return fmt.Sprintf(GetString("format", lang), totalCalories, todayCalories, totalProtein, totalFat, totalCarbs)
}

func GetString(key, lang string) string {
	en := map[string]string{
		"welcome":     "Welcome to the daily calorie count bot! Just enter the name or description of the dish, and its nutritional value will be recorded. You can also specify how much you ate in grams, cups, or plates. The bot can track your daily statistics. Keep track of how much you're eating :)",
		"calories":    "Calories",
		"format":      "CAL: *%.0f* (Total: %.0f)\n*Protein:* %.2fg.\n*Fats:* %.2fg.\n*Carbs:* %.2fg.\n",
		"deleted":     "Deleted: *%s*",
		"added":       "Added: *%s* (%.0fg.)\n%s",
		"total_today": "Total today:\n%s\n%s",
		"morning":     "Morning",
		"afternoon":   "Afternoon",
		"evening":     "Evening",
		"timezone":    "Your current timezone: %s. For accurate daily calorie calculation, specify your location. For example, New York.",
		"unknown":     "I couldn't understand what you wrote.",
		"tz_updated":  "Timezone updated to %s.",
	}

	es := map[string]string{
		"welcome":     "¡Bienvenido al bot de conteo de calorías diarias! Simplemente ingresa el nombre o la descripción del plato, y su valor nutricional se registrará. También puedes especificar cuánto comiste en gramos, tazas o platos. El bot puede hacer un seguimiento de tus estadísticas diarias. ¡Mantente al tanto de cuánto estás comiendo! :)",
		"calories":    "Calorías",
		"format":      "CAL: *%.0f* (Total: %.0f)\n*Proteínas:* %.2fg.\n*Grasas:* %.2fg.\n*Carbohidratos:* %.2fg.\n",
		"deleted":     "Eliminado: *%s*",
		"added":       "Añadido: *%s* (%.0fg.)\n%s",
		"total_today": "Total hoy:\n%s\n%s",
		"morning":     "Mañana",
		"afternoon":   "Tarde",
		"evening":     "Noche",
		"timezone":    "Tu zona horaria actual: %s. Para un cálculo preciso de calorías diarias, especifica tu ubicación. Por ejemplo, Madrid.",
		"unknown":     "No pude entender lo que escribiste.",
		"tz_updated":  "Zona horaria actualizada a %s.",
	}

	ru := map[string]string{
		"welcome":     "Добро пожаловать в бот расчета дневных калорий! Просто введи название или описание блюда, и его пищевая ценность будет записана. Можешь также указать, сколько ты съел в грамах, стаканах, тарелках. Бот умеет считать статистику за сегодня. Следи за тем, сколько ты ешь :)",
		"calories":    "Калории",
		"format":      "ККАЛ: *%.0f* (Всего: %.0f)\n*Белки:* %.2fг.\n*Жиры:* %.2fг.\n*Углеводы:* %.2fг.\n",
		"deleted":     "Удалено: *%s*",
		"added":       "Добавлено: *%s* (%.0fг.)\n%s",
		"total_today": "Всего сегодня:\n%s\n%s",
		"morning":     "Утро",
		"afternoon":   "День",
		"evening":     "Вечер",
		"timezone":    "Ваша текущий часовой пояс: %s. Для точного расчета калорий за день, укажи свою локацию. Например, Ижевск.",
		"unknown":     "Мне не удалось понять, что ты написал.",
		"tz_updated":  "Часовой пояс обновлен на %s.",
	}

	ua := map[string]string{
		"welcome":     "Ласкаво просимо до боту підрахунку денних калорій! Просто введіть назву або опис страви, і її харчова цінність буде записана. Ви також можете вказати, скільки ви з'їли в грамах, чашках або тарілках. Бот може відстежувати вашу щоденну статистику. Слідкуйте за тим, скільки ви їсте :)",
		"calories":    "Калорії",
		"format":      "ККАЛ: *%.0f* (Всього: %.0f)\n*Білки:* %.2fг.\n*Жири:* %.2fг.\n*Вуглеводи:* %.2fг.\n",
		"deleted":     "Видалено: *%s*",
		"added":       "Додано: *%s* (%.0fг.)\n%s",
		"total_today": "Всього сьогодні:\n%s\n%s",
		"morning":     "Ранок",
		"afternoon":   "День",
		"evening":     "Вечір",
		"timezone":    "Ваш часовий пояс: %s. Для точного розрахунку калорій за день, вкажіть свою локацію. Наприклад, Київ.",
		"unknown":     "Мені не вдалося зрозуміти, що ви написали.",
		"tz_updated":  "Часовий пояс оновлено на %s.",
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
