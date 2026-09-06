package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func main() {
	r := chi.NewRouter()
	r.Get("/health", handler.Health)

	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
