package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"employee-management-system/models"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	Appointments *repository.AppointmentRepository
}

type appointmentCreateRequest struct {
	Title       string `json:"title" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"`
	Status      string `json:"status" binding:"required"`
}

type appointmentPatchRequest struct {
	Title       *string `json:"title"`
	ScheduledAt *string `json:"scheduled_at"`
	Status      *string `json:"status"`
}

func NewAppointmentHandler(repo *repository.AppointmentRepository) *AppointmentHandler {
	return &AppointmentHandler{Appointments: repo}
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user context"})
		return
	}

	var req appointmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ScheduledAt))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be RFC3339"})
		return
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	appointment, err := h.Appointments.Create(c.Request.Context(), models.Appointment{
		UserID:      userID,
		Title:       title,
		ScheduledAt: scheduledAt,
		Status:      status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create appointment"})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

func (h *AppointmentHandler) List(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user context"})
		return
	}
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))

	var (
		appointments []models.Appointment
		err          error
	)
	if role == "admin" {
		appointments, err = h.Appointments.ListAll(c.Request.Context())
	} else {
		appointments, err = h.Appointments.ListByUser(c.Request.Context(), userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list appointments"})
		return
	}

	c.JSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) Get(c *gin.Context) {
	appointment, allowed, found := h.loadAuthorizedAppointment(c)
	if !found {
		return
	}
	if !allowed {
		return
	}

	c.JSON(http.StatusOK, appointment)
}

func (h *AppointmentHandler) Patch(c *gin.Context) {
	appointment, allowed, found := h.loadAuthorizedAppointment(c)
	if !found {
		return
	}
	if !allowed {
		return
	}

	var req appointmentPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := appointment
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be empty"})
			return
		}
		updated.Title = title
	}
	if req.ScheduledAt != nil {
		scheduledAt, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.ScheduledAt))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_at must be RFC3339"})
			return
		}
		updated.ScheduledAt = scheduledAt
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status cannot be empty"})
			return
		}
		updated.Status = status
	}

	if req.Title == nil && req.ScheduledAt == nil && req.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return
	}

	saved, err := h.Appointments.Update(c.Request.Context(), updated)
	if err != nil {
		if errors.Is(err, repository.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update appointment"})
		return
	}

	c.JSON(http.StatusOK, saved)
}

func (h *AppointmentHandler) Delete(c *gin.Context) {
	appointment, allowed, found := h.loadAuthorizedAppointment(c)
	if !found {
		return
	}
	if !allowed {
		return
	}

	if err := h.Appointments.Delete(c.Request.Context(), appointment.ID); err != nil {
		if errors.Is(err, repository.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete appointment"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AppointmentHandler) loadAuthorizedAppointment(c *gin.Context) (models.Appointment, bool, bool) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user context"})
		return models.Appointment{}, false, false
	}
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment id"})
		return models.Appointment{}, false, false
	}

	appointment, err := h.Appointments.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
			return models.Appointment{}, false, false
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load appointment"})
		return models.Appointment{}, false, false
	}

	if role != "admin" && appointment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return models.Appointment{}, false, false
	}

	return appointment, true, true
}

func userIDFromContext(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}
