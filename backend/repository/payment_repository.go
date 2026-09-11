package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"employee-management-system/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) UpsertPending(ctx context.Context, payment models.Payment) (models.Payment, error) {
	return r.upsertPayment(ctx, payment, "pending", true)
}

func (r *PaymentRepository) UpsertSucceeded(ctx context.Context, payment models.Payment) (models.Payment, error) {
	return r.upsertPayment(ctx, payment, "succeeded", false)
}

func (r *PaymentRepository) UpsertFailed(ctx context.Context, payment models.Payment) (models.Payment, error) {
	return r.upsertPayment(ctx, payment, "failed", false)
}

func (r *PaymentRepository) ProcessStripeWebhookEvent(ctx context.Context, stripeEventID, eventType string, payment models.Payment) (bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var recordedID string
	err = tx.QueryRow(
		ctx,
		`INSERT INTO stripe_webhook_events (stripe_event_id, event_type)
		 VALUES ($1, $2)
		 ON CONFLICT (stripe_event_id) DO NOTHING
		 RETURNING stripe_event_id`,
		stripeEventID,
		eventType,
	).Scan(&recordedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	switch eventType {
	case "payment_intent.created":
		_, err = r.upsertPaymentTx(ctx, tx, payment, "pending", true)
	case "payment_intent.succeeded":
		_, err = r.upsertPaymentTx(ctx, tx, payment, "succeeded", false)
	case "payment_intent.payment_failed", "payment_failed":
		_, err = r.upsertPaymentTx(ctx, tx, payment, "failed", false)
	default:
		// Keep the webhook idempotent for unsupported event types as well.
		_, err = r.upsertPaymentTx(ctx, tx, payment, "pending", true)
	}
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}

func (r *PaymentRepository) upsertPayment(ctx context.Context, payment models.Payment, status string, preserveFinal bool) (models.Payment, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Payment{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	updated, err := r.upsertPaymentTx(ctx, tx, payment, status, preserveFinal)
	if err != nil {
		return models.Payment{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Payment{}, err
	}

	return updated, nil
}

func (r *PaymentRepository) upsertPaymentTx(ctx context.Context, querier pgx.Tx, payment models.Payment, status string, preserveFinal bool) (models.Payment, error) {
	if payment.StripePaymentIntentID == "" {
		return models.Payment{}, errors.New("stripe payment intent id is required")
	}

	if strings.TrimSpace(payment.Currency) == "" {
		payment.Currency = "usd"
	}

	sql := `INSERT INTO payments (user_id, appointment_id, stripe_payment_intent_id, amount, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_payment_intent_id) DO UPDATE SET
			user_id = CASE WHEN payments.status = 'pending' OR payments.status IS NULL THEN EXCLUDED.user_id ELSE payments.user_id END,
			appointment_id = CASE WHEN payments.status = 'pending' OR payments.status IS NULL THEN EXCLUDED.appointment_id ELSE payments.appointment_id END,
			amount = CASE WHEN payments.status = 'pending' OR payments.status IS NULL THEN EXCLUDED.amount ELSE payments.amount END,
			currency = CASE WHEN payments.status = 'pending' OR payments.status IS NULL THEN EXCLUDED.currency ELSE payments.currency END,
			status = CASE WHEN $7::boolean THEN CASE WHEN payments.status = 'pending' OR payments.status IS NULL THEN EXCLUDED.status ELSE payments.status END ELSE EXCLUDED.status END
		RETURNING id, user_id, appointment_id, stripe_payment_intent_id, amount, currency, status, created_at`

	var updated models.Payment
	err := querier.QueryRow(
		ctx,
		sql,
		payment.UserID,
		payment.AppointmentID,
		payment.StripePaymentIntentID,
		payment.Amount,
		payment.Currency,
		status,
		preserveFinal,
	).Scan(
		&updated.ID,
		&updated.UserID,
		&updated.AppointmentID,
		&updated.StripePaymentIntentID,
		&updated.Amount,
		&updated.Currency,
		&updated.Status,
		&updated.CreatedAt,
	)
	if err != nil {
		return models.Payment{}, err
	}

	return updated, nil
}

func (r *PaymentRepository) GetByAppointmentID(ctx context.Context, appointmentID int64) (models.Payment, error) {
	var payment models.Payment
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, user_id, appointment_id, stripe_payment_intent_id, amount, currency, status, created_at
		 FROM payments
		 WHERE appointment_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT 1`,
		appointmentID,
	).Scan(
		&payment.ID,
		&payment.UserID,
		&payment.AppointmentID,
		&payment.StripePaymentIntentID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Payment{}, ErrPaymentNotFound
		}

		return models.Payment{}, err
	}

	return payment, nil
}

func (r *PaymentRepository) ListPendingPaymentAppointments(ctx context.Context, start, endTime time.Time) ([]models.Appointment, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT DISTINCT a.id, a.user_id, a.title, a.scheduled_at, a.status, a.created_at
		 FROM appointments a
		 INNER JOIN payments p ON p.appointment_id = a.id
		 WHERE a.scheduled_at >= $1
		   AND a.scheduled_at <= $2
		   AND p.status = 'pending'
		 ORDER BY a.scheduled_at ASC, a.id ASC`,
		start,
		endTime,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appointments := make([]models.Appointment, 0)
	for rows.Next() {
		var appointment models.Appointment
		if err := rows.Scan(&appointment.ID, &appointment.UserID, &appointment.Title, &appointment.ScheduledAt, &appointment.Status, &appointment.CreatedAt); err != nil {
			return nil, err
		}
		appointments = append(appointments, appointment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r *PaymentRepository) TotalSuccessfulRevenueSince(ctx context.Context, since time.Time) (int64, error) {
	var revenue int64
	err := r.pool.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM payments
		 WHERE status = 'succeeded'
		   AND created_at >= $1`,
		since,
	).Scan(&revenue)
	if err != nil {
		return 0, err
	}

	return revenue, nil
}
