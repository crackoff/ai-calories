# AI Calories - Mobile & Web Application Architecture

## 1. Overview

A standalone cross-platform calorie tracking application inspired by the existing AI Calories Telegram bot. The app shares the same MySQL database and AI-powered food recognition approach, but is an independent product with its own backend and UI.

**Scope: Food/calorie tracking only.** Expense tracking is not part of this application.

### Goals
- Single codebase for iOS, Android, and Web
- Shared MySQL database with the existing Telegram bot (food-related tables)
- Independent Go backend (REST API)
- Nutrition cache to minimize AI token usage — known products are served from DB
- Modern, responsive UI with structured food input (name + weight/kcal)

---

## 2. Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Frontend** | React Native + Expo (TypeScript) | Single codebase for iOS/Android/Web, mature ecosystem |
| **Routing** | Expo Router | File-based routing, works across all platforms |
| **State Management** | Zustand + TanStack Query | Lightweight global state + server state caching |
| **Backend** | Go (net/http or Chi router) | Matches existing codebase, high performance |
| **API Protocol** | REST (JSON) | Simple, well-supported by React Native |
| **Database** | MySQL (shared with Telegram bot) | Existing schema, GORM ORM |
| **AI Integration** | OpenAI API (via Go backend) | Nutrition extraction for unknown foods only |
| **Auth** | JWT tokens | Stateless, mobile-friendly |
| **Charts** | Victory Native or react-native-chart-kit | Cross-platform chart rendering |

---

## 3. System Architecture

```
+---------------------------------------------------+
|                    Clients                         |
|  +------------+  +------------+  +------------+   |
|  |   iOS App  |  |Android App |  |  Web App   |   |
|  | (Expo)     |  | (Expo)     |  | (Expo Web) |   |
|  +-----+------+  +-----+------+  +-----+------+  |
|        |               |               |          |
+--------|---------------|---------------|----------+
         |               |               |
         +-------+-------+-------+-------+
                 |                |
                 v                v
+----------------+----------------+------------------+
|              Load Balancer / API Gateway            |
|               (nginx / Traefik)                     |
+---------------------+------------------------------+
                      |
                      v
+---------------------+------------------------------+
|                  Go Backend API                      |
|                                                      |
|  +----------+ +------------+ +---------------+     |
|  |  Auth    | |  Food API  | |   Plan API    |     |
|  |  Handler | |  Handler   | |   Handler     |     |
|  +----+-----+ +-----+------+ +------+--------+    |
|       |              |               |              |
|  +----+-----+ +-----+------+ +------+--------+    |
|  |  Auth    | |   Food     | |    Plan       |    |
|  |  Service | |  Service   | |   Service     |    |
|  +----------+ +-----+------+ +---------------+    |
|                      |                              |
|                +-----+------+                      |
|                | AI Service |                      |
|                | (OpenAI)   |                      |
|                +------------+                      |
|                                                      |
+---------------------+-------------------------------+
                      |
                      v
+---------------------+-------------------------------+
|                    MySQL                              |
|                                                       |
|  +-------+ +--------+ +----------------+             |
|  | users | | foods  | | user_timezones |             |
|  +-------+ +--------+ +----------------+             |
|  +----------------+ +------------+                    |
|  | refresh_tokens | | food_cache |                    |
|  +----------------+ +------------+                    |
|  +------------------+                                  |
|  | payment_history  |                                  |
|  +------------------+                                  |
+-------------------------------------------------------+
```

---

## 4. Database Schema

### Extended Existing Table: `users`

All tables are managed via GORM `db.AutoMigrate()` — the SQL below is **reference only** to illustrate the target schema. GORM auto-migrates from struct definitions (adds missing columns/tables, never drops).

The existing `users` table is extended with new columns for app auth.
Telegram bot continues using `user_id` and `username`; new columns are nullable to stay backward-compatible.

### Existing Tables (shared with Telegram bot)

