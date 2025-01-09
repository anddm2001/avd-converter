package main

import (
	"log"

	"github.com/anddm2001/avd-converter/cmd"
	"github.com/anddm2001/avd-converter/internal/config"
	"github.com/anddm2001/avd-converter/pkg/logger"
)

func main() {
	// 1. Загружаем конфиг через Viper
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Инициализируем «основной» логгер
	mainLogger, err := logger.InitLogger(cfg.LogFilePath)
	if err != nil {
		log.Fatalf("Error initializing main logger: %v", err)
	}
	defer mainLogger.Sync()

	// 3. Инициализируем «info-логгер» (в особый лог-файл)
	infoLogger, err := logger.InitLogger(cfg.InfoLogFilePath)
	if err != nil {
		log.Fatalf("Error initializing info logger: %v", err)
	}
	defer infoLogger.Sync()

	// 4. Собираем корневую команду и регистрируем подкоманды
	rootCmd := cmd.RootCmd(mainLogger, infoLogger, cfg)
	rootCmd.AddCommand(
		cmd.InfoCmd(infoLogger, cfg),    // команда info
		cmd.ConvertCmd(mainLogger, cfg), // команда convert
	)

	// 5. Запускаем
	if err := rootCmd.Execute(); err != nil {
		mainLogger.Sugar().Fatalf("Error executing rootCmd: %v", err)
	}
}
