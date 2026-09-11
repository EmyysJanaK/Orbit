package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"employee-management-system/models"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type EventHandler struct {
	Events *repository.EventRepository
	Redis  *redis.Client
}

type usageEventRequest struct {
	EventType string         `json:"event_type" binding:"required"`
	Metadata  map[string]any `json:"metadata"`
}

func NewEventHandler(events *repository.EventRepository, redisClient *redis.Client) *EventHandler {
	return &EventHandler{Events: events, Redis: redisClient}
}

func (h *EventHandler) Create(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user context"})
		return
	}

	var req usageEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventType := normalizeEventType(req.EventType)
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_type is required"})
		return
	}

	event, err := h.Events.Create(c.Request.Context(), models.Event{
		UserID:    userID,
		EventType: eventType,
		Metadata:  req.Metadata,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store event"})
		return
	}

	dayKey := time.Now().UTC().Format("2006-01-02")
	dauKey := fmt.Sprintf("analytics:dau:%s", dayKey)
	eventCountKey := fmt.Sprintf("analytics:event_count:%s:%s", dayKey, sanitizeRedisKeyPart(eventType))

	pipe := h.Redis.Pipeline()
	pipe.SAdd(c.Request.Context(), dauKey, strconv.FormatInt(userID, 10))
	pipe.Expire(c.Request.Context(), dauKey, 8*24*time.Hour)
	pipe.Incr(c.Request.Context(), eventCountKey)
	pipe.Expire(c.Request.Context(), eventCountKey, 8*24*time.Hour)
	if _, err := pipe.Exec(c.Request.Context()); err != nil {
		if deleteErr := h.Events.DeleteByID(c.Request.Context(), event.ID); deleteErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store event"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update analytics"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

func normalizeEventType(eventType string) string {
	return strings.ToLower(strings.TrimSpace(eventType))
}

func sanitizeRedisKeyPart(value string) string {
	value = normalizeEventType(value)
	replacer := strings.NewReplacer(" ", "_", ":", "_", "/", "_", "\\", "_")
	return replacer.Replace(value)
}
