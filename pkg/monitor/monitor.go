package monitor

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// StartTemperatureMonitor запускает горутину, которая периодически проверяет температуру CPU/GPU.
// Если обнаруживает, что температура выше переданных лимитов, вызывает cancel() и прекращает мониторинг.
func StartTemperatureMonitor(
	ctx context.Context,
	logger *zap.Logger,
	maxCPUTemp float64,
	maxGPUTemp float64,
	cancel context.CancelFunc,
	interval time.Duration,
) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Контекст отменён снаружи — выходим
				return
			case <-ticker.C:
				cpuTemp, gpuTemp, err := getTemperatures(logger)
				if err != nil {
					// Можно залогировать ошибку парсинга, но не останавливать
					logger.Warn("Failed to get temperatures", zap.Error(err))
					continue
				}

				logger.Info("Temperature check",
					zap.Float64("cpu_temp", cpuTemp),
					zap.Float64("gpu_temp", gpuTemp),
				)

				// Сравниваем с лимитами
				if cpuTemp > maxCPUTemp || gpuTemp > maxGPUTemp {
					logger.Error("Temperature limit exceeded!",
						zap.Float64("cpu_temp", cpuTemp),
						zap.Float64("gpu_temp", gpuTemp),
						zap.Float64("max_cpu", maxCPUTemp),
						zap.Float64("max_gpu", maxGPUTemp),
					)
					// Отменяем контекст => останавливаем кодировку / операцию
					cancel()
					return
				}
			}
		}
	}()
}

// getTemperatures — пример вызова powermetrics и парсинга вывода.
func getTemperatures(logger *zap.Logger) (cpuTemp float64, gpuTemp float64, err error) {
	cmd := exec.Command("sudo", "powermetrics", "--samplers", "smc", "-i1", "-n1")

	var out bytes.Buffer
	cmd.Stdout = &out
	if errRun := cmd.Run(); errRun != nil {
		return 0, 0, errRun
	}

	reCPU := regexp.MustCompile(`CPU die temperature:\s+([\d\.]+)\s+C`)
	reGPU := regexp.MustCompile(`GPU die temperature:\s+([\d\.]+)\s+C`)

	strOut := out.String()

	var matchCPU = reCPU.FindStringSubmatch(strOut)
	var matchGPU = reGPU.FindStringSubmatch(strOut)

	if len(matchCPU) == 2 {
		cpuTemp, err = strconv.ParseFloat(matchCPU[1], 64)
		if err != nil {
			return 0, 0, err
		}
	}

	if len(matchGPU) == 2 {
		gpuTemp, err = strconv.ParseFloat(matchGPU[1], 64)
		if err != nil {
			return 0, 0, err
		}
	}

	return cpuTemp, gpuTemp, nil
}
