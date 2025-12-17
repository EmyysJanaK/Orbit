package main

import (
	"context"
	"log"
	"net/http"

	"employee-management-system/auth"
	"employee-management-system/config"
	"employee-management-system/handlers"
	"employee-management-system/jobs"
	"employee-management-system/middleware"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	postgresPool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer postgresPool.Close()

	if err := postgresPool.Ping(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}

	userRepository := repository.NewUserRepository(postgresPool)
	appointmentRepository := repository.NewAppointmentRepository(postgresPool)
	paymentRepository := repository.NewPaymentRepository(postgresPool)
	eventRepository := repository.NewEventRepository(postgresPool)
	authService := auth.NewService(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(userRepository, authService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentRepository)
	paymentHandler := handlers.NewPaymentHandler(appointmentRepository, paymentRepository, cfg.StripeSecretKey)
	stripeWebhookHandler := handlers.NewStripeWebhookHandler(paymentRepository, cfg.StripeWebhookSecret)
	reminderJob := jobs.NewReminderJob(appointmentRepository, eventRepository)
	cronScheduler := cron.New()
	if _, err := cronScheduler.AddFunc("@hourly", func() {
		if err := reminderJob.Run(context.Background()); err != nil {
			log.Printf("reminder job failed: %v", err)
		}
	}); err != nil {
		log.Fatalf("schedule reminder job: %v", err)
	}
	cronScheduler.Start()
	defer func() {
		ctx := cronScheduler.Stop()
		<-ctx.Done()
	}()

	router := gin.Default()
	router.POST("/auth/signup", authHandler.Signup)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/api/webhooks/stripe", stripeWebhookHandler.Handle)
	api := router.Group("/api")
	api.Use(middleware.JWT(cfg.JWTSecret))
	{
		api.GET("/me", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetInt64("user_id"),
				"role":    c.GetString("role"),
			})
		})
		appointments := api.Group("/appointments")
		{
			appointments.POST("", appointmentHandler.Create)
			appointments.GET("", appointmentHandler.List)
			appointments.GET("/:id", appointmentHandler.Get)
			appointments.PATCH("/:id", appointmentHandler.Patch)
			appointments.DELETE("/:id", appointmentHandler.Delete)
		}
		payments := api.Group("/payments")
		{
			payments.POST("/intent", paymentHandler.CreateIntent)
		}
	}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: router,
	}

	log.Printf("listening on %s", cfg.ServerAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("run server: %v", err)
	}
}
