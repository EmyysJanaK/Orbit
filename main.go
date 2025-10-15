package main

import (
	"context"
	"log"
	"net/http"

	"employee-management-system/auth"
	"employee-management-system/config"
	"employee-management-system/handlers"
	"employee-management-system/middleware"
	"employee-management-system/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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
	authService := auth.NewService(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(userRepository, authService)

	router := gin.Default()
	router.POST("/auth/signup", authHandler.Signup)
	router.POST("/auth/login", authHandler.Login)
	authorized := router.Group("/")
	authorized.Use(middleware.JWT(cfg.JWTSecret))
	authorized.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.GetInt64("user_id"),
			"role":    c.GetString("role"),
		})
	})
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
