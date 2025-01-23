package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const limit = 50

// Task представляет задачу
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// nextDateHandler обрабатывает запрос для расчета следующей даты задачи
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.URL.Query().Get("now")
	repeatParam := r.URL.Query().Get("repeat")
	dateParam := r.URL.Query().Get("date")

	now, err := time.Parse(DateFormat, nowParam)
	if err != nil {
		http.Error(w, "Invalid 'now' date format, must be YYYYMMDD", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	if _, err := w.Write([]byte(nextDate)); err != nil {
		log.Printf("Response writing error: %v", err)
	}
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "JSON Deserialization Error", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendErrorResponse(w, "Task title is not specified")
		return
	}

	now := time.Now()
	today := now.Format(DateFormat)

	if task.Date == "" {
		task.Date = today
	} else {
		parsedDate, err := time.Parse(DateFormat, task.Date)
		if err != nil {
			sendErrorResponse(w, "Date is in an incorrect format")
			return
		}
		if parsedDate.Before(now) {
			if task.Repeat != "" {
				nextDate, err := NextDate(now, task.Date, task.Repeat)
				if err != nil {
					sendErrorResponse(w, "Error in repetition rule")
					return
				}
				task.Date = nextDate
			} else {
				task.Date = today
			}
		}
	}

	if task.Repeat != "" {
		if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
			sendErrorResponse(w, "Invalid repetition rule")
			return
		}
	}

	res, err := DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		sendErrorResponse(w, "Error adding task to the database")
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		sendErrorResponse(w, "Error retrieving task ID")
		return
	}

	response := map[string]string{"id": fmt.Sprintf("%d", id)}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(fmt.Sprintf("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT %d", limit))
	if err != nil {
		log.Printf("Database query error: %v", err)
		sendErrorResponse(w, "Error retrieving tasks from database")
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			log.Printf("Error scanning row: %v", err)
			sendErrorResponse(w, "Error reading tasks from database")
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		sendErrorResponse(w, "Error iterating tasks")
		return
	}

	response := map[string]interface{}{"tasks": tasks}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSON encoding error: %v", err)
		sendErrorResponse(w, "Error encoding tasks to JSON")
	}
}

func GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Task ID is not specified")
		return
	}

	var task Task
	if err := DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "Task not found")
		} else {
			log.Printf("Database error: %v", err)
			sendErrorResponse(w, "Error retrieving task")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("JSON encoding error: %v", err)
		sendErrorResponse(w, "Error encoding task to JSON")
	}
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "JSON Deserialization Error", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		sendErrorResponse(w, "Task ID is not specified")
		return
	}
	if task.Title == "" {
		sendErrorResponse(w, "Task title is not specified")
		return
	}

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	} else {
		parsedDate, err := time.Parse(DateFormat, task.Date)
		if err != nil {
			sendErrorResponse(w, "Date is in an incorrect format")
			return
		}
		if parsedDate.Before(now) && task.Repeat != "" {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				sendErrorResponse(w, "Error in repetition rule")
				return
			}
			task.Date = nextDate
		}
	}

	res, err := DB.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("Database error: %v", err)
		sendErrorResponse(w, "Error updating task")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		sendErrorResponse(w, "Задача не найдена")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// sendErrorResponse отправляет JSON-ответ с ошибкой
func sendErrorResponse(w http.ResponseWriter, errorMessage string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получение идентификатора задачи
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Task ID is not specified")
		return
	}

	// Выполнение запроса на удаление и получение количества затронутых строк
	res, err := DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		log.Printf("Error deleting task: %v", err)
		sendErrorResponse(w, "Error deleting task")
		return
	}

	// Проверяем, была ли удалена хотя бы одна строка
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error retrieving affected rows: %v", err)
		sendErrorResponse(w, "Error retrieving affected rows")
		return
	}

	if rowsAffected == 0 {
		// Если строк с таким id не найдено, возвращаем ошибку
		sendErrorResponse(w, "Task not found")
		return
	}

	// Возвращаем пустой JSON-ответ в случае успешного удаления
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}
