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

var ErrAppointmentNotFound = errors.New("appointment not found")

type AppointmentRepository struct {
	pool *pgxpool.Pool
}

func NewAppointmentRepository(pool *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{pool: pool}
}

func (r *AppointmentRepository) Create(ctx context.Context, appointment models.Appointment) (models.Appointment, error) {
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO appointments (user_id, title, scheduled_at, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		appointment.UserID,
		appointment.Title,
		appointment.ScheduledAt,
		appointment.Status,
	).Scan(&appointment.ID, &appointment.CreatedAt)
	if err != nil {
		return models.Appointment{}, err
	}

	return appointment, nil
}

func (r *AppointmentRepository) ListByUser(ctx context.Context, userID int64) ([]models.Appointment, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT id, user_id, title, scheduled_at, status, created_at
		 FROM appointments
		 WHERE user_id = $1
		 ORDER BY scheduled_at DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAppointments(rows)
}

func (r *AppointmentRepository) ListAll(ctx context.Context) ([]models.Appointment, error) {
	rows, err := r.pool.Query(
		ctx,
		`SELECT id, user_id, title, scheduled_at, status, created_at
		 FROM appointments
		 ORDER BY scheduled_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAppointments(rows)
}

func (r *AppointmentRepository) GetByID(ctx context.Context, id int64) (models.Appointment, error) {
	var appointment models.Appointment
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, user_id, title, scheduled_at, status, created_at
		 FROM appointments
		 WHERE id = $1`,
		id,
	).Scan(&appointment.ID, &appointment.UserID, &appointment.Title, &appointment.ScheduledAt, &appointment.Status, &appointment.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Appointment{}, ErrAppointmentNotFound
		}

		return models.Appointment{}, err
	}

	return appointment, nil
}

func (r *AppointmentRepository) Update(ctx context.Context, appointment models.Appointment) (models.Appointment, error) {
	err := r.pool.QueryRow(
		ctx,
		`UPDATE appointments
		 SET title = $1,
		     scheduled_at = $2,
		     status = $3
		 WHERE id = $4
		 RETURNING id, user_id, title, scheduled_at, status, created_at`,
		appointment.Title,
		appointment.ScheduledAt,
		appointment.Status,
		appointment.ID,
	).Scan(&appointment.ID, &appointment.UserID, &appointment.Title, &appointment.ScheduledAt, &appointment.Status, &appointment.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Appointment{}, ErrAppointmentNotFound
		}

		return models.Appointment{}, err
	}

	return appointment, nil
}

func (r *AppointmentRepository) Delete(ctx context.Context, id int64) error {
	commandTag, err := r.pool.Exec(
		ctx,
		`DELETE FROM appointments WHERE id = $1`,
		id,
	)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return ErrAppointmentNotFound
	}

	return nil
}

func (r *AppointmentRepository) ListUpcomingWithPendingPayments(ctx context.Context, start, end time.Time) ([]models.Appointment, error) {
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
		end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAppointments(rows)
}

func scanAppointments(rows pgx.Rows) ([]models.Appointment, error) {
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

func NormalizeAppointmentStatus(status string) string {
	return strings.TrimSpace(strings.ToLower(status))
}

func AppointmentCreatedAtDefault() time.Time {
	return time.Now().UTC()
}
