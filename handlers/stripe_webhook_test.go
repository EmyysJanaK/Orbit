package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"employee-management-system/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeStripeWebhookRepository struct {
	seen map[string]bool
}

func newFakeStripeWebhookRepository() *fakeStripeWebhookRepository {
	return &fakeStripeWebhookRepository{seen: map[string]bool{}}
}

func (f *fakeStripeWebhookRepository) ProcessStripeWebhookEvent(ctx context.Context, stripeEventID, eventType string, payment models.Payment) (bool, error) {
	if f.seen[stripeEventID] {
		return false, nil
	}

	f.seen[stripeEventID] = true
	return true, nil
}

func TestStripeWebhookHandlerIdempotency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeStripeWebhookRepository()
	handler := NewStripeWebhookHandler(repo, "whsec_test")

	body := []byte(`{
		"id":"evt_123",
		"type":"payment_intent.succeeded",
		"data": {
			"object": {
				"id":"pi_123",
				"amount":5000,
				"currency":"usd",
				"metadata": {
					"appointment_id":"42",
					"user_id":"7"
				}
			}
		}
	}`)

	first := performStripeWebhookRequest(t, handler, body, "whsec_test")
	require.Equal(t, http.StatusOK, first.Code)
	require.Contains(t, first.Body.String(), `"processed":true`)

	second := performStripeWebhookRequest(t, handler, body, "whsec_test")
	require.Equal(t, http.StatusOK, second.Code)
	require.Contains(t, second.Body.String(), `"processed":false`)
}

func TestStripeWebhookHandlerRejectsBadSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewStripeWebhookHandler(newFakeStripeWebhookRepository(), "whsec_test")
	body := []byte(`{"id":"evt_123","type":"payment_intent.succeeded","data":{"object":{"id":"pi_123","amount":5000,"currency":"usd","metadata":{"appointment_id":"42","user_id":"7"}}}}`)

	recorder := performStripeWebhookRequest(t, handler, body, "wrong_secret")
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func performStripeWebhookRequest(t *testing.T, handler *StripeWebhookHandler, body []byte, secret string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	router := gin.New()
	router.POST("/api/webhooks/stripe", handler.Handle)

	request := httptest.NewRequest(http.MethodPost, "/api/webhooks/stripe", bytes.NewReader(body))
	request.Header.Set("Stripe-Signature", stripeSignature(body, secret, time.Unix(1710000000, 0)))
	router.ServeHTTP(recorder, request)

	return recorder
}

func stripeSignature(payload []byte, secret string, ts time.Time) string {
	base := fmt.Sprintf("%d.%s", ts.Unix(), payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(base))
	signature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts.Unix(), signature)
}
