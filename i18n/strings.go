package i18n

import "fmt"

func FormatNutrition(totalCalories, todayCalories, totalFat, totalCarbs, totalProtein float64, lang string) string {
	return fmt.Sprintf(GetString("format", lang), totalCalories, todayCalories, totalProtein, totalFat, totalCarbs)
}

func GetString(key, lang string) string {
	en := map[string]string{
		"welcome_food":      "Welcome to the daily calorie count bot! Just enter the name or description of the dish, and its nutritional value will be recorded. You can also specify how much you ate in grams, cups, or plates. The bot can track your daily statistics. Keep track of how much you're eating :)",
		"welcome_expenses":  "Welcome to the budget calculation bot! Just enter your expenses in any format. The bot can calculate statistics for the month. Keep track of your finances :)",
		"unauthorized":      "You are not authorized to use this bot.",
		"calories":          "Calories",
		"format":            "CAL: *%.0f* (Total: %.0f)\n*Protein:* %.2fg.\n*Fats:* %.2fg.\n*Carbs:* %.2fg.\n",
		"deleted":           "Deleted: *%s*",
		"added":             "Added: *%s* (%.0fg.)\n%s",
		"total_today":       "Total today:\n%s\n%s",
		"morning":           "Morning",
		"afternoon":         "Afternoon",
		"evening":           "Evening",
		"timezone":          "Your current timezone: %s. For accurate daily calorie calculation, specify your location. For example, New York.",
		"unknown":           "I couldn't understand what you wrote.",
		"tz_updated":        "Timezone updated to %s.",
		"unknown_command":   "Sorry, I don't know that command.",
		"error_adding":      "Error adding item %s",
		"added_to_category": "Added %s to category %s",
		"error_tz_update":   "Error updating timezone",
		"error_category":    "Error processing category",
		"category_added":    "Category '%s' added",
		"category_deleted":  "Category '%s' deleted",
		"total_year":        "Total spending for the past 12 months: $%.2f",
		"house":             "🏠 House",
		"food":              "🌮 Food",
		"transport":         "🚙 Transport",
		"gasoline":          "⛽️ Gasoline",
		"entertainment":     "🎮 Entertainment",
		"education":         "🎓 Education",
		"health":            "🏥 Health",
		"bills":             "🧾 Bills",
		"other":             "👽 Other",
	}

	es := map[string]string{
		"welcome_food":      "¡Bienvenido al bot de conteo de calorías diarias! Simplemente ingresa el nombre o la descripción del plato, y su valor nutricional se registrará. También puedes especificar cuánto comiste en gramos, tazas o platos. El bot puede hacer un seguimiento de tus estadísticas diarias. ¡Mantente al tanto de cuánto estás comiendo! :)",
		"welcome_expenses":  "¡Bienvenido al bot de cálculo de presupuesto! Simplemente ingresa tus gastos en cualquier formato. El bot puede calcular estadísticas para el mes. ¡Mantente al tanto de tus finanzas! :)",
		"unauthorized":      "No tienes autorización para usar este bot.",
		"calories":          "Calorías",
		"format":            "CAL: *%.0f* (Total: %.0f)\n*Proteínas:* %.2fg.\n*Grasas:* %.2fg.\n*Carbohidratos:* %.2fg.\n",
		"deleted":           "Eliminado: *%s*",
		"added":             "Añadido: *%s* (%.0fg.)\n%s",
		"total_today":       "Total hoy:\n%s\n%s",
		"morning":           "Mañana",
		"afternoon":         "Tarde",
		"evening":           "Noche",
		"timezone":          "Tu zona horaria actual: %s. Para un cálculo preciso de calorías diarias, especifica tu ubicación. Por ejemplo, Madrid.",
		"unknown":           "No pude entender lo que escribiste.",
		"tz_updated":        "Zona horaria actualizada a %s.",
		"unknown_command":   "Lo siento, no conozco ese comando.",
		"error_adding":      "Error al agregar el elemento %s",
		"added_to_category": "Añadido %s a la categoría %s",
		"error_tz_update":   "Error al actualizar la zona horaria",
		"error_category":    "Error al procesar la categoría",
		"category_added":    "Categoría '%s' añadida",
		"category_deleted":  "Categoría '%s' eliminada",
		"total_year":        "Total gastado en los últimos 12 meses: $%.2f",
		"house":             "🏠 Casa",
		"food":              "🌮 Comida",
		"transport":         "🚙 Transporte",
		"gasoline":          "⛽️ Gasolina",
		"entertainment":     "🎮 Entretenimiento",
		"education":         "🎓 Educación",
		"health":            "🏥 Salud",
		"bills":             "🧾 Facturas",
		"other":             "👽 Otro",
	}

	ru := map[string]string{
		"welcome_food":      "Добро пожаловать в бот расчета дневных калорий! Просто введи название или описание блюда, и его пищевая ценность будет записана. Можешь также указать, сколько ты съел в грамах, стаканах, тарелках. Бот умеет считать статистику за сегодня. Следи за тем, сколько ты ешь :)",
		"welcome_expenses":  "Добро пожаловать в бот расчета бюджета! Просто введи свои расходы в любом формате. Бот может считать статистику за месяц. Следи за своими финансами :)",
		"unauthorized":      "У вас нет доступа к этому боту.",
		"calories":          "Калории",
		"format":            "ККАЛ: *%.0f* (Всего: %.0f)\n*Белки:* %.2fг.\n*Жиры:* %.2fг.\n*Углеводы:* %.2fг.\n",
		"deleted":           "Удалено: *%s*",
		"added":             "Добавлено: *%s* (%.0fг.)\n%s",
		"total_today":       "Всего сегодня:\n%s\n%s",
		"morning":           "Утро",
		"afternoon":         "День",
		"evening":           "Вечер",
		"timezone":          "Ваша текущий часовой пояс: %s. Для точного расчета калорий за день, укажи свою локацию. Например, Ижевск.",
		"unknown":           "Мне не удалось понять, что ты написал.",
		"tz_updated":        "Часовой пояс обновлен на %s.",
		"unknown_command":   "Извините, я не знаю этот команду.",
		"error_adding":      "Ошибка добавления элемента %s",
		"added_to_category": "Добавлено %s в категорию %s",
		"error_tz_update":   "Ошибка обновления часового пояса",
		"error_category":    "Ошибка обработки категории",
		"category_added":    "Категория '%s' добавлена",
		"category_deleted":  "Категория '%s' удалена",
		"total_year":        "Общие расходы за последние 12 месяцев: $%.2f",
		"house":             "🏠 Дом",
		"food":              "🌮 Еда",
		"transport":         "🚙 Транспорт",
		"gasoline":          "⛽️ Бензин",
		"entertainment":     "🎮 Развлечения",
		"education":         "🎓 Образование",
		"health":            "🏥 Здоровье",
		"bills":             "🧾 Платежи",
		"other":             "👽 Другое",
	}

	ua := map[string]string{
		"welcome_food":      "Ласкаво просимо до боту підрахунку денних калорій! Просто введіть назву або опис страви, і її харчова цінність буде записана. Ви також можете вказати, скільки ви з'їли в грамах, чашках або тарілках. Бот може відстежувати вашу щоденну статистику. Слідкуйте за тим, скільки ви їсте :)",
		"welcome_expenses":  "Ласкаво просимо до бота підрахунку бюджету! Просто введіть свої витрати в будь-якому форматі. Бот може відстежувати статистику за місяць. Слідкуйте за своїми фінансами :)",
		"unauthorized":      "Ви не маєте доступу до цього бота.",
		"calories":          "Калорії",
		"format":            "ККАЛ: *%.0f* (Всього: %.0f)\n*Білки:* %.2fг.\n*Жири:* %.2fг.\n*Вуглеводи:* %.2fг.\n",
		"deleted":           "Видалено: *%s*",
		"added":             "Додано: *%s* (%.0fг.)\n%s",
		"total_today":       "Всього сьогодні:\n%s\n%s",
		"morning":           "Ранок",
		"afternoon":         "День",
		"evening":           "Вечір",
		"timezone":          "Ваш часовий пояс: %s. Для точного розрахунку калорій за день, вкажіть свою локацію. Наприклад, Київ.",
		"unknown":           "Мені не вдалося зрозуміти, що ви написали.",
		"tz_updated":        "Часовий пояс оновлено на %s.",
		"unknown_command":   "Вибачте, я не знаю цю команду.",
		"error_adding":      "Помилка додавання елемента %s",
		"added_to_category": "Додано %s в категорію %s",
		"error_tz_update":   "Помилка оновлення часового поясу",
		"error_category":    "Помилка обробки категорії",
		"category_added":    "Категорія '%s' додана",
		"category_deleted":  "Категорія '%s' видалена",
		"total_year":        "Загальні витрати за останні 12 місяців: $%.2f",
		"house":             "🏠 Будинок",
		"food":              "🌮 Їжа",
		"transport":         "🚙 Транспорт",
		"gasoline":          "⛽️ Бензин",
		"entertainment":     "🎮 Розваги",
		"education":         "🎓 Освіта",
		"health":            "🏥 Здоров'я",
		"bills":             "🧾 Платежі",
		"other":             "👽 Інше",
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
