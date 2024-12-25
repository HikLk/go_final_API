package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// nextDateHandler обрабатывает запрос для расчета следующей даты задачи
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем параметры из запроса
	nowParam := r.URL.Query().Get("now")
	repeatParam := r.URL.Query().Get("repeat")
	dateParam := r.URL.Query().Get("date")

	// Преобразуем строку в дату
	now, err := time.Parse(DateFormat, nowParam)
	if err != nil {
		http.Error(w, "Неверный формат даты 'now', должен быть в формате YYYYMMDD", http.StatusBadRequest)
		return
	}

	// Вычисляем следующую дату задачи с учетом повторений
	nextDate, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем результат
	w.Write([]byte(nextDate))
}

// addTaskHandler добавляет новую задачу в базу данных
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendErrorResponse(w, "Необходимо указать название задачи")
		return
	}

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	} else {
		parsedDate, err := time.Parse(DateFormat, task.Date)
		if err != nil {
			sendErrorResponse(w, "Неверный формат даты")
			return
		}

		if parsedDate.Before(now) && task.Repeat != "" {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				sendErrorResponse(w, "Неверное правило повторения")
				return
			}
			task.Date = nextDate
		}
	}

	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		sendErrorResponse(w, "Ошибка при добавлении задачи в базу данных")
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		sendErrorResponse(w, "Ошибка при получении ID задачи")
		return
	}

	response := map[string]string{"id": fmt.Sprintf("%d", id)}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

// getTasksHandler получает список задач из базы данных
func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(fmt.Sprintf(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT %d`, limit))
	if err != nil {
		log.Printf("Ошибка при запросе к базе данных: %v", err)
		sendErrorResponse(w, "Ошибка при получении задач")
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			log.Printf("Ошибка при чтении строки: %v", err)
			sendErrorResponse(w, "Ошибка при чтении задач из базы данных")
			return
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при переборе строк: %v", err)
		sendErrorResponse(w, "Ошибка при переборе задач")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

// getTaskByIDHandler обрабатывает запрос на получение задачи по ID
func getTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Необходимо указать ID задачи")
		return
	}

	var task Task
	err := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "Задача не найдена")
		} else {
			log.Printf("Ошибка базы данных: %v", err)
			sendErrorResponse(w, "Ошибка при получении задачи")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(task)
}

// updateTaskHandler обновляет задачу в базе данных
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" || task.Title == "" {
		sendErrorResponse(w, "Необходимо указать ID и название задачи")
		return
	}

	_, err := db.Exec(`UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("Ошибка базы данных: %v", err)
		sendErrorResponse(w, "Ошибка при обновлении задачи")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{"message": "Задача успешно обновлена"})
}

// deleteTaskHandler удаляет задачу по ID
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Необходимо указать ID задачи")
		return
	}

	_, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		log.Printf("Ошибка базы данных: %v", err)
		sendErrorResponse(w, "Ошибка при удалении задачи")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{"message": "Задача успешно удалена"})
}

// sendErrorResponse отправляет ошибку с сообщением в формате JSON
func sendErrorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusBadRequest)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

// markTaskDoneHandler обрабатывает запросы для отметки задачи как выполненной и удаляет её
func markTaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "ID задачи не передан", http.StatusBadRequest)
		log.Println("Ошибка: ID задачи не передан.")
		return
	}

	// Логика для отметки задачи как выполненной
	err := markTaskDone(taskID)
	if err != nil {
		http.Error(w, "Не удалось обновить задачу", http.StatusInternalServerError)
		log.Printf("Ошибка при обновлении задачи с ID %s: %v\n", taskID, err)
		return
	}

	// Удаление задачи из базы данных после выполнения
	err = deleteTask(taskID)
	if err != nil {
		http.Error(w, "Не удалось удалить задачу", http.StatusInternalServerError)
		log.Printf("Ошибка при удалении задачи с ID %s: %v\n", taskID, err)
		return
	}

	// Ответ при успешном удалении задачи
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Задача выполнена и удалена"}`))
}

// markTaskDone обновляет статус задачи в базе данных и отмечает её как выполненную
func markTaskDone(taskID string) error {
	log.Printf("Отмечаем задачу с ID %s как выполненную.\n", taskID)

	query := "UPDATE scheduler SET status = ? WHERE id = ?"
	_, err := db.Exec(query, "done", taskID)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус задачи с ID %s как выполненную: %v", taskID, err)
	}

	log.Printf("Задача с ID %s успешно отмечена как выполненная.\n", taskID)
	return nil
}

// deleteTask удаляет задачу из базы данных
func deleteTask(taskID string) error {
	_, err := db.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу с ID %s: %v", taskID, err)
	}
	log.Printf("Задача с ID %s успешно удалена.\n", taskID)
	return nil
}
