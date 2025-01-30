package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleMessage(update tgbotapi.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	text := msg.Text

	if msg == nil {
		return
	}

	log.Printf("Получено сообщение: %s от %d", text, msg.From.ID) // Добавили лог

	if h.IsAdmin(msg.From.ID) {
		log.Println("Пользователь является администратором") // Лог для проверки
		if strings.HasPrefix(text, "/add_key ") {
			log.Println("Обнаружена команда /add_key") // Лог перед вызовом
			h.handleAddKeyCommand(chatID, text)
			return
		}
	}

	switch text {
	case "/start", "Купить VPN", "Мои ключи", "Продлить ключ", "Статус ключа":
		h.handleUserCommand(chatID, text, int(msg.From.ID), msg.From.UserName)

	default:
		log.Println("Команда не распознана, отправляем стандартный ответ")
		h.sendMessageText(chatID, "❓ Неизвестная команда. Выберите действие:")
	}
}

func (h *Handler) handleUserCommand(chatID int64, text string, userID int, username string) {
	switch text {
	case "/start":
		err := h.userService.RegisterUser(
			int64(userID),
			username,
			fmt.Sprintf("t.me/%s", username),
		)
		if err != nil {
			log.Printf("Ошибка регистрации пользователя %d: %v", userID, err)
		}

		h.sendMessageText(chatID, "Добро пожаловать! Выберите действие:")
		h.sendMenuKeyboard(chatID)

	case "Купить VPN":
		h.processBuyVPN(chatID, userID)

	case "Мои ключи":
		h.processMyKeys(chatID, userID)

	case "Продлить ключ":
		h.processRenewKey(chatID, userID)

	case "Статус ключа":
		h.processKeyStatus(chatID, userID)
	}
}

func (h *Handler) processBuyVPN(chatID int64, userID int) {
	user, err := h.userService.GetUserByTelegramID(int64(userID))
	if err != nil {
		log.Printf("Ошибка получения пользователя %d: %v", userID, err)
		h.sendErrorMessage(chatID, "Ошибка получения данных пользователя. Попробуйте позже.")
		return
	}

	hasKeys, err := h.vpnKeyService.HasFreeKeys()
	if err != nil || !hasKeys {
		h.sendErrorMessage(chatID, "⚠️ Временно нет свободных VPN-ключей. Попробуйте позже.")
		return
	}

	paymentURL, err := h.paymentService.CreatePayment(user.ID, 299, "Покупка VPN")
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при создании платежа. Попробуйте позже.")
		return
	}

	h.sendMessageText(chatID, fmt.Sprintf("💳 Оплатите по ссылке: %s", paymentURL))
}

func (h *Handler) processMyKeys(chatID int64, userID int) {
	keys, err := h.vpnKeyService.GetKeysByUserTelegramID(int64(userID))
	if err != nil {
		log.Printf("Ошибка получения ключей пользователя %d: %v", userID, err)
		h.sendErrorMessage(chatID, "Ошибка при получении ключей.")
		return
	}

	if len(keys) == 0 {
		h.sendMessageText(chatID, "🔑 У вас нет купленных ключей.")
		return
	}

	var text strings.Builder
	text.WriteString("🔑 Ваши VPN-ключи:\n")
	for _, k := range keys {
		text.WriteString(fmt.Sprintf("Ключ: `%s`, истекает: *%v*\n", k.Key, k.ExpiresAt.Format("02.01.2006")))
	}
	h.sendMessageMarkdown(chatID, text.String())
}

func (h *Handler) processRenewKey(chatID int64, userID int) {
	user, err := h.userService.GetUserByTelegramID(int64(userID))
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка получения данных пользователя.")
		return
	}

	paymentURL, err := h.paymentService.CreatePayment(user.ID, 199, "Продление VPN")
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при создании платежа. Попробуйте позже.")
		return
	}

	h.sendMessageText(chatID, fmt.Sprintf("🔄 Оплатите продление по ссылке: %s", paymentURL))
}