**users** — extended with new nullable columns:
```sql
-- Existing: id, user_id, username, created_at, updated_at, deleted_at
-- New columns (added by AutoMigrate from updated GORM struct):
email           VARCHAR(255) UNIQUE          -- app login (nullable: Telegram-only users won't have it)
password        VARCHAR(255)                  -- bcrypt hash (nullable for OAuth and Telegram-only users)
auth_provider   VARCHAR(10)                   -- 'email', 'google', 'apple', or NULL (Telegram-only)
language        VARCHAR(5) DEFAULT 'en'       -- preferred language (en/es/ru)
```

- **user_timezones** (user_id, timezone) - per-user timezone
- **foods** (user_id, timestamp, food_item, weight, calories, fat, carbohydrates, protein)

### New Tables

**refresh_tokens** — JWT refresh tokens:
```sql
id          BIGINT PRIMARY KEY AUTO_INCREMENT
user_id     BIGINT NOT NULL               -- FK to users
token       VARCHAR(512) NOT NULL
expires_at  TIMESTAMP NOT NULL
created_at  TIMESTAMP
```

**food_cache** — nutrition per 100g for known foods (populated from AI responses):
```sql
id              BIGINT PRIMARY KEY AUTO_INCREMENT
food_name       VARCHAR(255) NOT NULL UNIQUE   -- normalized lowercase name
calories_100g   FLOAT NOT NULL                 -- kcal per 100g
protein_100g    FLOAT NOT NULL                 -- grams per 100g
fat_100g        FLOAT NOT NULL                 -- grams per 100g
carbs_100g      FLOAT NOT NULL                 -- grams per 100g
image_url       VARCHAR(512)                   -- nullable; product image for future UI
source          VARCHAR(10) DEFAULT 'ai'       -- 'ai' or 'manual'
created_at      TIMESTAMP
updated_at      TIMESTAMP
```

**payment_history** — source of truth for subscription status.
Current plan = latest non-expired row. No rows (or all expired) = free plan:
```sql
id              BIGINT PRIMARY KEY AUTO_INCREMENT
user_id         BIGINT NOT NULL               -- FK to users
sku             VARCHAR(12) NOT NULL           -- e.g. 'pro', 'premium'
payment_date    TIMESTAMP NOT NULL
expiration_date TIMESTAMP NOT NULL
amount          DECIMAL(10,2) NOT NULL
created_at      TIMESTAMP
```

### Food Cache Design

The `food_cache` table stores **nutrition values per 100g** for every food the system has seen. This enables:

1. **Autocomplete** — as the user types, the frontend suggests known foods from the cache
2. **Instant calculation** — when a cached food is selected, nutrition is computed locally (no AI call)
3. **Automatic population** — every AI response for an unknown food is saved to the cache
4. **Cost savings** — repeated queries for common foods (chicken breast, rice, banana) hit the DB, not OpenAI
5. **Product images** — `image_url` is nullable now, but in the future cached foods can display product thumbnails in autocomplete and food logs

**Calculation logic:**
- If user enters **grams**: `calories = (grams / 100) * calories_100g` (same for protein, fat, carbs)
- If user enters **kcal**: `grams = (kcal / calories_100g) * 100`, then derive protein/fat/carbs from grams

---

## 5. Backend API Design

### Base URL: `/api/v1`

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Create account (email + password) |
| POST | `/auth/login` | Login with email + password |
| POST | `/auth/google` | Register or login via Google (ID token from client) |
| POST | `/auth/apple` | Register or login via Apple (ID token from client) |
| POST | `/auth/refresh` | Refresh expired access token |

**OAuth flow (Google / Apple):**

The client handles the native OAuth UI (Expo: `expo-auth-session` / `expo-apple-authentication`). The backend only receives and verifies the resulting ID token — no redirects or server-side OAuth needed.

```
Client (Expo)              Go Backend                Google/Apple
  │                          │                          │
  │  Native sign-in UI       │                          │
  ├─────────────────────────────────────────────────────>│
  │                          │                          │
  │  ID token                │                          │
  │<─────────────────────────────────────────────────────┤
  │                          │                          │
  │  POST /auth/google       │                          │
  │  {id_token: "..."}       │                          │
  ├─────────────────────────>│                          │
  │                          │  Verify token signature  │
  │                          ├─────────────────────────>│
  │                          │  {email, name, sub}      │
  │                          │<─────────────────────────┤
  │                          │                          │
  │                          │  Find or create user     │
  │                          │  by email                │
  │                          │                          │
  │  {access_token,          │                          │
  │   refresh_token}         │                          │
  │<─────────────────────────┤                          │
```

