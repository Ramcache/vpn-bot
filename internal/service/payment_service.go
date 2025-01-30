// internal/service/payment_service.go
package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"time"
	"vpn-bot/internal/repository"
)

type paymentServiceImpl struct {
	repo          repository.PaymentRepository
	vpnKeyService VPNKeyService

	yooShopID string
	yooSecret string
}

// Структуры для запросов/ответов к YooKassa (упрощённо)

// Создание платежа (request)
type yooCreatePaymentRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Capture      bool   `json:"capture"`
	Description  string `json:"description"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
}

// Ответ на создание платежа
type yooCreatePaymentResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

func NewPaymentService(
	payRepo repository.PaymentRepository,
	vpnService VPNKeyService,
	shopID, secret string,
) PaymentService {
	return &paymentServiceImpl{
		repo:          payRepo,
		vpnKeyService: vpnService,
		yooShopID:     shopID,
		yooSecret:     secret,
	}
}

// CreatePayment создаёт платёж через YooKassa и возвращает paymentID и ссылку на оплату.
func (s *paymentServiceImpl) CreatePayment(
	userID int,
	amount float64,
	description string,
) (string, error) {

	// 1. Формируем JSON для запроса
	reqBody := yooCreatePaymentRequest{}
	reqBody.Amount.Value = fmt.Sprintf("%.2f", amount)
	reqBody.Amount.Currency = "RUB"
	reqBody.Capture = true
	reqBody.Description = description
	// Адрес, куда вернётся пользователь после оплаты (может быть любым)
	reqBody.Confirmation.Type = "redirect"
	reqBody.Confirmation.ReturnURL = "https://ramcache.online/payment-success"

	// 2. Кодируем в JSON
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// 3. Отправляем POST-запрос к YooKassa
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	// Указываем заголовки
	// shopId и секретный ключ обычно в basic auth:
	//   Authorization: Basic base64("SHOP_ID:SECRET_KEY")
	// Либо как указано в документации (ключ API)
	req.SetBasicAuth(s.yooShopID, s.yooSecret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("my-key-%d", time.Now().UnixNano())) // уникальный ключ запроса
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Считываем тело, чтобы увидеть, что пошло не так
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка от YooKassa: %d %s", resp.StatusCode, string(errBody))
	}

	var yooResp yooCreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&yooResp); err != nil {
		return "", err
	}

	paymentID := yooResp.ID  // Уникальный ID от YooKassa
	status := yooResp.Status // pending
	confirmationURL := yooResp.Confirmation.ConfirmationURL

	// 4. Сохраняем платёж в нашей БД
	err = s.repo.CreatePayment(userID, amount, status, paymentID)
	if err != nil {
		return "", err
	}

	// 5. Возвращаем paymentID (для вебхука) и ссылку на оплату
	//    Чтобы пользователь мог перейти и оплатить
	return confirmationURL, nil
}

// ConfirmPayment – подтверждает платёж после вебхука "payment.succeeded"

func (s *paymentServiceImpl) ConfirmPayment(paymentID string) error {
	pay, err := s.repo.GetByPaymentID(paymentID)
	if err != nil {
		return err
	}
	if pay.Status == "succeeded" {
		return nil // Уже подтвержден
	}

	// Обновляем статус в БД
	err = s.repo.UpdatePaymentStatus(pay.ID, "succeeded")
	if err != nil {
		return err
	}

	// Пробуем выдать VPN-ключ
	key, err := s.vpnKeyService.AssignFreeKeyToUser(pay.UserID)
	if err != nil {
		log.Println("❌ Ошибка при выдаче VPN-ключа:", err)

		// Уведомляем пользователя в Telegram
		msg := fmt.Sprintf("✅ Оплата прошла, но пока нет свободных VPN-ключей. Мы скоро их добавим и пришлём вам ключ.")
		s.sendTelegramMessage(pay.UserID, msg)

		return errors.New("нет свободных VPN-ключей")
	}

	// Отправляем ключ пользователю в Telegram
	msg := fmt.Sprintf("✅ Оплата прошла успешно! Ваш VPN-ключ: %s", key)
	s.sendTelegramMessage(pay.UserID, msg)

	log.Println("✅ Пользователю отправлен VPN-ключ:", key)
	return nil
}

// Отправка сообщения пользователю
func (s *paymentServiceImpl) sendTelegramMessage(userID int, text string) {
	bot, err := tgbotapi.NewBotAPI("YOUR_BOT_TOKEN")
	if err != nil {
		log.Println("❌ Ошибка создания Telegram-бота:", err)
		return
	}

	msg := tgbotapi.NewMessage(int64(userID), text)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("❌ Ошибка отправки сообщения пользователю:", err)
	}
}
