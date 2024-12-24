package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/HikLk/go_final_API/db"
	"github.com/HikLk/go_final_API/utils"
)

const DateFormat = "20060102"

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// NextDateHandler calculates the next occurrence of a task based on the repetition rule
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.URL.Query().Get("now")
	repeatParam := r.URL.Query().Get("repeat")
	dateParam := r.URL.Query().Get("date")

	now, err := time.Parse(DateFormat, nowParam)
	if err != nil {
		http.Error(w, "Invalid 'now' date format, must be YYYYMMDD", http.StatusBadRequest)
		return
	}

	nextDate, err := utils.NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(nextDate)); err != nil {
		log.Printf("Response writing error: %v", err)
	}
}

// AddTaskHandler handles the addition of new tasks
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
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
			if task.Repeat == "" {
				task.Date = today
			} else {
				nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
				if err != nil {
					sendErrorResponse(w, "Error in repetition rule")
					return
				}
				task.Date = nextDate
			}
		}
	}

	res, err := db.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
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

// GetTasksHandler retrieves all tasks
func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC")
	if err != nil {
		sendErrorResponse(w, "Error retrieving tasks from database")
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			sendErrorResponse(w, "Error reading tasks from database")
			return
		}
		tasks = append(tasks, task)
	}

	response := map[string]interface{}{"tasks": tasks}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

// UpdateTaskHandler updates an existing task
func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
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
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				sendErrorResponse(w, "Error in repetition rule")
				return
			}
			task.Date = nextDate
		}
	}

	res, err := db.DB.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		sendErrorResponse(w, "Error updating task")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		sendErrorResponse(w, "Task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTaskHandler deletes a task
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Task ID is not specified")
		return
	}

	res, err := db.DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		sendErrorResponse(w, "Error deleting task")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		sendErrorResponse(w, "Task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// sendErrorResponse sends a JSON error response
func sendErrorResponse(w http.ResponseWriter, errorMessage string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}
