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
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получение параметров из строки запроса
	nowParam := r.URL.Query().Get("now")
	repeatParam := r.URL.Query().Get("repeat")
	dateParam := r.URL.Query().Get("date")

	// Проверка и парсинг параметра now
	now, err := time.Parse(DateFormat, nowParam)
	if err != nil {
		http.Error(w, "Invalid 'now' date format, must be YYYYMMDD", http.StatusBadRequest)
		return
	}

	// Вызов функции NextDate для получения следующей даты задачи
	nextDate, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write([]byte(nextDate)); err != nil {
		log.Printf("Ошибка при записи ответа: %v", err)
	}
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	// Проверка обязательного поля title
	if task.Title == "" {
		sendErrorResponse(w, "Не указано название задачи")
		return
	}

	// Определение текущей даты
	now := time.Now()
	today := now.Format(DateFormat)

	// Проверка и обработка поля date
	if task.Date == "" {
		task.Date = today
	} else {
		parsedDate, err := time.Parse(DateFormat, task.Date)
		if err != nil {
			sendErrorResponse(w, "Дата в неверном формате")
			return
		}

		parsedDate = parsedDate.Truncate(24 * time.Hour)
		now = now.Truncate(24 * time.Hour)

		// Если дата меньше сегодняшней
		if parsedDate.Before(now) {
			if task.Repeat == "" {
				// Если нет правила повторения, подставляем сегодняшнюю дату
				task.Date = today
			} else {
				// Применяем правило повторения, чтобы получить следующую дату
				nextDate, err := NextDate(now, task.Date, task.Repeat)
				if err != nil {
					sendErrorResponse(w, "Ошибка в правиле повторения")
					return
				}

				// Если правило повторения - "d 1" и сегодняшняя дата допустима, устанавливаем её
				if task.Repeat == "d 1" && nextDate == today {
					task.Date = today
				} else {
					task.Date = nextDate
				}
			}
		}
	}

	// Проверка правила повторения с использованием NextDate()
	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendErrorResponse(w, "Неверное правило повторения")
			return
		}
	}

	// Добавление задачи в базу данных
	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		sendErrorResponse(w, "Ошибка добавления задачи в базу данных")
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		sendErrorResponse(w, "Ошибка получения ID задачи")
		return
	}

	// Формируем успешный ответ с ID задачи
	response := map[string]string{"id": fmt.Sprintf("%d", id)}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse отправляет JSON-ответ с полем error
func sendErrorResponse(w http.ResponseWriter, errorMessage string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]string{"error": errorMessage})
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(fmt.Sprintf(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		ORDER BY date ASC
		LIMIT %d
	`, limit))
	if err != nil {
		log.Printf("Ошибка при запросе к базе данных: %v", err)
		sendErrorResponse(w, "Ошибка получения задач из базы данных")
		return
	}
	defer rows.Close()

	tasks := []Task{}

	// Заполнение массива задач
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Printf("Ошибка при сканировании строки: %v", err)
			sendErrorResponse(w, "Ошибка чтения задач из базы данных")
			return
		}
		tasks = append(tasks, task)
	}

	// Проверка на наличие ошибок после завершения итерации
	if err = rows.Err(); err != nil {
		log.Printf("Ошибка при итерации по строкам: %v", err)
		sendErrorResponse(w, "Ошибка итерации по задачам")
		return
	}

	// Формирование JSON-ответа
	response := map[string]interface{}{
		"tasks": tasks,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		sendErrorResponse(w, "Ошибка кодирования задач в JSON")
		return
	}
}

func getTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Не указан ID задачи")
		return
	}

	var task Task
	err := db.QueryRow(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?
	`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "Задача не найдена")
		} else {
			log.Printf("Ошибка при запросе к базе данных: %v", err)
			sendErrorResponse(w, "Ошибка получения задачи")
		}
		return
	}

	// Формирование JSON-ответа с задачей
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		log.Printf("Ошибка при кодировании JSON: %v", err)
		sendErrorResponse(w, "Ошибка кодирования задачи в JSON")
	}
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		sendErrorResponse(w, "Не указан ID задачи")
		return
	}
	if task.Title == "" {
		sendErrorResponse(w, "Не указано название задачи")
		return
	}

	// Проверка и обработка поля date
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	} else {
		parsedDate, err := time.Parse(DateFormat, task.Date)
		if err != nil {
			sendErrorResponse(w, "Дата в неверном формате")
			return
		}
		if parsedDate.Before(now) && task.Repeat != "" {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				sendErrorResponse(w, "Ошибка в правиле повторения")
				return
			}
			task.Date = nextDate
		}
	}

	// Обновление записи в базе данных
	res, err := db.Exec(`
		UPDATE scheduler
		SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?
	`, task.Date, task.Title, task.Comment, task.Repeat, task.ID)

	if err != nil {
		log.Printf("Ошибка при запросе к базе данных: %v", err)
		sendErrorResponse(w, "Ошибка обновления задачи")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		sendErrorResponse(w, "Задача не найдена")
		return
	}

	// Возвращаем пустой JSON-ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}

// markTaskDoneHandler обрабатывает POST-запрос для отметки задачи выполненной
func markTaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	// Получение идентификатора задачи
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Не указан ID задачи")
		return
	}

	// Извлечение задачи из базы данных
	var task Task
	err := db.QueryRow(`
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?
	`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, "Задача не найдена")
		} else {
			log.Printf("Ошибка при запросе к базе данных: %v", err)
			sendErrorResponse(w, "Ошибка получения задачи")
		}
		return
	}

	// Если задача одноразовая (repeat пуст), удаляем её
	if task.Repeat == "" {
		_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			log.Printf("Ошибка при удалении задачи: %v", err)
			sendErrorResponse(w, "Ошибка удаления задачи")
			return
		}
	} else {
		// Для периодической задачи вычисляем следующую дату
		now := time.Now()
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			log.Printf("Ошибка при расчете следующей даты: %v", err)
			sendErrorResponse(w, "Ошибка расчета следующей даты для периодической задачи")
			return
		}

		// Обновляем дату выполнения задачи в базе данных
		_, err = db.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
		if err != nil {
			log.Printf("Ошибка при обновлении даты задачи: %v", err)
			sendErrorResponse(w, "Ошибка обновления даты задачи")
			return
		}
	}

	// Возвращаем пустой JSON-ответ
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}

// deleteTaskHandler обрабатывает DELETE-запрос для удаления задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получение идентификатора задачи
	id := r.URL.Query().Get("id")
	if id == "" {
		sendErrorResponse(w, "Не указан ID задачи")
		return
	}

	// Выполнение запроса на удаление и получение количества затронутых строк
	res, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		log.Printf("Ошибка при удалении задачи: %v", err)
		sendErrorResponse(w, "Ошибка удаления задачи")
		return
	}

	// Проверяем, была ли удалена хотя бы одна строка
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("Ошибка при получении количества удаленных строк: %v", err)
		sendErrorResponse(w, "Ошибка при получении количества удаленных строк")
		return
	}

	if rowsAffected == 0 {
		// Если строк с таким id не найдено, возвращаем ошибку
		sendErrorResponse(w, "Задача не найдена")
		return
	}

	// Возвращаем пустой JSON-ответ в случае успешного удаления
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}
