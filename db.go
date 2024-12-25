package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Константы и глобальные переменные
const DateFormat = "20060102"
const limit = 50

var db *sql.DB

// initializDB инициализирует подключение к базе данных и создает таблицу, если она не существует
func initializDB() (*sql.DB, error) {
	// Определяем путь к файлу базы данных
	appPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")

	// Проверяем, существует ли файл базы данных
	_, err = os.Stat(dbFile)
	install := os.IsNotExist(err)

	// Открываем базу данных
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	// Если база данных новая, создаем таблицу
	if install {
		createTableQuery := `
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT,
            status TEXT
        )`
		_, err := db.Exec(createTableQuery)
		if err != nil {
			return nil, fmt.Errorf("Ошибка при создании таблицы: %v", err)
		}
	}

	return db, nil
}
