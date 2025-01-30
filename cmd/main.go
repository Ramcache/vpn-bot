// cmd/main.go
package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"vpn-bot/internal/config"
	"vpn-bot/internal/repository"
	"vpn-bot/internal/service"
	"vpn-bot/internal/telegram"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.LoadConfig()
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не задан")
	}

	// 2. Инициализация подключения к БД
	db, err := initDB(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	// 3. Инициализируем репозитории
	userRepo := repository.NewUserRepository(db)
	vpnRepo := repository.NewVPNKeyRepository(db)
	payRepo := repository.NewPaymentRepository(db)

	// 4. Инициализируем сервисы
	//    Параметры shopID и secretKey передаются в сервис для запросов к YooKassa
	userService := service.NewUserService(userRepo)
	vpnService := service.NewVPNKeyService(vpnRepo)
	paymentService := service.NewPaymentService(payRepo, vpnService, cfg.YooKassaShopID, cfg.YooKassaSecret)

	// 5. Создаём экземпляр Telegram-бота
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)
	}
	bot.Debug = true

	// 6. Регистрируем обработчики Telegram (Handlers)
	// Допустим, вы вычисляете base64 для shopId:secretKey
	encoded := base64.StdEncoding.EncodeToString([]byte(cfg.YooKassaShopID + ":" + cfg.YooKassaSecret))

	tgHandler := telegram.NewHandler(
		bot,
		userService,
		vpnService,
		paymentService,
		cfg.AdminIDs,
		"Basic "+encoded, // Здесь передаём сразу готовую строку
		[]byte(cfg.YooKassaSecret),
	)

	// 7. Запускаем отдельную горутину для вебхуков ЮKassa
	go func() {
		http.HandleFunc("/yookassa-webhook", tgHandler.HandleYooKassaWebhook)
		addr := ":" + strconv.Itoa(cfg.Port)
		log.Printf("Запуск HTTP-сервера на порту %d для вебхуков ЮKassa...", cfg.Port)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()

	// 8. Запускаем основной цикл обработки (polling) от Telegram
	tgHandler.RunBot()
}

func initDB(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	log.Println("Подключение к PostgreSQL успешно!")
	return pool, nil
}
