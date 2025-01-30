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

func (s *paymentServiceImpl) CreatePayment(
	userID int,
	amount float64,
	description string,
) (string, error) {

	reqBody := yooCreatePaymentRequest{}
	reqBody.Amount.Value = fmt.Sprintf("%.2f", amount)
	reqBody.Amount.Currency = "RUB"
	reqBody.Capture = true
	reqBody.Description = description
	reqBody.Confirmation.Type = "redirect"
	reqBody.Confirmation.ReturnURL = "https://ramcache.online/payment-success"

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(s.yooShopID, s.yooSecret)
	req.Header.Set("Idempotence-Key", fmt.Sprintf("my-key-%d", time.Now().UnixNano()))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ошибка от YooKassa: %d %s", resp.StatusCode, string(errBody))
	}

	var yooResp yooCreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&yooResp); err != nil {
		return "", err
	}

	paymentID := yooResp.ID
	status := yooResp.Status
	confirmationURL := yooResp.Confirmation.ConfirmationURL

	err = s.repo.CreatePayment(userID, amount, status, paymentID)
	if err != nil {
		return "", err
	}

	return confirmationURL, nil
}

func (s *paymentServiceImpl) ConfirmPayment(paymentID string) error {
	pay, err := s.repo.GetByPaymentID(paymentID)
	if err != nil {
		return err
	}
	if pay.Status == "succeeded" {
		return nil
	}

	err = s.repo.UpdatePaymentStatus(pay.ID, "succeeded")
	if err != nil {
		return err
	}

	key, err := s.vpnKeyService.AssignFreeKeyToUser(pay.UserID)
	if err != nil {
		log.Println("❌ Ошибка при выдаче VPN-ключа:", err)

		msg := fmt.Sprintf("✅ Оплата прошла, но пока нет свободных VPN-ключей. Мы скоро их добавим и пришлём вам ключ.")
		s.sendTelegramMessage(pay.UserID, msg)

		return errors.New("нет свободных VPN-ключей")
	}

	msg := fmt.Sprintf("✅ Оплата прошла успешно! Ваш VPN-ключ: %s", key)
	s.sendTelegramMessage(pay.UserID, msg)

	log.Println("✅ Пользователю отправлен VPN-ключ:", key)
	return nil
}

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
