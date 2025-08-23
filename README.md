📦 DemoServ

---

Микросервис для обработки и управления заказами электронной коммерции с интеграцией Kafka, PostgreSQL и кэшем в памяти.

---
✨ **Описание**

DemoServ — это демонстрационный сервис для работы с заказами. Он принимает сообщения из Kafka, валидирует их, сохраняет в PostgreSQL и кэширует в памяти для быстрого доступа. Также сервис предоставляет HTTP API для получения данных о заказах и простой фронтенд для отображения.

---
🧩 **Основной функционал**

* Получение заказа по `order_uid` через HTTP API
* Потокобезопасный кэш с инициализацией из базы данных
* Интеграция с Apache Kafka (producer + consumer)
* Валидация данных заказов
* Работа с PostgreSQL через транзакции и миграции
* Поддержка Docker Compose для инфраструктуры
* Веб-интерфейс для просмотра заказов
---
🌐 **API**

`GET /order/{order_uid}` — получение заказа по ID

📌 **Пример ответа:**

```json
{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": { "name": "Test Testov", "phone": "+9720000000" },
  "payment": { "transaction": "b563feb7b2b84b6test", "amount": 1817 },
  "items": [ { "name": "Mascaras", "price": 453 } ]
}
```
---
🛠️ **Технологии**

* Язык: Go 1.24.6
* Фреймворк: Chi
* База данных: PostgreSQL 15
* Очередь сообщений: Apache Kafka
* Кэш: in-memory cache
* Конфигурации: YAML + cleanenv
* Тестирование: testify + pgxmock
* Контейнеризация: Docker + Docker Compose
---
📁 **Структура проекта**

```
test_task/
├── cmd
│   └── main.go
├── config
│   └── config.yaml
├── db
│   └── migrations
│       ├── 1_init.up.sql
│       └── 1_init.down.sql
├── frontend
│   ├── index.html
│   └── styles/styles.css
├── internal
│   ├── cache
│   ├── config
│   ├── http-server/handlers/getOrder
│   ├── kafka
│   ├── models
│   ├── postgres
│   ├── testutils
│   └── validate
├── docker-compose.yaml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```
---
⚙️ **Установка и запуск**

1. Клонируйте репозиторий

```bash
git clone https://github.com/A1imuhammad/wb-tetst-task.git
cd test_task
```

2. Запустите инфраструктуру

```bash
docker-compose up -d
```

Поднимется:

* PostgreSQL (порт 5432)
* Kafka (порты 9092-9094) + Zookeeper
* Kafka UI (порт 9001)

3. Примените миграции

Применяются автоматически при первом запуске.

4. Запустите приложение

```bash
make run
# или вручную
go build -o bin/demoserv ./cmd
./bin/demoserv
```

Сервис доступен по адресу: [http://localhost:8085](http://localhost:8085)
---
🖥️ **Фронтенд**

Для запуска интерфейса достаточно открыть файл `frontend/index.html` в браузере:

* **Windows:** двойной клик по `frontend/index.html`
* **Linux / macOS:**

```bash
xdg-open frontend/index.html
# или
open frontend/index.html
```
---
🔧 **Конфигурация**

Перед запуском необходимо настроить конфиг:

1. Скопируйте шаблон:
   ```bash
   cp config/config.example.yaml config/config.yaml
2. Заполните config/config.yaml своими значениями (PostgreSQL, Kafka и HTTP сервер).

Файл `config/config.yaml`:

```yaml
POSTGRES:
  POSTGRES_USER: root   
  POSTGRES_PASSWORD: 1234 
  POSTGRES_PORT: 5432
  POSTGRES_DB: orders_db
  POSTGRES_HOST: localhost

KAFKA:
  KAFKA_BROKER: "localhost:9092"
  KAFKA_TOPIC: "my-topic"
  KAFKA_GROUP: "group"

HTTP_SERVER:
  ADDRESS: "localhost:8085"
  TIMEOUT: 4s
```
---
🧪 **Тестирование**

```bash
make test       # все тесты
go test ./...   # запустить вручную
```

Примеры:

```bash
go test ./internal/cache -v
go test ./internal/postgres -v
```
---
📊 **Мониторинг**

* Kafka UI: [http://localhost:9001](http://localhost:9001)
* PostgreSQL: `localhost:5432`
* HTTP API: [http://localhost:8085](http://localhost:8085)
* Фронтенд: открыть `frontend/index.html`
---
🧑‍💻 **Автор**

Telegram: [@zag1rov](https://t.me/zag1rov)
