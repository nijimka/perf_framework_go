package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// 1. Эндпоинт логина: выдает токен
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": "super-secret-token-123",
			"user":  "test_user",
		})
	})

	// 2. Эндпоинт данных: требует токен
	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Проверяем токен
		if !strings.Contains(authHeader, "super-secret-token-123") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		delay := time.Duration(50+rand.Intn(150)) * time.Millisecond
		time.Sleep(delay)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"delay_ms": delay.Milliseconds(),
			"data":     "some secret data",
		})
	})

	port := ":8080"
	log.Printf("Mock server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