func (h *Handler) processKeyStatus(chatID int64, userID int) {
	keys, err := h.vpnKeyService.GetKeysByUserTelegramID(int64(userID))
	if err != nil {
		h.sendErrorMessage(chatID, "Ошибка при получении ключей.")
		return
	}

	if len(keys) == 0 {
		h.sendMessageText(chatID, "📌 У вас нет активных VPN-ключей.")
		return
	}

	var activeKeys []string
	for _, k := range keys {
		activeKeys = append(activeKeys, fmt.Sprintf("🔑 `%s` (Истекает: *%v*)", k.Key, k.ExpiresAt.Format("02.01.2006")))
	}

	h.sendMessageMarkdown(chatID, "📌 Ваши активные ключи:\n"+strings.Join(activeKeys, "\n"))
}

func (h *Handler) sendMenuKeyboard(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = mainMenuKeyboard()
	h.bot.Send(msg)
}

func (h *Handler) handleAddKeyCommand(chatID int64, text string) {
	// text может быть "/add_key 12345-ABC-666..."
	parts := splitBySpace(text)
	if len(parts) < 2 {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка: нужно указать ключ. Пример: /add_key 12345"))
		return
	}
	key := parts[1]

	// Обращаемся к vpnKeyService. Можно сделать метод service.AddKey(...)
	// Но здесь для простоты дернем repo напрямую (хотя по SOLID — лучше через сервис).
	err := h.vpnKeyService.AddNewKey(key)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "Ошибка добавления ключа: "+err.Error()))
		return
	}
	h.bot.Send(tgbotapi.NewMessage(chatID, "Ключ успешно добавлен."))
}

func splitBySpace(s string) []string {
	// или strings.Fields(s)
	// но Fields() режет по всем пробелам
	// если нужно строго разделить по первому пробелу — можно сделать иначе
	return strings.Split(s, " ")
}

func (h *Handler) handleCallbackQuery(update tgbotapi.Update) {
	cb := update.CallbackQuery
	data := cb.Data
	chatID := cb.Message.Chat.ID

	if h.bot == nil {
		log.Println("❌ Ошибка: bot не инициализирован")
		return
	}

	switch data {
	case "buy_vpn":
		user, err := h.userService.GetUserByTelegramID(cb.From.ID)
		if err != nil {
			log.Println("❌ Ошибка получения пользователя:", err)
			h.sendErrorMessage(chatID, "Ошибка получения данных пользователя. Попробуйте позже.")
			return
		}

		// ✅ Проверяем наличие свободных VPN-ключей перед оплатой
		hasKeys, err := h.vpnKeyService.HasFreeKeys()
		if err != nil {
			log.Println("❌ Ошибка проверки VPN-ключей:", err)
			h.sendErrorMessage(chatID, "Ошибка при проверке VPN-ключей. Попробуйте позже.")
			return
		}
		if !hasKeys {
			log.Println("⛔ Нет свободных VPN-ключей, отменяем оплату.")
			h.sendErrorMessage(chatID, "⚠️ Временно нет свободных VPN-ключей. Попробуйте позже.")
			return
		}

		// ✅ Если ключи есть — создаем ссылку на оплату
		confirmationURL, err := h.paymentService.CreatePayment(user.ID, 299, "Покупка VPN")
		if err != nil {
			log.Println("❌ Ошибка создания платежа:", err)
			h.sendErrorMessage(chatID, "Ошибка при создании платежа. Попробуйте позже.")
			return
		}

		// Отправляем пользователю ссылку на оплату
		h.sendMessageText(chatID, fmt.Sprintf("💳 Оплатите по ссылке: %s", confirmationURL))

	case "my_keys":
		keys, err := h.vpnKeyService.GetKeysByUserTelegramID(cb.From.ID)
		if err != nil {
			log.Println("❌ Ошибка получения ключей:", err)
			h.sendErrorMessage(chatID, "Ошибка при получении ключей.")
			return
		}

		if len(keys) == 0 {
			h.sendMessageText(chatID, "🔑 У вас нет купленных ключей.")
			return
		}

		var text strings.Builder
		text.WriteString("🔑 Ваши VPN-ключи:\n")
		for _, k := range keys {
			text.WriteString(fmt.Sprintf("Ключ: `%s`, истекает: *%v*\n", k.Key, k.ExpiresAt.Format("02.01.2006")))
		}
		h.sendMessageMarkdown(chatID, text.String())

	case "renew_key":
		user, err := h.userService.GetUserByTelegramID(cb.From.ID)
		if err != nil {
			log.Println("❌ Ошибка получения пользователя:", err)
			h.sendErrorMessage(chatID, "Ошибка получения данных пользователя.")
			return
		}

		confirmationURL, err := h.paymentService.CreatePayment(user.ID, 199, "Продление VPN")
		if err != nil {
			log.Println("❌ Ошибка создания платежа:", err)
			h.sendErrorMessage(chatID, "Ошибка при создании платежа. Попробуйте позже.")
			return
		}

		h.sendMessageText(chatID, fmt.Sprintf("🔄 Оплатите продление по ссылке: %s", confirmationURL))

	case "status_key":
		keys, err := h.vpnKeyService.GetKeysByUserTelegramID(cb.From.ID)
		if err != nil {
			log.Println("❌ Ошибка получения ключей:", err)
			h.sendErrorMessage(chatID, "Ошибка при получении ключей.")
			return
		}

		if len(keys) == 0 {
			h.sendMessageText(chatID, "📌 У вас нет активных VPN-ключей.")
			return
		}

		var activeKeys []string
		for _, k := range keys {
			activeKeys = append(activeKeys, fmt.Sprintf("🔑 `%s` (Истекает: *%v*)", k.Key, k.ExpiresAt.Format("02.01.2006")))
		}

		h.sendMessageMarkdown(chatID, "📌 Ваши активные ключи:\n"+strings.Join(activeKeys, "\n"))

	default:
		h.sendMessageText(chatID, "❓ Неизвестная команда.")
	}

	// ✅ Подтверждаем callback
	callback := tgbotapi.NewCallback(cb.ID, "")
	h.bot.Request(callback)
}

func mainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	buttons := [][]tgbotapi.KeyboardButton{
		{tgbotapi.NewKeyboardButton("Купить VPN")},
		{tgbotapi.NewKeyboardButton("Мои ключи")},
		{tgbotapi.NewKeyboardButton("Продлить ключ")},
		{tgbotapi.NewKeyboardButton("Статус ключа")},
	}

	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        buttons,
		ResizeKeyboard:  true,  // Уменьшает кнопки, делая их удобными
		OneTimeKeyboard: false, // Клавиатура не исчезает после нажатия
	}
}

func (h *Handler) handleBuyVPN(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	userID := callback.From.ID

	if h.bot == nil {
		log.Println("❌ Ошибка: bot не инициализирован")
		return
	}

	// Проверяем наличие свободных VPN-ключей
	hasKeys, err := h.vpnKeyService.HasFreeKeys()
	if err != nil {
		log.Println("Ошибка проверки VPN-ключей:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при проверке VPN-ключей. Попробуйте позже.")
		_, _ = h.bot.Send(msg)
		return
	}

	if !hasKeys {
		log.Println("⛔ Нет свободных VPN-ключей, отменяем оплату.")
		msg := tgbotapi.NewMessage(chatID, "⚠️ Временно нет свободных VPN-ключей. Пожалуйста, попробуйте позже.")
		_, _ = h.bot.Send(msg)
		return
	}

	// ✅ Если ключи есть — создаем ссылку на оплату
	paymentURL, err := h.paymentService.CreatePayment(
		int(userID),           // ID пользователя
		299.00,                // Цена (например, 299 рублей)
		"Оплата VPN-подписки", // Описание платежа
	)
	if err != nil {
		log.Println("Ошибка создания платежа:", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при создании платежа. Попробуйте позже.")
		_, _ = h.bot.Send(msg)
		return
	}

	// Отправляем ссылку на оплату пользователю
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("💳 Оплатите по ссылке: %s", paymentURL))
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Println("Ошибка отправки сообщения в Telegram:", err)
	}
}

func (h *Handler) sendMessageText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	response, err := h.bot.Send(msg)

	if err != nil {
		log.Printf("❌ Ошибка отправки сообщения: %v", err)
	} else {
		log.Printf("✅ Сообщение отправлено: ID %d", response.MessageID)
	}
}

func (h *Handler) sendMessageMarkdown(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Println("❌ Ошибка отправки Markdown-сообщения:", err)
	}
}

func (h *Handler) sendErrorMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, "❌ "+text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Println("❌ Ошибка отправки сообщения об ошибке:", err)
	}
}
