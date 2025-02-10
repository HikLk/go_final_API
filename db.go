package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

func initializeDB() (*sql.DB, error) {
	// Определяем путь к файлу базы данных
	appPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")

	// Проверяем, существует ли файл базы данных
	_, err = os.Stat(dbFile)
	isNewDatabase := os.IsNotExist(err)

	// Открываем базу данных
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}

	// Создаем таблицу и индекс, если база данных новая
	if isNewDatabase {
		createTableQuery := `
		CREATE TABLE IF NOT EXISTS scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			title TEXT NOT NULL,
			comment TEXT,
			repeat TEXT CHECK(LENGTH(repeat) <= 128)
		);
		CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
		`
		_, err = db.Exec(createTableQuery)
		if err != nil {
			return nil, err
		}
		log.Println("Таблица scheduler успешно создана.")
	}

	return db, nil
}
