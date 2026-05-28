package main

import (
	"akatengu/internal/bootstrap"
	"akatengu/internal/enums"
	"akatengu/internal/handler"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 0. register enums
	enums.InitEnums()

	// 1. Config
	cfg := bootstrap.LoadConfig()

	// 2. OpenTelemetry
	//shutdown := bootstrap.InitOTel(cfg.OTel)
	//defer shutdown()

	// 3. logger
	logger := bootstrap.InitLogger()

	// 3. DB
	db, err := bootstrap.InitDB(cfg.DB, logger)
	if err != nil {
		logger.Fatal("init db fail", zap.Error(err))
	}
	defer db.Close()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler.NewMux(db, cfg, logger),
	}

	go func() {
		logger.Info("server started at :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("listen fail", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
