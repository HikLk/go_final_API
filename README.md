# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

В директории `tasks` содержатся данные задач.

В директории `handlers` содержатся обработчики запросов.

В директории `dates` содержатся функции и обработчики для расчета дат.

В директории `database` содержатся функции для работы с БД.

Директория `constants` содержит список используемых констант.

Команды для запуска приложения: 

go mod download

go run main.go

Команды для запуска тестов:

go test ./tests -v

Для очистки кэша перед повторными тестами

go clean -testcache

Реализованы задания со звездочками:

Реализована возможность определять извне порт при запуске сервера (переменная окружения TODO_PORT)

Реализована возможность определять путь к файлу базы данных через переменную окружения (переменная окружения TODO_DBFILE)

Для добавления переменных для запуска тестов TODO_PORT=7540 go test ./tests для bash и zsh и set TODO_PORT=7540 && go test ./tests для cmd

Обработка поисковых запросов для обработчика Get("/api/tasks")