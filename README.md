```markdown
# VPN Telegram Bot

Этот бот позволяет пользователям покупать, продлевать и проверять статус VPN-ключей через Telegram.

## 🚀 Функциональность
- 📌 **Покупка VPN-ключа**
- 🔄 **Продление ключа**
- 🔑 **Просмотр купленных ключей**
- ✅ **Проверка статуса ключа**
- 💳 **Оплата через YooKassa**

## 📦 Установка
1. Убедитесь, что установлен **Go 1.20+**.
2. Склонируйте репозиторий:
   ```sh
   git clone https://github.com/your-repo/vpn-bot.git
   cd vpn-bot
   ```
3. Установите зависимости:
   ```sh
   go mod tidy
   ```

## ⚙️ Конфигурация
Создайте файл `.env` и укажите:
```ini
ADMIN_IDS=your_admin_idt_tg
BOT_TOKEN=your_telegram_bot_token
DB_URL=your_postgres_connection
YOOKASSA_SHOP_ID=your_shop_id
YOOKASSA_SECRET_KEY=your_secret_key
```

## ▶️ Запуск
```sh
go run cmd/main.go
```

## 📜 API Вебхуков (YooKassa)
Бот обрабатывает вебхуки платежей от YooKassa на порту `8080`.

## 🛠 Технологии
- **Go** (Telegram Bot API, pgx, zap)
- **PostgreSQL**
- **Docker & Kubernetes**
- **YooKassa API**
