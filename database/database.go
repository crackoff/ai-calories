package database

import (
	"ai-calories/i18n"
	"database/sql"
	"fmt"
	"log"
)

func CreateTableIfNotExists(db *sql.DB) error {
	// Create the table if it doesn't exist
	_, err := db.Exec(`create table if not exists foods (
		id                  int auto_increment primary key,
		user_id             bigint                     not null,
		timestamp           bigint                     not null,
		food_item           varchar(255)               not null,
		total_weight        int                        not null,
		calories            int                        not null,
		total_fat           decimal(5, 2) default 0.00 null,
		saturated_fat       decimal(5, 2) default 0.00 null,
		cholesterol         decimal(5, 2) default 0.00 null,
		sodium              decimal(5, 2) default 0.00 null,
		total_carbohydrates decimal(5, 2) default 0.00 null,
		dietary_fiber       decimal(5, 2) default 0.00 null,
		sugars              decimal(5, 2) default 0.00 null,
		protein             decimal(5, 2) default 0.00 null
	);`)
	// Create the index if it doesn't exist
	_, _ = db.Exec(`CREATE INDEX idx_user_timestamp ON foods (user_id, timestamp);`)
	return err
}

func GetTodayNutrition(db *sql.DB, userID int64, lang string) (string, error) {
	// Prepare the query
	query := `SELECT SUM(calories), SUM(total_fat), SUM(total_carbohydrates), SUM(protein) 
              FROM foods 
              WHERE user_id = ? AND DATE(FROM_UNIXTIME(timestamp)) = CURDATE()`

	// Execute the query
	row := db.QueryRow(query, userID)

	// Declare variables to hold the results
	var totalCalories sql.NullInt64
	var totalFat, totalCarbs, totalProtein sql.NullFloat64

	// Scan the result into the variables
	err := row.Scan(&totalCalories, &totalFat, &totalCarbs, &totalProtein)

	// Format the results
	total := i18n.FormatNutrition(int(totalCalories.Int64), totalFat.Float64, totalCarbs.Float64, totalProtein.Float64, lang)
	groups, err := getFoodGroups(db, userID, lang)
	if err != nil {
		log.Print(err)
		return "", err
	}
	s := fmt.Sprintf(i18n.GetString("total_today", lang), total, groups)

	return s, nil
}

func getFoodGroups(db *sql.DB, userID int64, lang string) (string, error) {
	query := fmt.Sprintf(`
		SELECT
			CASE
				WHEN HOUR(FROM_UNIXTIME(timestamp)) BETWEEN 0 AND 11 THEN '%s: '
				WHEN HOUR(FROM_UNIXTIME(timestamp)) BETWEEN 12 AND 17 THEN '%s: '
				ELSE '%s: '
				END as time_of_day,
		GROUP_CONCAT(food_item SEPARATOR ', ') as food_items
		FROM foods
		WHERE user_id = ?
		  AND DATE(FROM_UNIXTIME(timestamp)) = CURDATE()
		GROUP BY time_of_day
		ORDER BY CASE WHEN time_of_day = '%s: ' THEN 1  WHEN time_of_day = '%s: ' THEN 2 ELSE 3 END;`,
		i18n.GetString("morning", lang), i18n.GetString("afternoon", lang), i18n.GetString("evening", lang),
		i18n.GetString("morning", lang), i18n.GetString("afternoon", lang))

	rows, err := db.Query(query, userID)
	if err != nil {
		return "", err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var foodGroups string
	for rows.Next() {
		var timeOfDay, foodItems string
		if err := rows.Scan(&timeOfDay, &foodItems); err != nil {
			return "", err
		}
		foodGroups += fmt.Sprintf("*%s* %s\n", timeOfDay, foodItems)
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	return foodGroups, nil
}

func InsertFood(db *sql.DB, f Food) error {
	query := `INSERT INTO foods(user_id, timestamp, food_item, total_weight, calories, total_fat, saturated_fat, cholesterol, sodium, total_carbohydrates, dietary_fiber, sugars, protein) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, f.UserID, f.Timestamp, f.FoodItem, f.TotalWeight, f.Calories, f.TotalFat, f.SaturatedFat, f.Cholesterol, f.Sodium, f.TotalCarbohydrates, f.DietaryFiber, f.Sugars, f.Protein)
	return err
}

func DeleteLastFood(db *sql.DB, userId int64) (string, error) {
	selectQuery := `SELECT id, food_item FROM foods WHERE user_id = ? ORDER BY timestamp DESC LIMIT 1`
	row := db.QueryRow(selectQuery, userId)

	var id int
	var foodItem string
	err := row.Scan(&id, &foodItem)
	if err != nil {
		return "", err
	}

	deleteQuery := `DELETE FROM foods WHERE id = ?`
	_, err = db.Exec(deleteQuery, id)
	if err != nil {
		return "", err
	}

	return foodItem, nil
}
