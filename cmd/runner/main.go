package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"perf-framework/internal/client"
	"perf-framework/internal/config"
	"perf-framework/internal/metrics"
)

//go:embed report.html.tmpl
var reportTemplateFS embed.FS

func main() {
	configPath := flag.String("config", "config.yaml", "Путь к конфигу")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	fmt.Printf("Запуск тестов для: %s\n", cfg.Target.BaseURL)
	fmt.Printf("Профиль нагрузки: %s\n", cfg.LoadProfile.Type)
	if cfg.LoadProfile.Type == "ramp-up" {
		fmt.Printf("Нарастание до %d воркеров за %d секунд\n",
			cfg.LoadProfile.TargetWorkers, cfg.LoadProfile.RampUpDurationSec)
	}

	timeout := time.Duration(cfg.Target.TimeoutMs) * time.Millisecond
	httpClient := client.NewClient(cfg.Target.BaseURL, timeout)
	m := metrics.NewMetrics()

	m.StartTimer()

	for _, scenario := range cfg.Scenarios {
		fmt.Printf("\n=== Сценарий: %s (Итераций: %d) ===\n",
			scenario.Name, scenario.Iterations)

		var wg sync.WaitGroup

		// Запускаем воркеров в зависимости от профиля нагрузки
		if cfg.LoadProfile.Type == "ramp-up" {
			runRampUp(&wg, cfg, scenario, httpClient, m)
		} else {
			runConstant(&wg, cfg, scenario, httpClient, m)
		}

		wg.Wait()
	}

	m.StopTimer()
	m.Print()
	generateReport(cfg.Target.BaseURL, m)
}

// Constant: все воркеры стартуют одновременно
func runConstant(wg *sync.WaitGroup, cfg *config.Config, scenario config.ScenarioConfig,
	httpClient *client.Client, m *metrics.Metrics) {

	fmt.Printf("Запуск %d воркеров одновременно...\n", cfg.LoadProfile.TargetWorkers)

	for w := 0; w < cfg.LoadProfile.TargetWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(workerID, scenario, httpClient, m)
		}(w)
	}
}

// Ramp-up: воркеры запускаются постепенно
func runRampUp(wg *sync.WaitGroup, cfg *config.Config, scenario config.ScenarioConfig,
	httpClient *client.Client, m *metrics.Metrics) {

	targetWorkers := cfg.LoadProfile.TargetWorkers
	rampUpDuration := time.Duration(cfg.LoadProfile.RampUpDurationSec) * time.Second

	// Вычисляем интервал между запусками воркеров
	interval := rampUpDuration / time.Duration(targetWorkers)

	fmt.Printf("Запуск %d воркеров с интервалом %v...\n", targetWorkers, interval)

	for w := 0; w < targetWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(workerID, scenario, httpClient, m)
		}(w)

		// Ждем перед запуском следующего воркера (кроме последнего)
		if w < targetWorkers-1 {
			time.Sleep(interval)
			fmt.Printf("[%v] Запущен воркер #%d\n", time.Now().Format("15:04:05"), w+2)
		}
	}
}

// Общая функция для выполнения работы воркера
func runWorker(workerID int, scenario config.ScenarioConfig,
	httpClient *client.Client, m *metrics.Metrics) {

	var userToken string

	for i := 0; i < scenario.Iterations; i++ {
		for _, step := range scenario.Steps {
			headers := make(map[string]string)

			if step.UseToken && userToken != "" {
				headers["Authorization"] = "Bearer " + userToken
			}

			resp := httpClient.DoRequest(step.Method, step.Path, headers)
			m.Record(resp.Duration, resp.Error)

			if step.SaveToken && resp.StatusCode == 200 {
				token := client.ExtractToken(resp.Body)
				if token != "" {
					userToken = token
				}
			}
		}
	}
}

func generateReport(baseURL string, m *metrics.Metrics) {
	reportDir := "/app/reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		log.Printf("Не удалось создать директорию %s: %v", reportDir, err)
		return
	}

	tmplContent, err := reportTemplateFS.ReadFile("report.html.tmpl")
	if err != nil {
		log.Printf("Не удалось прочитать шаблон: %v", err)
		return
	}

	tmpl, err := template.New("report").Parse(string(tmplContent))
	if err != nil {
		log.Printf("Ошибка парсинга шаблона: %v", err)
		return
	}

	fileName := fmt.Sprintf("report_%s.html", time.Now().Format("20060102_150405"))
	filePath := filepath.Join(reportDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		log.Printf("Не удалось создать файл отчета: %v", err)
		return
	}
	defer file.Close()

	data := m.GetReportData(baseURL)
	if err := tmpl.Execute(file, data); err != nil {
		log.Printf("Ошибка генерации отчета: %v", err)
		return
	}

	fmt.Printf("✅ HTML-отчет успешно сохранен: %s\n", filePath)
}
