package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"employee-management-system/models"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

type PaymentHandler struct {
	Appointments *repository.AppointmentRepository
	Payments     *repository.PaymentRepository
	StripeKey    string
}

type paymentIntentRequest struct {
	AppointmentID int64  `json:"appointment_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
}

func NewPaymentHandler(appointments *repository.AppointmentRepository, payments *repository.PaymentRepository, stripeKey string) *PaymentHandler {
	return &PaymentHandler{Appointments: appointments, Payments: payments, StripeKey: stripeKey}
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user context"})
		return
	}
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))

	var req paymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AppointmentID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "appointment_id must be greater than zero"})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than zero"})
		return
	}

	currency := strings.ToLower(strings.TrimSpace(req.Currency))
	if len(currency) != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be a 3-letter code"})
		return
	}
	if h.StripeKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stripe is not configured"})
		return
	}

	stripe.Key = h.StripeKey

	appointment, err := h.Appointments.GetByID(c.Request.Context(), req.AppointmentID)
	if err != nil {
		if errors.Is(err, repository.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load appointment"})
		return
	}

	if role != "admin" && appointment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"appointment_id": strconv.FormatInt(appointment.ID, 10),
			"user_id":        strconv.FormatInt(appointment.UserID, 10),
			"title":          appointment.Title,
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("could not create payment intent: %v", err)})
		return
	}

	payment, err := h.Payments.UpsertPending(c.Request.Context(), models.Payment{
		UserID:                appointment.UserID,
		AppointmentID:         appointment.ID,
		StripePaymentIntentID: intent.ID,
		Amount:                req.Amount,
		Currency:              currency,
		Status:                "pending",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"payment":       payment,
		"client_secret": intent.ClientSecret,
	})
}
