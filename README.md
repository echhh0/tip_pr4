# Практическая работа №4

## Создание CRUD-сервиса «ToDo» с маршрутизацией на **chi** (альтернатива — gorilla/mux)

### Рузин Иван Александрович ЭФМО-01-25

---

## Основные задачи

1. Освоить использование маршрутизатора **chi** для обработки HTTP-запросов.
2. Реализовать эндпоинты с методами GET, POST, PUT, DELETE.
3. Создать простое приложение «Список задач» с хранением данных в памяти.
4. Добавить middleware для логирования и CORS.
5. Отработать тестирование API при помощи curl и Postman.

---

## Структура проекта

```
tip_pr4/
├── cmd/server/main.go
├── internal/task/
│   ├── handler.go
│   ├── model.go
│   └── repo.go
├── pkg/middleware/
│   ├── cors.go
│   └── logger.go
├── go.mod
└── go.sum
```

---

## Основные элементы реализации

### Маршрутизация

```go
r := chi.NewRouter()
r.Use(chimw.RequestID)
r.Use(chimw.Recoverer)
r.Use(myMW.Logger)
r.Use(myMW.SimpleCORS)

r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("OK"))
})

r.Route("/api/v1", func(api chi.Router) {
	api.Mount("/tasks", h.Routes())
})
```

### Middleware

```go
func SimpleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
```

---

## Обработчики запросов

* `GET /tasks` — вернуть список всех задач, поддерживает `done`, `page`, `limit`.
* `GET /tasks/{id}` — вернуть конкретную задачу.
* `POST /tasks` — добавить новую.
* `PUT /tasks/{id}` — обновить данные задачи.
* `DELETE /tasks/{id}` — удалить задачу.

Поддерживаются фильтрация и пагинация. Пример реализации:

```go
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := parsePositiveInt(q.Get("page"), 1)
	limit := parsePositiveInt(q.Get("limit"), 0)
	var filterDone *bool
	if raw := q.Get("done"); raw != "" {
		if b, err := strconv.ParseBool(raw); err == nil {
			filterDone = &b
		} else {
			httpError(w, http.StatusBadRequest, "invalid done param")
			return
		}
	}
	all := h.repo.List()
	// фильтрация и пагинация далее...
}
```

---

## Примеры запросов и ответов

`GET /tasks` — получить все задачи
![tasks](/img/get_all_no_params.png)

`POST /tasks` — успешное создание
![tasks](/img/post_201.png)

`POST /tasks` — ошибка валидации
![tasks](/img/post_422.png)

`GET /tasks/{id}` — получить по идентификатору
![tasks](/img/get_id.png)

`PUT /tasks/{id}` — обновление задачи
![tasks](/img/put.png)

`DELETE /tasks/{id}` — удаление
![tasks](/img/delete.png)

`GET /tasks?done=true` — фильтрация по статусу выполнения
![tasks](/img/get_all_done_filtered.png)

`GET /tasks?page=2&limit=1` — пагинация
![tasks](/img/get_all_paging.png)

`GET /tasks?page=1&limit=2&done=false` — фильтр и пагинация одновременно
![tasks](/img/get_all_params_everything.png)

---

## Коды состояния и логика ошибок

| Код     | Когда возвращается | Описание                         |
|---------|--------------------|----------------------------------|
| **200** | успешный запрос    | возвращаются данные              |
| **201** | при создании       | новая задача добавлена           |
| **204** | при удалении       | тело отсутствует                 |
| **400** | неверный запрос    | ошибка в параметрах, ID или JSON |
| **404** | не найдено         | указанная задача отсутствует     |
| **422** | ошибка валидации   | нарушены ограничения по `title`  |

---

## Проверка функциональности

| Тест                       | Запрос                      | Ожидаемый результат      |
|----------------------------|-----------------------------|--------------------------|
| Создание корректной задачи | `POST /tasks`               | 201 Created              |
| Создание с коротким title  | `POST /tasks`               | 422 Unprocessable Entity |
| Получение списка           | `GET /tasks`                | 200 OK                   |
| Фильтр `done=true`         | `GET /tasks?done=true`      | 200 OK                   |
| Пагинация                  | `GET /tasks?page=1&limit=1` | 200 OK                   |
| Удаление существующей      | `DELETE /tasks/1`           | 204 No Content           |
| Удаление несуществующей    | `DELETE /tasks/999`         | 404 Not Found            |

---

## Итог

В проекте реализован REST-сервис на Go с маршрутизатором **chi**.
Все CRUD-операции для сущности `tasks` работают корректно.
Добавлены middleware, фильтры, пагинация и базовая валидация.
Структура кода модульная: бизнес-логика отделена от маршрутов и вспомогательных функций.

Дополнительные задачи:

1. Проверка длины `title` (3–100 символов).
2. Пагинация и фильтрация по статусу.
3. Версионирование API через `/api/v1`.
4. Простая логика сохранения в файл (опционально).
