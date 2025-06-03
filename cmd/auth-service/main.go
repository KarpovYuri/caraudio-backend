// cmd/auth-service/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Простая маршрутизация для примера
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Auth Service!")
	})

	// Запуск HTTP-сервера
	port := ":8080"
	log.Printf("Auth Service starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
