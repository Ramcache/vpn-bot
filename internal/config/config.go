package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	DBDSN            string
	YooKassaShopID   string
	YooKassaSecret   string
	Port             int

	AdminIDs []int64
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		log.Fatalf("Ошибка чтения PORT: %v", err)
	}

	adminIDsStr := getEnv("ADMIN_IDS", "")
	adminIDs := parseAdminIDs(adminIDsStr)

	return &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		DBDSN:            getEnv("DB_DSN", ""),
		YooKassaShopID:   getEnv("YOOKASSA_SHOP_ID", ""),
		YooKassaSecret:   getEnv("YOOKASSA_SECRET_KEY", ""),
		Port:             port,
		AdminIDs:         adminIDs,
	}
}

func parseAdminIDs(s string) []int64 {
	if s == "" {
		return []int64{}
	}
	parts := strings.Split(s, ",")
	var result []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			result = append(result, id)
		}
	}
	return result
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
