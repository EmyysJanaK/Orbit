package jobs

import (
	"context"
	"log"
	"time"

	"employee-management-system/models"
	"employee-management-system/repository"
)

type ReminderJob struct {
	Appointments *repository.AppointmentRepository
	Events       *repository.EventRepository
}

func NewReminderJob(appointments *repository.AppointmentRepository, events *repository.EventRepository) *ReminderJob {
	return &ReminderJob{Appointments: appointments, Events: events}
}

func (j *ReminderJob) Run(ctx context.Context) error {
	now := time.Now().UTC()
	windowEnd := now.Add(24 * time.Hour)

	appointments, err := j.Appointments.ListUpcomingWithPendingPayments(ctx, now, windowEnd)
	if err != nil {
		return err
	}

	for _, appointment := range appointments {
		log.Printf("appointment payment reminder: appointment_id=%d user_id=%d scheduled_at=%s", appointment.ID, appointment.UserID, appointment.ScheduledAt.UTC().Format(time.RFC3339))

		_, err := j.Events.Create(ctx, models.Event{
			UserID:    appointment.UserID,
			EventType: "appointment_payment_reminder",
			Metadata: map[string]any{
				"appointment_id": appointment.ID,
				"title":          appointment.Title,
				"scheduled_at":   appointment.ScheduledAt.UTC().Format(time.RFC3339),
			},
		})
		if err != nil {
			log.Printf("appointment reminder event insert failed: appointment_id=%d error=%v", appointment.ID, err)
		}
	}

	return nil
}
