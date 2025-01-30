package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type YooKassaWebhook struct {
	Type   string `json:"type"`
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
	} `json:"object"`
}

func (h *Handler) HandleYooKassaWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("🔔 Получен вебхук от YooKassa!")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Ошибка чтения тела:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	for k, v := range r.Header {
		log.Printf("📜 Заголовок: %s = %s\n", k, v)
	}

	if r.Header.Get("X-Content-Signature") != "" {
		if !h.verifySignature(body, r.Header.Get("X-Content-Signature")) {
			log.Println("❌ Ошибка: подпись вебхука не совпадает!")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var webhook YooKassaWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Println("Ошибка парсинга JSON:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📩 Вебхук: %s, ID платежа: %s, статус: %s, сумма: %s %s",
		webhook.Event, webhook.Object.ID, webhook.Object.Status, webhook.Object.Amount.Value, webhook.Object.Amount.Currency)

	switch webhook.Event {
	case "payment.succeeded":
		err = h.paymentService.ConfirmPayment(webhook.Object.ID)
		if err != nil {
			log.Println("Ошибка подтверждения платежа:", err)
		} else {
			log.Println("✅ Платёж успешно обработан, VPN-ключ выдан.")
		}

	case "payment.waiting_for_capture":
		log.Println("⚠️ Платёж требует подтверждения (waiting_for_capture).")

	case "payment.canceled":
		log.Println("❌ Платёж отменён или произошла ошибка.")

	case "refund.succeeded":
		log.Println("🔄 Возврат средств выполнен. Нужно аннулировать VPN-ключ.")

	default:
		log.Println("❓ Неизвестное событие:", webhook.Event)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) verifySignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(h.secretKey))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)

	receivedMAC, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Println("Ошибка декодирования подписи:", err)
		return false
	}

	return hmac.Equal(expectedMAC, receivedMAC)
}
