package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"vpn-bot/internal/service"
)

type Handler struct {
	bot                *tgbotapi.BotAPI
	userService        service.UserService
	vpnKeyService      service.VPNKeyService
	paymentService     service.PaymentService
	adminIDs           []int64
	expectedAuthHeader string
	secretKey          []byte
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	userService service.UserService,
	vpnKeyService service.VPNKeyService,
	paymentService service.PaymentService,
	adminIDs []int64,
	expectedAuthHeader string,
	secretKey []byte,
) *Handler {
	return &Handler{
		bot:                bot,
		userService:        userService,
		vpnKeyService:      vpnKeyService,
		paymentService:     paymentService,
		adminIDs:           adminIDs,
		expectedAuthHeader: expectedAuthHeader,
		secretKey:          secretKey,
	}
}

func (h *Handler) IsAdmin(telegramID int64) bool {
	for _, id := range h.adminIDs {
		if id == telegramID {
			return true
		}
	}
	return false
}

// internal/telegram/bot.go

func (h *Handler) RunBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			h.handleMessage(update)
		} else if update.CallbackQuery != nil {
			h.handleCallbackQuery(update)
		}
	}
}
