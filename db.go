package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // Импорт SQLite-драйвера
)

// Схема базы данных
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT, -- Уникальный идентификатор записи
	date TEXT NOT NULL,                   -- Дата задачи (обязательное поле)
	title TEXT NOT NULL,                  -- Название задачи (обязательное поле)
	comment TEXT,                         -- Комментарий к задаче (необязательное поле)
	repeat TEXT                           -- Поле для информации о повторении задачи
);
`

// Глобальная переменная для хранения подключения к базе данных
var DB *sql.DB

// Функция для инициализации базы данных
func InitializeDatabase() (*sql.DB, error) {
	// Открываем соединение с SQLite базой данных (файл scheduler.db)
	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		log.Printf("Ошибка при открытии базы данных: %v", err)
		return nil, err // Возвращаем ошибку, если не удалось открыть базу данных
	}

	// Выполняем создание таблицы, если она еще не существует
	_, err = db.Exec(schema)
	if err != nil {
		log.Printf("Ошибка при выполнении схемы базы данных: %v", err)
		db.Close() // Закрываем соединение перед возвратом ошибки
		return nil, err
	}

	log.Println("База данных успешно инициализирована")
	DB = db        // Сохраняем подключение в глобальную переменную
	return db, nil // Возвращаем соединение с базой данных
}