- Backend verifies the ID token against Google/Apple public keys
- If email already exists in `users` → login (return tokens)
- If email is new → create user (password stays NULL for OAuth-only users) → return tokens
- `users` table gains a new column: `auth_provider VARCHAR(10)` — `'email'`, `'google'`, or `'apple'`

### Food Tracking
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/food` | Log food entry (from cache or free-text for AI) |
| GET | `/food/today` | Get today's food entries |
| GET | `/food/date/:date` | Get entries for specific date |
| GET | `/food/summary/today` | Daily nutrition summary (totals + chart data) |
| GET | `/food/summary/:date` | Nutrition summary for specific date |
| GET | `/food/history?period=week` | Calorie history (week/month/year) for charts |
| DELETE | `/food/last` | Delete last food entry |
| DELETE | `/food/:id` | Delete a specific food entry by ID |

### Food Cache (Autocomplete)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/food-cache/search?q=chic` | Search cached foods by prefix (autocomplete) |
| GET | `/food-cache/:id` | Get full nutrition data for a cached food |

### User Settings
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/user/profile` | Get user profile (includes current plan) |
| PUT | `/user/timezone` | Set timezone |
| PUT | `/user/language` | Set preferred language (en/es-419/pt-BR/ru/de/fr) |

### Payments
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/payments/current` | Get current active subscription (latest non-expired payment) |
| POST | `/payments` | Record a payment (sku, amount, expiration) |
| GET | `/payments/history` | Get user's payment history |

### Request/Response Examples

**POST /api/v1/food** (cached food, by grams)
```json
{
  "food_cache_id": 42,
  "input_mode": "grams",
  "value": 200
}
```

**POST /api/v1/food** (cached food, by kcal)
```json
{
  "food_cache_id": 42,
  "input_mode": "kcal",
  "value": 330
}
```

**POST /api/v1/food** (unknown food, by grams — triggers AI)
```json
{
  "free_text": "homemade mushroom risotto",
  "input_mode": "grams",
  "value": 300
}
```

**POST /api/v1/food** (unknown food, by kcal — triggers AI)
```json
{
  "free_text": "homemade mushroom risotto",
  "input_mode": "kcal",
  "value": 330
}
```

**Response (all variants):**
```json
{
  "id": 1,
  "food_item": "Chicken breast",
  "weight": 200,
  "calories": 330,
  "protein": 62.0,
  "fat": 7.2,
  "carbohydrates": 0.0,
  "from_cache": true,
  "timestamp": "2026-04-06T12:30:00Z"
}
```

**GET /api/v1/food-cache/search?q=chic**
```json
[
  { "id": 42, "food_name": "chicken breast", "calories_100g": 165, "image_url": null },
  { "id": 78, "food_name": "chicken thigh", "calories_100g": 209, "image_url": "https://..." },
  { "id": 103, "food_name": "chickpeas", "calories_100g": 164, "image_url": null }
]
```

**GET /api/v1/food/history?period=week**
```json
{
  "period": "week",
  "data": [
    { "date": "2026-03-31", "calories": 2100, "protein": 120.5, "fat": 75.0, "carbs": 210.0 },
    { "date": "2026-04-01", "calories": 1850, "protein": 105.0, "fat": 65.0, "carbs": 180.0 },
    { "date": "2026-04-02", "calories": 2200, "protein": 130.0, "fat": 80.0, "carbs": 230.0 },
    { "date": "2026-04-03", "calories": 1950, "protein": 110.0, "fat": 70.0, "carbs": 195.0 },
    { "date": "2026-04-04", "calories": 0,    "protein": 0.0,   "fat": 0.0,  "carbs": 0.0 },
    { "date": "2026-04-05", "calories": 2050, "protein": 115.0, "fat": 72.0, "carbs": 205.0 },
    { "date": "2026-04-06", "calories": 1850, "protein": 120.5, "fat": 65.0, "carbs": 180.0 }
  ]
}
```

