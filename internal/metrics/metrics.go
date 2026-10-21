package metrics

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Metrics struct {
	mu              sync.Mutex
	TotalRequests   int
	SuccessCount    int
	ErrorCount      int
	TotalDuration   time.Duration
	ResponseTimes   []time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	TestStartTime   time.Time
	TestEndTime     time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		ResponseTimes: make([]time.Duration, 0),
	}
}

func (m *Metrics) StartTimer() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TestStartTime = time.Now()
}

func (m *Metrics) StopTimer() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TestEndTime = time.Now()
}

func (m *Metrics) Record(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	if err != nil {
		m.ErrorCount++
	} else {
		m.SuccessCount++
	}

	m.TotalDuration += duration
	m.ResponseTimes = append(m.ResponseTimes, duration)

	if m.MinResponseTime == 0 || duration < m.MinResponseTime {
		m.MinResponseTime = duration
	}
	if duration > m.MaxResponseTime {
		m.MaxResponseTime = duration
	}
}

// ReportData содержит данные, готовые для HTML-шаблона
type ReportData struct {
	BaseURL         string
	TotalRequests   int
	SuccessCount    int
	ErrorCount      int
	RPS             float64
	AvgResponseTime string
	MinResponseTime string
	MaxResponseTime string
	P95ResponseTime string
	TotalTestTime   string
	GeneratedAt     string
}

func (m *Metrics) GetReportData(baseURL string) ReportData {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := ReportData{
		BaseURL:       baseURL,
		TotalRequests: m.TotalRequests,
		SuccessCount:  m.SuccessCount,
		ErrorCount:    m.ErrorCount,
		GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	if m.TotalRequests > 0 {
		data.AvgResponseTime = (m.TotalDuration / time.Duration(m.TotalRequests)).String()
		data.MinResponseTime = m.MinResponseTime.String()
		data.MaxResponseTime = m.MaxResponseTime.String()

		if len(m.ResponseTimes) > 0 {
			sort.Slice(m.ResponseTimes, func(i, j int) bool {
				return m.ResponseTimes[i] < m.ResponseTimes[j]
			})
			p95Index := int(float64(len(m.ResponseTimes)) * 0.95)
			if p95Index >= len(m.ResponseTimes) {
				p95Index = len(m.ResponseTimes) - 1
			}
			data.P95ResponseTime = m.ResponseTimes[p95Index].String()
		}

		if !m.TestStartTime.IsZero() && !m.TestEndTime.IsZero() {
			totalTestTime := m.TestEndTime.Sub(m.TestStartTime)
			data.TotalTestTime = totalTestTime.String()
			data.RPS = float64(m.TotalRequests) / totalTestTime.Seconds()
		}
	}

	return data
}

func (m *Metrics) Print() {
	data := m.GetReportData("")
	fmt.Println("\n========== Метрики ==========")
	fmt.Printf("Всего запросов: %d\n", data.TotalRequests)
	fmt.Printf("Успешных: %d\n", data.SuccessCount)
	fmt.Printf("Ошибок: %d\n", data.ErrorCount)
	fmt.Printf("Среднее время ответа: %s\n", data.AvgResponseTime)
	fmt.Printf("Минимальное время: %s\n", data.MinResponseTime)
	fmt.Printf("Максимальное время: %s\n", data.MaxResponseTime)
	fmt.Printf("95-й перцентиль: %s\n", data.P95ResponseTime)
	fmt.Printf("\nОбщее время теста: %s\n", data.TotalTestTime)
	fmt.Printf("RPS (запросов в секунду): %.2f\n", data.RPS)
	fmt.Println("=============================\n")
}
