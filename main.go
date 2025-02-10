package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	// Инициализация базы данных
	var err error
	db, err = initializeDB()
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	// Получение порта из переменной окружения или значения по умолчанию
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Настройка маршрутов
	http.Handle("/", http.FileServer(http.Dir("./web")))   // Обслуживание статических файлов из каталога "./web"
	http.HandleFunc("/api/nextdate", nextDateHandler)      // Обработка API для получения следующей даты
	http.HandleFunc("/api/tasks", getTasksHandler)         // Обработка API для получения всех задач
	http.HandleFunc("/api/task/done", markTaskDoneHandler) // Обработка API для отметки задачи как выполненной
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		// Обработка запросов к конкретной задаче
		switch r.Method {
		case http.MethodGet:
			getTaskByIDHandler(w, r) // Получение задачи по ID
		case http.MethodPut:
			updateTaskHandler(w, r) // Обновление задачи
		case http.MethodPost:
			addTaskHandler(w, r) // Добавление новой задачи
		case http.MethodDelete:
			deleteTaskHandler(w, r) // Удаление задачи
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// Запуск сервера
	log.Printf("Сервер запущен на http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
