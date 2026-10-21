package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Response struct {
	StatusCode int
	Body       string
	Duration   time.Duration
	Error      error
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

func (c *Client) DoRequest(method, path string, headers map[string]string) Response {
	start := time.Now()

	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return Response{Duration: time.Since(start), Error: err}
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		return Response{Duration: duration, Error: err}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Считаем не-2xx ответы ошибками
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return Response{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Duration:   duration,
		Error:      err,
	}
}

// Правильный парсинг JSON вместо поиска подстрок
func ExtractToken(body string) string {
	var result map[string]string
	// Пытаемся распаковать JSON в карту
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return "" // Если не JSON или ошибка - возвращаем пусто
	}
	// Достаём токен по ключу
	return result["token"]
}
