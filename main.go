package main

import (
	"log"
	"net/http"
	"os"

	"github.com/HikLk/go_final_API/db"
	"github.com/HikLk/go_final_API/handlers"
)

func main() {
	// Инициализация базы данных
	_, err := db.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.DB.Close()

	// Настройка маршрутов
	http.HandleFunc("/api/next-date", handlers.NextDateHandler)
	http.HandleFunc("/api/add-task", handlers.AddTaskHandler)
	http.HandleFunc("/api/get-tasks", handlers.GetTasksHandler)
	http.HandleFunc("/api/update-task", handlers.UpdateTaskHandler)
	http.HandleFunc("/api/delete-task", handlers.DeleteTaskHandler)

	// Чтение порта из окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
