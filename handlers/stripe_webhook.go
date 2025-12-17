package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"employee-management-system/models"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type StripeWebhookHandler struct {
	Payments      *repository.PaymentRepository
	WebhookSecret string
}

func NewStripeWebhookHandler(payments *repository.PaymentRepository, webhookSecret string) *StripeWebhookHandler {
	return &StripeWebhookHandler{Payments: payments, WebhookSecret: webhookSecret}
}

func (h *StripeWebhookHandler) Handle(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read request body"})
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing stripe signature"})
		return
	}

	event, err := webhook.ConstructEvent(payload, signature, h.WebhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stripe signature"})
		return
	}

	eventType := string(event.Type)
	if !strings.HasPrefix(eventType, "payment_intent.") {
		c.JSON(http.StatusOK, gin.H{"received": true, "processed": false})
		return
	}

	var paymentIntent stripe.PaymentIntent
	if len(event.Data.Raw) > 0 {
		if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment intent payload"})
			return
		}
	}

	payment, err := paymentFromIntent(paymentIntent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processed, err := h.Payments.ProcessStripeWebhookEvent(c.Request.Context(), event.ID, eventType, payment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process stripe event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"received":  true,
		"processed": processed,
	})
}

func paymentFromIntent(intent stripe.PaymentIntent) (models.Payment, error) {
	if intent.ID == "" {
		return models.Payment{}, simpleError("missing payment intent payload")
	}

	appointmentID, err := metadataInt64(intent.Metadata, "appointment_id")
	if err != nil {
		return models.Payment{}, err
	}

	userID, err := metadataInt64(intent.Metadata, "user_id")
	if err != nil {
		return models.Payment{}, err
	}

	return models.Payment{
		UserID:                userID,
		AppointmentID:         appointmentID,
		StripePaymentIntentID: intent.ID,
		Amount:                intent.Amount,
		Currency:              strings.ToLower(string(intent.Currency)),
	}, nil
}

func metadataInt64(metadata map[string]string, key string) (int64, error) {
	value, ok := metadata[key]
	if !ok || strings.TrimSpace(value) == "" {
		return 0, fmtError("missing %s metadata", key)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmtError("invalid %s metadata", key)
	}

	return parsed, nil
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

func fmtError(format string, args ...any) error {
	return simpleError(fmt.Sprintf(format, args...))
}
