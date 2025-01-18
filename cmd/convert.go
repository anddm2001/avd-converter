package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/anddm2001/avd-converter/internal/config"
	"github.com/anddm2001/avd-converter/pkg/monitor"
)

// ConvertCmd возвращает команду для различных преобразований видео.
func ConvertCmd(logger *zap.Logger, cfg *config.Config, maxCPUTemp, maxGPUTemp float64) *cobra.Command {
	var (
		bitrate      string
		codec        string
		container    string
		removeMeta   bool
		removeAudio  bool
		extractAudio bool
		resolution   string
		orientation  string
		fileMask     string
	)

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Конвертация и пакетная обработка видеофайлов",
		Long: `Позволяет уменьшать битрейт, менять кодек, контейнер,
удалять метаданные, аудио-дорожку, а также сохранять аудио отдельно, 
менять разрешение и ориентацию, 
причем все операции могут выполняться пакетно и комбинированно.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("Starting convert command...")

			// Инициализируем контекст
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Запускаем мониторинг
			monitor.StartTemperatureMonitor(ctx, logger, maxCPUTemp, maxGPUTemp, cancel, 10*time.Second)

			// Получаем список файлов из каталога импорта
			var files []string
			filepath.Walk(cfg.ImportDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() {
					files = append(files, path)
				}
				return nil
			})

			for _, inFilePath := range files {
				select {
				case <-ctx.Done():
					// Значит, температура превысила лимит или что-то ещё отменило контекст
					logger.Error("Operation canceled due to high temperature or external cancel",
						zap.String("file", inFilePath))
					return
				default:
					// Продолжаем
				}

				ext := filepath.Ext(inFilePath)
				base := strings.TrimSuffix(filepath.Base(inFilePath), ext)

				// Формируем выходной файл (можем использовать маску, например: {originalName}_converted
				// или берём из флага `fileMask`)
				var outFileName string
				if fileMask != "" {
					// Пример: fileMask = "{name}_myconvert"
					// Заменим {name} на реальное имя файла без расширения
					outFileName = strings.ReplaceAll(fileMask, "{name}", base)
				} else {
					outFileName = base + "_converted"
				}

				// Если пользователь указал container, формируем новое расширение
				// иначе используем исходное
				var outExt string
				if container != "" {
					outExt = "." + container
				} else {
					outExt = ext
				}

				outFilePath := filepath.Join(cfg.ExportDir, outFileName+outExt)

				// Подготавливаем аргументы ffmpeg
				ffmpegArgs := []string{
					"-i", inFilePath,
				}

				// Удаление аудио дорожки
				if removeAudio {
					ffmpegArgs = append(ffmpegArgs, "-an")
				}

				// Извлечение аудио в отдельный файл
				if extractAudio && !removeAudio {
					// Сохраним аудио дорожку в тот же каталог экспорта
					audioOutPath := filepath.Join(cfg.ExportDir, base+"_audio.aac")
					audioArgs := []string{
						"-i", inFilePath,
						"-vn", // без видео
						"-c:a", "copy",
						audioOutPath,
					}
					// Запустим отдельный ffmpeg-процесс для извлечения аудио
					if err := runFFmpeg(audioArgs, logger); err != nil {
						logger.Error("Failed to extract audio",
							zap.String("file", inFilePath),
							zap.Error(err))
					}
				}

				// Уменьшение битрейта
				if bitrate != "" {
					ffmpegArgs = append(ffmpegArgs, "-b:v", bitrate)
				}

				// Перекодировка (указание нового кодека)
				if codec != "" {
					ffmpegArgs = append(ffmpegArgs, "-c:v", codec)
				}

				// Удаление метаданных
				if removeMeta {
					ffmpegArgs = append(ffmpegArgs, "-map_metadata", "-1")
				}

				// Изменение разрешения
				if resolution != "" {
					ffmpegArgs = append(ffmpegArgs, "-vf", fmt.Sprintf("scale=%s", resolution))
				}

				// Изменение ориентации (поворот)
				// Например, orientation = "90" (повернуть на 90 градусов)
				// В ffmpeg фильтры поворота: transpose=1 (90 по часовой)
				// Можно расширить логику.
				if orientation != "" {
					transpose := "1" // default
					if orientation == "90" {
						transpose = "1"
					} else if orientation == "180" {
						// Повернуть на 180 — это flip+flip или transpose=2 + transpose=2,
						// проще воспользоваться фильтром rotation
						// orientation=180 => "-vf rotate=PI"
						ffmpegArgs = append(ffmpegArgs, "-vf", "rotate=PI")
					} else if orientation == "270" {
						transpose = "2"
					}
					if orientation == "90" || orientation == "270" {
						ffmpegArgs = append(ffmpegArgs, "-vf", "transpose="+transpose)
					}
				}

				ffmpegArgs = append(ffmpegArgs, outFilePath)

				// Выполняем ffmpeg
				if err := runFFmpeg(ffmpegArgs, logger); err != nil {
					logger.Error("Conversion failed",
						zap.String("file", inFilePath),
						zap.Error(err),
					)
					continue
				}

				// По заданию сказано «сохранение оригиналов», т.е. удалять оригинал мы не должны.
				// Если нужно удалять, то можно раскомментировать или добавить флаг.

				logger.Info("File converted successfully",
					zap.String("input", inFilePath),
					zap.String("output", outFilePath),
				)
			}
			logger.Info("Convert command finished.")
		},
	}

	cmd.Flags().StringVar(&bitrate, "bitrate", "", "Уменьшить битрейт (например, 800k)")
	cmd.Flags().StringVar(&codec, "codec", "", "Указать другой видео-кодек (например, libx264)")
	cmd.Flags().StringVar(&container, "container", "", "Сменить контейнер (mp4, mkv и т.д.)")
	cmd.Flags().BoolVar(&removeMeta, "remove-meta", false, "Удалить метаданные")
	cmd.Flags().BoolVar(&removeAudio, "remove-audio", false, "Удалить аудио-дорожку")
	cmd.Flags().BoolVar(&extractAudio, "extract-audio", false, "Сохранить аудио в отдельный файл")
	cmd.Flags().StringVar(&resolution, "resolution", "", "Изменить разрешение, например 1280:720")
	cmd.Flags().StringVar(&orientation, "orientation", "", "Повернуть видео (90, 180, 270)")
	cmd.Flags().StringVar(&fileMask, "file-mask", "", "Маска для имен обработанных файлов (использовать {name} для имени файла)")

	return cmd
}

func runFFmpeg(args []string, logger *zap.Logger) error {
	logger.Info("Running ffmpeg", zap.Strings("args", args))
	cmd := exec.Command("ffmpeg", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("ffmpeg error", zap.Error(err), zap.String("output", string(out)))
		return err
	}
	logger.Info("ffmpeg success", zap.String("output", string(out)))
	return nil
}
