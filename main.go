package main

import (
	"context"
	"database/sql"
	_ "github.com/ewik2k21/clicker_test/docs"
	"github.com/ewik2k21/clicker_test/internal/config"
	"github.com/ewik2k21/clicker_test/internal/handler"
	"github.com/ewik2k21/clicker_test/internal/repository"
	"github.com/ewik2k21/clicker_test/internal/service"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title clickerCountApi
// @version 1.0
// @description API server for count clicks
// @host localhost:8080
// @BasePath /
// @schemes http

// @in header
// @name Clicker
func main() {

	//init uber logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	//init cfg with conn str
	cfg := config.InitConfig()

	//connection to DB
	db, err := sql.Open("postgres", cfg.PostgresConnectionStr)
	if err != nil {
		logger.Fatal("Failed to connect to postgres", zap.Error(err))
	}

	//ping db
	if err = db.Ping(); err != nil {
		logger.Fatal("Failed to ping database")
	}

	//migrations
	goose.SetLogger(zap.NewStdLog(logger))
	if err = goose.Up(db, "migrations"); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
	}

	//layers init
	repo := repository.NewRepository(db, logger)
	newService := service.NewService(repo, logger)
	newHandler := handler.NewHandler(newService, logger)

	//gin setup
	g := gin.Default()
	g.POST("/counter/:bannerID", newHandler.PostCountClick)
	g.GET("/stats/:bannerID", newHandler.GetStatsForBanner)
	g.GET("/banners/random", newHandler.GetRandomBanners)
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//server cfg
	server := &http.Server{
		Addr:    cfg.HttpPort,
		Handler: g,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go newService.ProcessClicks(ctx)

	//graceful shutdown
	go func() {
		if err = server.ListenAndServe(); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	//signal... to shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	//request cancel
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	db.Close()
	logger.Info("Server exited")
}
