package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config — структура для хранения необходимых параметров.
type Config struct {
	ImportDir       string
	ExportDir       string
	LogFilePath     string
	InfoLogFilePath string
}

// LoadConfig — загрузка конфигурации с помощью Viper.
func LoadConfig(envFile string) (*Config, error) {
	v := viper.New()

	// Указываем файл, который будем читать
	v.SetConfigFile(envFile)
	v.SetConfigType("env") // для формата KEY=VAL

	// Разрешаем автоматическое чтение ENV
	v.AutomaticEnv()

	// Значения по умолчанию
	v.SetDefault("IMPORT_DIR", "./input")
	v.SetDefault("EXPORT_DIR", "./output")
	v.SetDefault("LOG_FILE", "./logs/videoconverter.log")
	v.SetDefault("INFO_LOG_FILE", "./logs/info.log")

	// Пробуем прочитать .env
	if err := v.ReadInConfig(); err != nil {
		// Можно вывести предупреждение, но не обязательно падать
		fmt.Printf("Warning: could not read config file %s: %v\n", envFile, err)
	}

	// Собираем конфиг
	cfg := &Config{
		ImportDir:       v.GetString("IMPORT_DIR"),
		ExportDir:       v.GetString("EXPORT_DIR"),
		LogFilePath:     v.GetString("LOG_FILE"),
		InfoLogFilePath: v.GetString("INFO_LOG_FILE"),
	}

	return cfg, nil
}
