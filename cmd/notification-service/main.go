package main

import (
	"notification-service/internal/app"
	"notification-service/internal/config"
	"notification-service/internal/lib/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.MustSetupLogger(cfg.Env)

	app.NewApp(log, cfg)

	select {}
}