**GET /api/v1/payments/current**
```json
{
  "sku": "pro",
  "payment_date": "2026-03-06T00:00:00Z",
  "expiration_date": "2026-04-06T00:00:00Z",
  "amount": 4.99
}
```
*Returns `null` if no active payment exists (= free plan).*

**GET /api/v1/food/summary/today**
```json
{
  "date": "2026-04-06",
  "total_calories": 1850,
  "total_protein": 120.5,
  "total_fat": 65.0,
  "total_carbohydrates": 180.0,
  "meals": [
    {
      "period": "Morning",
      "entries": [
        {
          "id": 1,
          "food_item": "Oatmeal with banana",
          "calories": 350,
          "protein": 12.0,
          "fat": 8.0,
          "carbohydrates": 58.0,
          "timestamp": "2026-04-06T08:15:00Z"
        }
      ]
    },
    {
      "period": "Afternoon",
      "entries": [...]
    },
    {
      "period": "Evening",
      "entries": [...]
    }
  ],
  "macros_breakdown": {
    "protein_pct": 26.0,
    "fat_pct": 31.6,
    "carbs_pct": 42.4
  }
}
```

---

## 6. Frontend Architecture

### Project Structure (Expo Router)

```
app/
  (auth)/
    login.tsx
    register.tsx
  (tabs)/
    index.tsx          -- Dashboard: today's nutrition overview
    add.tsx            -- Add food entry (text + camera)
    history.tsx        -- Food history by date
    settings.tsx       -- Profile, timezone, language
  _layout.tsx          -- Root layout with auth guard

components/
  FoodCard.tsx           -- Single food entry display
  NutritionSummary.tsx   -- Daily macro summary with totals
  MacroPieChart.tsx      -- Macronutrient pie chart (protein/fat/carbs)
  NutritionHistoryChart.tsx -- Stacked area chart: protein/fat/carbs over time
  PeriodToggle.tsx       -- Segmented control: week / month / year
  MealGroup.tsx          -- Food entries grouped by meal time
  FoodAutocomplete.tsx   -- Typeahead search against food cache
  GramsKcalToggle.tsx    -- Toggle between grams/kcal input mode
  NutritionPreview.tsx   -- Live-calculated nutrition preview before saving
  LanguagePicker.tsx     -- Language selector (6 languages)

hooks/
  useAuth.ts           -- Auth state and token management
  useFood.ts           -- TanStack Query hooks for food API
  useFoodCache.ts      -- TanStack Query hooks for autocomplete/cache search
  useFoodHistory.ts    -- TanStack Query hooks for calorie history (week/month/year)
  usePlan.ts           -- TanStack Query hooks for current plan and payment history
  useSettings.ts       -- User settings hooks

services/
  api.ts               -- Axios/fetch client with JWT interceptor
  storage.ts           -- SecureStore for tokens

stores/
  authStore.ts         -- Zustand auth store
  settingsStore.ts     -- Zustand settings store

i18n/
  en.ts
  es-419.ts
  pt-BR.ts
  ru.ts
  de.ts
  fr.ts
```

### Key UI Screens

1. **Dashboard (main screen)** - Two charts + today's food list (see below)
2. **Add Food** - Structured input form (see below)
3. **History** - Date picker to browse past days, same layout as dashboard for selected date
4. **Settings** - Timezone picker, language selector, current plan info

### Dashboard Screen — Main Screen

