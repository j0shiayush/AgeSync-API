package main

import (
	"errors"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"AgeSync-API/config"
	db "AgeSync-API/db/sqlc"
	"AgeSync-API/internal/handler"
	"AgeSync-API/internal/logger"
	"AgeSync-API/internal/repository"
	"AgeSync-API/internal/routes"
	"AgeSync-API/internal/service"
)

func main() {
	
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: load config: %v\n", err)
		os.Exit(1)
	}

	
	log, err := logger.New(cfg.App.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: init logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() 

	log.Info("starting userapi",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.Port),
	)

	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatal("failed to create database pool", zap.Error(err))
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		log.Fatal("failed to ping database", zap.Error(err), zap.String("dsn", cfg.DB.DSN()))
	}
	log.Info("database connection established")

	
	queries := db.New(pool)
	userRepo := repository.NewUserRepository(queries)
	userSvc := service.NewUserService(userRepo, log)
	userHandler := handler.NewUserHandler(userSvc, log)

	
	app := fiber.New(fiber.Config{
		
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if ok := errors.As(err, &fe); ok {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.App.Env == "development",
	}))

	routes.Register(app, userHandler, log)

	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		log.Info("HTTP server listening", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Error("HTTP server error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutdown signal received, draining connections...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()

	if err := app.ShutdownWithContext(shutCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}

	log.Info("server shut down cleanly")
}