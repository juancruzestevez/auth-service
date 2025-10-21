# 🔐 Auth Service - Go

Microservicio de autenticación con Go, PostgreSQL y JWT.

## 🚀 Características

- Registro y login de usuarios
- Autenticación JWT
- PostgreSQL con GORM
- Arquitectura en capas (Clean Architecture)
- Tests unitarios (~88% cobertura)

## 📁 Estructura

```
auth-service/
├── config/          # Configuración
├── models/          # Modelos de datos
├── dto/             # Request/Response
├── repository/      # Acceso a datos
├── service/         # Lógica de negocio
├── handlers/        # HTTP handlers
├── router/          # Rutas
├── middleware/      # Middlewares
├── auth/            # JWT utils
└── tests/           # Tests
```

## ⚙️ Setup

### 1. Variables de entorno

```env
PORT=3000
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=auth_db
JWT_SECRET=your-secret-key
JWT_EXPIRES_HOURS=24
```

### 2. Ejecutar

```bash
# Con Docker
docker-compose up -d

# Local
go run main.go
```

## 📡 API

**Base URL:** `http://localhost:3000/api/v1`

### Registro
```bash
POST /api/v1/auth/register
{
  "first_name": "Juan",
  "last_name": "Pérez",
  "nickname": "juanp",
  "email": "juan@example.com",
  "password": "password123"
}
```

### Login
```bash
POST /api/v1/auth/login
{
  "email": "juan@example.com",
  "password": "password123"
}
```

### Respuesta
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": 1234567890,
  "user": { "..." }
}
```

## 🧪 Tests

```bash
# Todos los tests
go test ./tests/... -v

# Con cobertura
go test ./tests/... -cover
```

## 🛠️ Desarrollo

```bash
# Hot reload
air

# Build
go build

# Docker
docker-compose up -d
```

## 📖 Arquitectura

```
Request → Router → Handler → Service → Repository → Database
```

Cada capa tiene responsabilidad única y es fácil de testear.

---

**Cobertura:** ~88% | **Tests:** 33 casos

---

## 📄 Licencia

MIT License - ver archivo LICENSE

---

## 👨‍💻 Autor

Juan Cruz Estevez - [@juancruzestevez](https://github.com/juancruzestevez)

---

## 🙏 Agradecimientos

- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [GORM](https://gorm.io/) - ORM
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Hashing

---

**¡Happy Coding! 🚀**
