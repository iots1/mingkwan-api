# mingkwan-api

A modular, scalable backend service built with **Golang** and **Fiber**, supporting **JWT authentication**, **MongoDB**, **Redis**, **Swagger documentation**, and **task queues with Asynq**.

> 🔥 Auto-reload with [Air](https://github.com/cosmtrek/air)  
> 📄 API documentation powered by [Swaggo](https://github.com/swaggo/swag)

GitHub: [github.com/iots1/mingkwan-api](https://github.com/iots1/mingkwan-api)

---

## 📁 Project Structure

```bash
├── cmd/
│   └── app/                 # Entry point for the application
│       └── main.go
├── config/                  # Environment and config management
│   └── config.go
├── docs/                    # Swagger auto-generated documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/                # Main application code
│   ├── auth/                # Authentication module
│   ├── media/               # Media-related logic (future use)
│   ├── modules/             # Module initializers for DI
│   ├── payment/             # Payment processing (future use)
│   ├── shared/              # Shared utilities, cache, event, etc.
│   └── user/                # User registration, profile, etc.
├── tmp/                     # Temporary files (ignored in prod)
├── go.mod / go.sum          # Go dependencies
├── Readme.md                # You’re reading it now!
```

---

## 🚀 Features

- ⚙️ **Modular architecture** with clear separation of concerns
- 🔐 **JWT-based Authentication**
- 📦 **MongoDB** and **Redis** support
- 🪝 **Asynchronous events** via Asynq
- 🔁 **Live reload** with [Air](https://github.com/cosmtrek/air)
- 📘 **Swagger docs** with [Swaggo](https://github.com/swaggo/swag)

---

## 🛠️ Getting Started

### 1. Clone and Install

```bash
git clone https://github.com/iots1/mingkwan-api.git
cd mingkwan-api
go mod tidy
```

### 2. Run with Air (auto-reload)

```bash
air
```

> Ensure you’ve installed Air:  
> `go install github.com/cosmtrek/air@latest`

### 3. Generate Swagger Docs

```bash
swag init -g cmd/app/main.go
```

Then access Swagger UI at:  
`http://localhost:3000/swagger/index.html`

---

## 📄 API Documentation

Generated using [Swaggo](https://github.com/swaggo/swag). You can add Swagger annotations directly to your handler methods.

---

## 📌 Environment Variables

Create a `.env` file at the root directory with the following example values:

```env
# Application Config
APP_PORT=3000
APP_ENV=development
APP_SECRET_KEY=your-secret-key
LOG_LEVEL=debug

# MongoDB Config
MONGO_URI=mongodb+srv://<username>:<password>@<cluster>.mongodb.net/?retryWrites=true&w=majority
MONGO_DB_NAME=mingkwan_db

# Redis Config
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0
```

> ⚠️ **Warning**: Do not commit your actual `.env` file to version control.  
> Add `.env` to `.gitignore`.

---

## 🧪 Sample Endpoints

| Method | Path                | Description        |
|--------|---------------------|--------------------|
| POST   | `/auth/login`       | Login via JWT      |
| GET    | `/user/profile`     | Get user profile   |
| POST   | `/user/register`    | Register new user  |

---

## 📦 Essential Libraries & Concepts

### 🔧 Required Libraries

| Library                             | Purpose                                                                 |
|-------------------------------------|-------------------------------------------------------------------------|
| `github.com/gofiber/fiber/v2`       | Fast, Express-like web framework for Go                                 |
| `github.com/redis/go-redis/v9`      | Redis client for caching and message queue (Asynq)                      |
| `github.com/swaggo/fiber-swagger`   | Swagger UI handler for Fiber                                            |
| `github.com/swaggo/swag`            | CLI for generating Swagger docs from annotations                        |
| `go.uber.org/zap`                   | High-performance structured logging                                     |
| `github.com/go-playground/validator/v10` | Input validation for request models                              |
| `github.com/golang-jwt/jwt/v5`      | JWT authentication and token handling                                  |
| `github.com/hibiken/asynq`          | Task queue library using Redis (for background jobs)                    |
| `github.com/joho/godotenv`          | Loads environment variables from .env file                              |
| `go.mongodb.org/mongo-driver`       | Official MongoDB driver for Go                                          |

---

### 🧠 Core Concepts

#### 🟡 In-Memory Event Bus (`InMemPubSub`)

The `InMemPubSub` is a simple pub/sub implementation for broadcasting events within the same process.  
It is ideal for fast, internal communication where persistence and durability are not required.

- Used in this project for broadcasting low-importance events (e.g., in-memory logging, counters)
- Subscribers are launched in goroutines and listen using channels
- Fast, zero external dependencies

Example:

```go
inMemPubSub := event.NewInMemoryBus()
subscriber := event.NewUserInmemoryEventSubscribers(inMemPubSub)
subscriber.StartAllSubscribers(ctx)
```

#### 🔴 Asynq (Redis-Based Task Queue)

[Asynq](https://github.com/hibiken/asynq) is a Go library for background job processing backed by Redis.

In this project:

- High-importance events (e.g., email sending, database syncing) are published using Asynq
- Redis is used for reliability, retries, and scheduling
- Tasks are enqueued via `AsynqClient`, and handlers are registered in a separate worker process (not shown here)

```go
asynqClient := event.NewAsynqClient(asynq.RedisClientOpt{Addr: "localhost:6379"})
publisher := event.NewHighImportancePublisher(asynqClient)
publisher.PublishUserCreated(ctx, payload)
```

---

### ⚙️ System Design Note

This application is designed as a **modular monolith**, meaning:

- Logic is organized by feature in modular folders (e.g., `auth`, `user`, `payment`)
- Event-driven communication is used to simulate microservice boundaries
- Designed with future migration to microservices in mind (clear dependency boundaries, async messaging)

This makes it easier to scale vertically first, then move specific modules to separate services later.


---

## 📦 Essential Libraries & Concepts

### 🔧 Required Libraries

| Library                             | Purpose                                                                 |
|-------------------------------------|-------------------------------------------------------------------------|
| `github.com/gofiber/fiber/v2`       | Fast, Express-like web framework for Go                                 |
| `github.com/redis/go-redis/v9`      | Redis client for caching and message queue (Asynq)                      |
| `github.com/swaggo/fiber-swagger`   | Swagger UI handler for Fiber                                            |
| `github.com/swaggo/swag`            | CLI for generating Swagger docs from annotations                        |
| `go.uber.org/zap`                   | High-performance structured logging                                     |
| `github.com/go-playground/validator/v10` | Input validation for request models                              |
| `github.com/golang-jwt/jwt/v5`      | JWT authentication and token handling                                  |
| `github.com/hibiken/asynq`          | Task queue library using Redis (for background jobs)                    |
| `github.com/joho/godotenv`          | Loads environment variables from .env file                              |
| `go.mongodb.org/mongo-driver`       | Official MongoDB driver for Go                                          |

---

## 📌 TODO / Coming Soon

- [ ] OAuth2 Support  
- [ ] File Upload (Media module)  
- [ ] Payment Gateway Integration  
- [ ] CI/CD with GitHub Actions

---

## 👨‍💻 Author

**Jirayu Mool-Ang (iots1)**  
GitHub: [@iots1](https://github.com/iots1)

---

## 📄 License

MIT © 2025 - iots1

---