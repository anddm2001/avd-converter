package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/anddm2001/avd-converter/internal/config"
)

// RootCmd — корневая команда.
// Если пользователь запустил без подкоманд => вызываем логику info.
func RootCmd(mainLogger, infoLogger *zap.Logger, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adv-videoconverter",
		Short: "CLI утилита для конвертации видеофайлов (обертка над ffmpeg).",
		Long:  `adv-videoconverter — CLI утилита, позволяющая пакетно конвертировать видео с помощью ffmpeg.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Дефолтная команда => пишем инфо
			infoLogger.Info("No subcommand specified, running info by default...")
			// Можно просто вызвать функцию, которая «делает» info
			// или переиспользовать логический метод из info.go (если он экспортируемый).
			err := runInfoLogic(infoLogger, cfg)
			if err != nil {
				infoLogger.Error("Failed to run default info logic", zap.Error(err))
			}
			fmt.Println("Default (info) command finished. Check the info log for details.")
		},
	}

	return cmd
}
