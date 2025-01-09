package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/anddm2001/avd-converter/internal/config"
)

// InfoCmd — отдельная команда, чтобы вывести (и залогировать) подробную инфу
// по всем файлам в каталоге импорта.
func InfoCmd(infoLogger *zap.Logger, cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Выводит подробную информацию по каждому файлу в каталоге импорта",
		Long: `Команда "info" обходит каталог импорта (IMPORT_DIR) 
и пишет в особый лог-файл (INFO_LOG_FILE) детальные сведения (имя файла, размер и т.п.).`,
		Run: func(cmd *cobra.Command, args []string) {
			infoLogger.Info("Running info command...")
			err := runInfoLogic(infoLogger, cfg)
			if err != nil {
				infoLogger.Error("Failed to run info logic", zap.Error(err))
			}
			fmt.Println("Info command finished. Check the info log for details.")
		},
	}

	return cmd
}

// runInfoLogic — общая функция логики, используемая и в default-команде, и в команде info.
func runInfoLogic(logger *zap.Logger, cfg *config.Config) error {
	return filepath.Walk(cfg.ImportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error("Error walking directory", zap.String("path", path), zap.Error(err))
			return nil
		}
		if !info.IsDir() {
			logger.Info("File found",
				zap.String("file", info.Name()),
				zap.Int64("size", info.Size()),
			)
		}
		return nil
	})
}