```
┌──────────────────────────────────────────┐
│  Today: 1,850 kcal                       │
│                                           │
│  ┌──────────────────────────────────────┐ │
│  │        Macros (pie chart)            │ │
│  │                                      │ │
│  │      Protein 26%  ████               │ │
│  │      Fat     32%  █████              │ │
│  │      Carbs   42%  ███████            │ │
│  │                                      │ │
│  └──────────────────────────────────────┘ │
│                                           │
│  ┌────────┬────────┬────────┐            │
│  │● Week  │ Month  │  Year  │            │
│  └────────┴────────┴────────┘            │
│  ┌──────────────────────────────────────┐ │
│  │  Nutrition History (stacked area)    │ │
│  │                                      │ │
│  │  ▲  ╱‾‾╲        Carbs  ████         │ │
│  │  │ ╱    ╲  ╱╲   Fat    ████         │ │
│  │  │╱  ╱╲  ╲╱  ╲  Protein████        │ │
│  │  │  ╱  ╲      ╲╱╲                   │ │
│  │  │ ╱    ╲╱╲     ╲                   │ │
│  │  │╱        ╲╱────╲──                │ │
│  │  ┗━━━━━━━━━━━━━━━━━━━━━━            │ │
│  │  Mo  Tu  We  Th  Fr  Sa  Su          │ │
│  └──────────────────────────────────────┘ │
│                                           │
│  ── Today's meals ──────────────────────  │
│  Morning                                  │
│    Oatmeal with banana    350 kcal        │
│  Afternoon                                │
│    Chicken breast, rice   480 kcal        │
│    Apple                   95 kcal        │
│  Evening                                  │
│    Salmon with salad      520 kcal        │
└──────────────────────────────────────────┘
```

**Behavior:**

1. **Macros pie chart** — shows today's protein/fat/carbs percentage breakdown. Data from `GET /food/summary/today`.
2. **Nutrition history chart** — stacked area chart with a **week / month / year toggle** (segmented control). Three stacked layers: protein (bottom), fat (middle), carbs (top) — each in a distinct color, with a legend.
   - **Week**: last 7 days, labels = day names
   - **Month**: last 30 days, labels = dates
   - **Year**: last 12 months (aggregated), labels = month names
   - Data from `GET /food/history?period=week|month|year`
3. **Today's meals** — scrollable list grouped by meal time (Morning/Afternoon/Evening)

### Add Food Screen — Input Flow

The Add Food screen has three input fields:

```
┌──────────────────────────────────────────┐
│  Food                                     │
│  ┌──────────────────────────────────────┐ │
│  │ chick...                        [x]  │ │
│  ├──────────────────────────────────────┤ │
│  │ ▸ Chicken breast     (165 kcal/100g) │ │
│  │ ▸ Chicken thigh      (209 kcal/100g) │ │
│  │ ▸ Chickpeas          (164 kcal/100g) │ │
│  └──────────────────────────────────────┘ │
│                                           │
│  Amount                                   │
│  ┌──────────┐  ┌────────┬────────┐       │
│  │ 200      │  │● grams │  kcal  │       │
│  └──────────┘  └────────┴────────┘       │
│                                           │
│  ── Nutrition preview ──────────────────  │
│  Calories: 330 kcal                       │
│  Protein:  62.0g                          │
│  Fat:       7.2g                          │
│  Carbs:     0.0g                          │
│  Weight:  200g                            │
│                                           │
│  [ Save ]                                 │
└──────────────────────────────────────────┘
```

**Behavior:**

1. **Food field** — typeahead autocomplete. As user types, searches `food_cache` via `GET /food-cache/search?q=...`. Debounced (300ms).
   - If user selects a cached item → nutrition preview updates instantly (client-side calculation)
   - If no match / user skips autocomplete → on Save, sends `free_text` to backend → AI classifies → result saved to `food_cache` for next time

2. **Amount field** — numeric input with a **grams / kcal toggle** (segmented control):
   - **grams mode** (default): user enters weight in grams. Calories and macros are calculated: `value = (grams / 100) * per_100g`
   - **kcal mode**: user enters target calories. Grams and macros are back-calculated: `grams = (kcal / calories_100g) * 100`
   - Toggle is only active when a cached food is selected (for free-text, the user describes the amount in the text itself)

3. **Nutrition preview** — live-updated as the user changes the amount. Shows what will be saved. Only visible when a cached food is selected.

---

## 7. Go Backend Structure

