package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"negative-ion-respirator/backend/internal/config"
	"negative-ion-respirator/backend/internal/handler"
	"negative-ion-respirator/backend/internal/middleware"
	"negative-ion-respirator/backend/internal/mqtt"
	"negative-ion-respirator/backend/internal/repository"
	"negative-ion-respirator/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Repos
	deviceRepo := repository.NewDeviceRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	userRepo := repository.NewUserRepo(db)
	telemetryRepo := repository.NewTelemetryRepo(db)
	adminRepo := repository.NewAdminRepo(db)

	// MQTT
	mqttClient, err := mqtt.NewClient(cfg.EMQXHost, cfg.EMQXClientID, deviceRepo, telemetryRepo, orderRepo)
	if err != nil {
		log.Printf("WARNING: MQTT connection failed: %v (continuing without MQTT)", err)
	}
	if mqttClient != nil {
		defer mqttClient.Close()
	}

	// Services
	deviceSvc := service.NewDeviceService(deviceRepo, mqttClient)
	orderSvc := service.NewOrderService(orderRepo, userRepo, deviceSvc)
	authSvc := service.NewAuthService(adminRepo, cfg.JWTSecret)
	batchSvc := service.NewBatchService(deviceRepo, deviceSvc)
	reportSvc := service.NewReportService(deviceRepo, orderRepo, telemetryRepo)

	// Handlers
	deviceH := handler.NewDeviceHandler(deviceSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	authH := handler.NewAuthHandler(authSvc)
	adminH := handler.NewAdminHandler(deviceSvc)
	batchH := handler.NewBatchHandler(batchSvc, reportSvc)

	// Routes
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	{
		api.POST("/order/create", orderH.Create)
		api.GET("/order/query", orderH.Query)

		api.POST("/device/start", orderH.Start)
		api.POST("/device/stop", orderH.Stop)
		api.GET("/device/status/:id", deviceH.GetDevice)

		api.POST("/auth/login", authH.Login)

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(authSvc))
		{
			admin.GET("/devices", deviceH.ListDevices)
			admin.GET("/device/:id", adminH.GetDeviceStatus)
			admin.POST("/device/register", deviceH.Register)
			admin.PUT("/device/config", adminH.UpdateDeviceConfig)

			// Batch config
			admin.POST("/batch/config", batchH.CreateBatchConfig)
			admin.GET("/batch/task/:id", batchH.GetTaskStatus)

			// Reports
			admin.GET("/dashboard", batchH.GetDashboard)
			admin.GET("/report", batchH.GetReport)
		}
	}

	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: r}

	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
