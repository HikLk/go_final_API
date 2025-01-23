package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Инициализация базы данных
	_, err := InitializeDatabase()
	if err != nil {
		log.Fatalf("Не удалось инициализировать базу данных: %v", err)
	}
	defer DB.Close() // Закрытие соединения с базой данных при завершении работы программы

	// Настройка маршрутов
	http.HandleFunc("/api/next-date", NextDateHandler)         // Обработчик для получения следующей даты
	http.HandleFunc("/api/add-task", AddTaskHandler)           // Обработчик для добавления задачи
	http.HandleFunc("/api/get-tasks", GetTasksHandler)         // Обработчик для получения списка задач
	http.HandleFunc("/api/update-task", UpdateTaskHandler)     // Обработчик для обновления задачи
	http.HandleFunc("/api/get-task-by-id", GetTaskByIDHandler) // Обработчик для удаления задачи
	http.HandleFunc("/api/delete-task", DeleteTaskHandler)     // Обработчик для удаления задачи

	// Чтение порта из переменных окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Использовать порт 7540, если переменная окружения PORT не задана
	}

	log.Printf("Запуск сервера на порту %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil)) // Запуск HTTP-сервера
}