```
cmd/
  api/
    main.go            -- API server entry point

internal/
  handler/
    auth.go            -- Auth endpoints
    food.go            -- Food endpoints
    food_cache.go      -- Food cache / autocomplete endpoints
    payment.go         -- Plan listing, payment recording, history endpoints
    user.go            -- User settings endpoints
    middleware.go       -- JWT auth middleware, CORS

  service/
    auth.go            -- Registration, login, token management
    food.go            -- Food business logic (cache-first logging, summaries, history)
    food_cache.go      -- Cache lookup, search, population from AI results
    payment.go         -- Plan resolution from payment history, record payments
    user.go            -- User settings

  repository/
    user.go            -- User DB queries
    food.go            -- Food DB queries
    food_cache.go      -- Food cache DB queries (search, insert, lookup)
    payment.go         -- Plans + payment_history DB queries

  ai/                  -- Inspired by existing ai/ package
    openai.go          -- OpenAI API client
    food_classifier.go -- Nutrition data extraction from text (returns per-100g values)

  model/
    entities.go        -- Domain models
    requests.go        -- API request DTOs
    responses.go       -- API response DTOs

  config/
    config.go          -- Environment-based config

pkg/
  jwt/
    jwt.go             -- JWT token generation and validation
```

### Relationship to Existing Bot Code

This is a separate Go application. It draws inspiration from the existing bot's AI integration patterns (OpenAI food classification, nutrition extraction) but has its own codebase. Shared concepts:
- Same OpenAI prompt strategies for food recognition
- Same database tables (`users`, `foods`, `user_timezones`)
- Same nutrition data model (calories, protein, fat, carbs, weight)
- Same meal-time grouping logic (Morning/Afternoon/Evening)

---

## 8. Authentication Flow

```
┌──────────┐                    ┌──────────┐                ┌──────────┐
│  Client   │                    │ Go API   │                │  MySQL   │
│ (Expo)    │                    │ Backend  │                │          │
└─────┬─────┘                    └─────┬────┘                └─────┬────┘
      │                                │                          │
      │  POST /auth/register           │                          │
      │  {email, password}             │                          │
      ├───────────────────────────────>│                          │
      │                                │  INSERT users (email,    │
      │                                │  password, plan_id=1)    │
      │                                ├─────────────────────────>│
      │                                │                          │
      │  {access_token, refresh_token} │                          │
      │<───────────────────────────────┤                          │
      │                                │                          │
      │  Store tokens (SecureStore)    │                          │
      │                                │                          │
      │  GET /food/today               │                          │
      │  Authorization: Bearer <jwt>   │                          │
      ├───────────────────────────────>│                          │
      │                                │  Verify JWT              │
      │                                │  SELECT foods...         │
      │                                ├─────────────────────────>│
      │                                │                          │
      │  {foods: [...]}                │                          │
      │<───────────────────────────────┤                          │
```

### Token Strategy
- **Access token**: Short-lived (15 min), JWT with user_id claim
- **Refresh token**: Long-lived (30 days), stored in DB, rotated on use
- **Mobile storage**: Expo SecureStore (encrypted keychain/keystore)
- **Web storage**: httpOnly cookie for refresh token, memory for access token

---

## 9. AI Integration Flow (Cache-First)

AI is only called for **unknown foods** — the cache-first approach minimizes token usage:

```
Client                    Go Backend                 food_cache     OpenAI
  │                          │                          │              │
  │  POST /food              │                          │              │
  │  {food_cache_id, grams}  │                          │              │
  ├─────────────────────────>│                          │              │
  │                          │  Lookup cache by ID      │              │
  │                          ├─────────────────────────>│              │
  │                          │  {per-100g values}       │              │
  │                          │<─────────────────────────┤              │
  │                          │  Calculate & save        │              │
  │  {food entry}            │                          │              │
  │<─────────────────────────┤  (no AI call)            │              │
  │                          │                          │              │
  │                          │                          │              │
  │  POST /food              │                          │              │
  │  {free_text}             │                          │              │
  ├─────────────────────────>│                          │              │
  │                          │  Text → OpenAI           │              │
  │                          ├────────────────────────────────────────>│
  │                          │                          │              │
  │                          │  {food, nutrition/100g}  │              │
  │                          │<────────────────────────────────────────┤
  │                          │                          │              │
  │                          │  Save to food_cache      │              │
  │                          ├─────────────────────────>│              │
  │                          │  Save to foods           │              │
  │  {food entry}            │                          │              │
  │<─────────────────────────┤                          │              │
```

**Two paths:**

1. **Cached food** (most common) — user selects from autocomplete. Backend looks up `food_cache`, calculates nutrition from grams/kcal, saves to `foods`. Zero AI cost.

