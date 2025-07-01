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
