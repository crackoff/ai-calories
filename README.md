# AI-Calories

AI-Calories is a Go-based Telegram bot to track and analyze food nutrition data. It uses AI to get nutrition data from food items and stores the data in a MySQL database.

## Features

- Track daily nutrition intake
- Get nutrition data for specific food items
- Delete last food item added
- Get total nutrition intake for the current day

## Setup

1. Clone the repository:

```bash
git clone https://github.com/crackoff/ai-calories.git
```

2. Navigate to the project directory:

```bash
cd ai-calories
```

3. Install the dependencies:

```bash
go mod download
```

4. Set up your environment variables:

```bash
export DATABASE_URL="your_database_url"
export TELEGRAM_BOT_TOKEN="your_telegram_bot_token"
export PPLX_TOKEN="your_perplexity_token"
```

5. Run the application:

```bash
go run main.go
```

## Usage

Once the bot is running, you can interact with it using the following commands:

- Text any food item to the bot to get its nutrition data and add it to the database.
- `/start`: Start the bot and get a welcome message.
- `/help`: Get help on how to use the bot.
- `/total`: Get total nutrition intake for the current day.
- `/dry`: Get nutrition data for a specific food item.
- `/set`: (In development)
- `/delete`: Delete the last food item added.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.