2. **Unknown food** (free-text) — user types something not in cache. Backend sends to OpenAI, gets nutrition per 100g, saves result to `food_cache` (so it's cached for next time), calculates actual values, saves to `foods`.

Over time, the cache grows and AI calls decrease significantly.

---

## 10. Internationalization

Frontend uses `i18next` + `react-i18next` with the following supported languages:

| Code | Language |
|------|----------|
| `en` | English (US) |
| `es-419` | Spanish (Latin America) |
| `pt-BR` | Portuguese (Brazil) |
| `ru` | Russian |
| `de` | German |
| `fr` | French |

Language is stored per-user in the backend and synced to the client on login.

---

## 11. Deployment

```
┌──────────────────────────────────────────────┐
│                 Production                    │
│                                               │
│  ┌─────────────┐    ┌─────────────────────┐  │
│  │  Expo Web   │    │   Mobile Stores     │  │
│  │  (Vercel /  │    │   (App Store /      │  │
│  │   Netlify)  │    │    Google Play)     │  │
│  └──────┬──────┘    └──────────┬──────────┘  │
│         │                      │              │
│         └──────────┬───────────┘              │
│                    │                          │
│         ┌──────────▼──────────┐               │
│         │   Go API Server     │               │
│         │  (Docker / VPS)     │               │
│         └──────────┬──────────┘               │
│                    │                          │
│         ┌──────────▼──────────┐               │
│         │      MySQL          │               │
│         │  (shared database)  │               │
│         └─────────────────────┘               │
│                                               │
│  ┌─────────────────────────────────────────┐  │
│  │  Telegram Bot (independent, same DB)    │  │
│  └─────────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

- **Web**: Deploy Expo Web build to Vercel/Netlify (static)
- **Mobile**: EAS Build + Submit to App Store / Google Play
- **Backend**: Docker container on same VPS as MySQL
- **Telegram bot**: Independent application, shares the same database

---

## 12. MVP Scope

### Phase 1 - Core
- [ ] Go REST API with auth (register/login/JWT) on extended `users` table
- [ ] Plans + payment history: free by default, current plan derived from latest active payment
- [ ] Food cache table + autocomplete search endpoint
- [ ] Add Food screen: autocomplete + grams/kcal toggle + nutrition preview
- [ ] Cache-first food logging (cached → instant, unknown → AI → cache)
- [ ] Dashboard: macros pie chart + calorie history bar chart (week/month/year toggle)
- [ ] Daily food summary with nutrition totals
- [ ] Meal-time grouping (Morning/Afternoon/Evening)
- [ ] Basic UI for all screens
- [ ] i18n (EN)

### Phase 2 - Full Features
- [ ] History view with date picker
- [ ] Delete last entry
- [ ] Timezone management
- [ ] Multi-language support (es-419, pt-BR, ru, de, fr)
- [ ] Plan upgrade/downgrade UI in settings

### Phase 3 - Enhancements
- [ ] Camera input for food photos (AI vision)
- [ ] Food cache product images
- [ ] Push notifications (daily summary reminders)
- [ ] Offline mode with sync
- [ ] Dark mode
- [ ] Export data (CSV)

---

## 13. Key Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| API style | REST | Simple, great React Native support, sufficient for CRUD-heavy app |
| Auth | JWT + refresh tokens | Stateless, mobile-friendly, industry standard |
| Users table | Extended existing (not separate) | Backward-compatible with Telegram bot, single source of truth |
| Subscriptions | Payment history with SKU + expiration | No plans table; active = latest non-expired payment; no payment = free |
| AI strategy | Cache-first, AI as fallback | Minimizes OpenAI token costs; cache grows over time |
| Nutrition storage | Per-100g in food_cache | Enables flexible calculation from either grams or kcal |
| Charts | Victory Native | Cross-platform, good web support via Expo |
| Offline | TanStack Query cache + optimistic updates | Built-in, minimal extra code |
| Shared DB | Direct sharing between bot and app | Both apps read/write same food data |
| Separate backend | Independent Go API (not embedded in bot) | Clean separation, independent deployment |
| Photo recognition | Deferred to Phase 3 | Text input + cache covers most use cases first |
