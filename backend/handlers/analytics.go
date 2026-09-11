package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AnalyticsHandler struct {
	Redis    *redis.Client
	Payments *repository.PaymentRepository
}

func NewAnalyticsHandler(redisClient *redis.Client, payments *repository.PaymentRepository) *AnalyticsHandler {
	return &AnalyticsHandler{Redis: redisClient, Payments: payments}
}

func (h *AnalyticsHandler) Summary(c *gin.Context) {
	if strings.ToLower(strings.TrimSpace(c.GetString("role"))) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	todayKey := time.Now().UTC().Format("2006-01-02")
	dauKey := fmt.Sprintf("analytics:dau:%s", todayKey)
	todayDAU, err := h.Redis.SCard(c.Request.Context(), dauKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load DAU"})
		return
	}

	eventCounts, err := h.todayEventCounts(c.Request.Context(), todayKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load event counts"})
		return
	}

	totalRevenue, err := h.Payments.TotalSuccessfulRevenueSince(c.Request.Context(), time.Now().UTC().AddDate(0, 0, -7))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load revenue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"today_dau":                todayDAU,
		"event_counts":             eventCounts,
		"total_revenue_last_7_days": totalRevenue,
	})
}

func (h *AnalyticsHandler) todayEventCounts(ctx context.Context, dayKey string) (map[string]int64, error) {
	prefix := fmt.Sprintf("analytics:event_count:%s:", dayKey)
	result := make(map[string]int64)

	iter := h.Redis.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		count, err := h.Redis.Get(ctx, key).Int64()
		if err != nil {
			if err == redis.Nil {
				continue
			}

			return nil, err
		}

		eventType := strings.TrimPrefix(key, prefix)
		result[eventType] = count
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
