package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"employee-management-system/models"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeAppointmentRepository struct {
	appointments map[int64]models.Appointment
	nextID       int64
	lastCreated  models.Appointment
}

func newFakeAppointmentRepository() *fakeAppointmentRepository {
	return &fakeAppointmentRepository{
		appointments: map[int64]models.Appointment{},
		nextID:       1,
	}
}

func (f *fakeAppointmentRepository) Create(ctx context.Context, appointment models.Appointment) (models.Appointment, error) {
	appointment.ID = f.nextID
	appointment.CreatedAt = time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	f.nextID++
	f.appointments[appointment.ID] = appointment
	f.lastCreated = appointment
	return appointment, nil
}

func (f *fakeAppointmentRepository) ListByUser(ctx context.Context, userID int64) ([]models.Appointment, error) {
	items := make([]models.Appointment, 0)
	for _, appointment := range f.appointments {
		if appointment.UserID == userID {
			items = append(items, appointment)
		}
	}
	return items, nil
}

func (f *fakeAppointmentRepository) ListAll(ctx context.Context) ([]models.Appointment, error) {
	items := make([]models.Appointment, 0, len(f.appointments))
	for _, appointment := range f.appointments {
		items = append(items, appointment)
	}
	return items, nil
}

func (f *fakeAppointmentRepository) GetByID(ctx context.Context, id int64) (models.Appointment, error) {
	appointment, ok := f.appointments[id]
	if !ok {
		return models.Appointment{}, repository.ErrAppointmentNotFound
	}
	return appointment, nil
}

func (f *fakeAppointmentRepository) Update(ctx context.Context, appointment models.Appointment) (models.Appointment, error) {
	if _, ok := f.appointments[appointment.ID]; !ok {
		return models.Appointment{}, repository.ErrAppointmentNotFound
	}
	f.appointments[appointment.ID] = appointment
	return appointment, nil
}

func (f *fakeAppointmentRepository) Delete(ctx context.Context, id int64) error {
	if _, ok := f.appointments[id]; !ok {
		return repository.ErrAppointmentNotFound
	}
	delete(f.appointments, id)
	return nil
}

func TestAppointmentHandlerCreateListGetPatchDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeAppointmentRepository()
	repo.appointments[1] = models.Appointment{ID: 1, UserID: 7, Title: "Existing", ScheduledAt: time.Date(2026, 6, 29, 9, 0, 0, 0, time.UTC), Status: "booked", CreatedAt: time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)}
	repo.appointments[2] = models.Appointment{ID: 2, UserID: 8, Title: "Other", ScheduledAt: time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC), Status: "booked", CreatedAt: time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)}
	handler := NewAppointmentHandler(repo)

	t.Run("create", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(7))
			c.Set("role", "user")
		})
		router.POST("/api/appointments", handler.Create)

		body := []byte(`{"title":"Consultation","scheduled_at":"2026-06-29T12:00:00Z","status":"booked"}`)
		request := httptest.NewRequest(http.MethodPost, "/api/appointments", bytes.NewReader(body))
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusCreated, recorder.Code)
		var created models.Appointment
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &created))
		require.Equal(t, int64(7), created.UserID)
		require.Equal(t, "Consultation", created.Title)
	})

	t.Run("list admin sees all", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(1))
			c.Set("role", "admin")
		})
		router.GET("/api/appointments", handler.List)

		request := httptest.NewRequest(http.MethodGet, "/api/appointments", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		var items []models.Appointment
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &items))
		require.Len(t, items, 2)
	})

	t.Run("list user sees own only", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(7))
			c.Set("role", "user")
		})
		router.GET("/api/appointments", handler.List)

		request := httptest.NewRequest(http.MethodGet, "/api/appointments", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		var items []models.Appointment
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &items))
		require.Len(t, items, 2)
	})

	t.Run("get forbidden for other user", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(999))
			c.Set("role", "user")
		})
		router.GET("/api/appointments/:id", handler.Get)

		request := httptest.NewRequest(http.MethodGet, "/api/appointments/1", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("patch update", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(7))
			c.Set("role", "user")
		})
		router.PATCH("/api/appointments/:id", handler.Patch)

		body := []byte(`{"status":"confirmed"}`)
		request := httptest.NewRequest(http.MethodPatch, "/api/appointments/1", bytes.NewReader(body))
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusOK, recorder.Code)
		var updated models.Appointment
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &updated))
		require.Equal(t, "confirmed", updated.Status)
	})

	t.Run("delete no content", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", int64(7))
			c.Set("role", "user")
		})
		router.DELETE("/api/appointments/:id", handler.Delete)

		request := httptest.NewRequest(http.MethodDelete, "/api/appointments/1", nil)
		router.ServeHTTP(recorder, request)

		require.Equal(t, http.StatusNoContent, recorder.Code)
		require.NotContains(t, repo.appointments, int64(1))
	})
}

func TestAppointmentHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAppointmentHandler(newFakeAppointmentRepository())

	recorder := httptest.NewRecorder()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Set("role", "user")
	})
	router.POST("/api/appointments", handler.Create)

	request := httptest.NewRequest(http.MethodPost, "/api/appointments", bytes.NewReader([]byte(`{"title":"","scheduled_at":"not-a-time","status":""}`)))
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAppointmentHandler404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAppointmentHandler(newFakeAppointmentRepository())

	recorder := httptest.NewRecorder()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(7))
		c.Set("role", "admin")
	})
	router.GET("/api/appointments/:id", handler.Get)

	request := httptest.NewRequest(http.MethodGet, "/api/appointments/99", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}
