# ms-authentication

Microservicio de autenticación con JWT para el sistema Sistal.

## 🎯 Estado Actual

✅ **Aplicación funcionando y conectada a Supabase**
- Servidor corriendo en: http://localhost:8080
- Documentación Swagger: http://localhost:8080/swagger/index.html
- Base de datos: Supabase PostgreSQL
- Estado: **OPERATIVO**

## Características

- ✅ Registro de usuarios
- ✅ Login con múltiples identificadores (username, email o RUT)
- ✅ Validación de tokens JWT
- ✅ Hash de contraseñas con bcrypt
- ✅ Tokens firmados con HS256
- ✅ Expiración de tokens (1 hora)
- ✅ Middleware JWT reutilizable
- ✅ Arquitectura limpia (handlers, services, repositories)
- ✅ Conexión a Supabase PostgreSQL
- ✅ Documentación Swagger completa
- 🐳 Docker y Docker Compose

## Requisitos

- Go 1.24+
- Conexión a Internet (para acceder a Supabase)

## 🚀 Inicio Rápido

### Opción 1: Script PowerShell (Recomendado)
```powershell
.\start.ps1
```

Este script:
- Verifica la instalación de Go
- Instala las dependencias
- Compila la aplicación
- Inicia el servidor

### Opción 2: Manual
```powershell
# Instalar dependencias
go mod download
go mod tidy

# Compilar y ejecutar
go build -o bin/ms-authentication.exe cmd/api/main.go
.\bin\ms-authentication.exe
```

### Opción 3: Ejecutar directamente
```powershell
go run cmd/api/main.go
```

## ⚙️ Configuración

La aplicación está configurada para conectarse a Supabase. Ver archivo `.env`:

```env
SERVER_PORT=8080

# Supabase PostgreSQL
DB_HOST=db.fbcdvhoectqyofnezwfe.supabase.co
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=LZGHbNuD5cJivgiH
DB_NAME=postgres
DB_SSLMODE=require

JWT_SECRET=a9F3kL7QxR2ZP8mVwD6H4JcN5EBySUTG
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
```

O usando Make:
```bash
make run
```

## Endpoints

### POST /auth/register

Registra un nuevo usuario.

**Request:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepass123",
  "id_rol": 1
}
```

**Response (201):**
```json
{
  "user_id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "message": "User registered successfully"
}
```

### POST /auth/login

Autentica un usuario y retorna un token JWT.

**Request:**
```json
{
  "username": "johndoe",
  "password": "securepass123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user_id": 1,
  "username": "johndoe",
  "role": 1,
  "expires_at": 1704672000
}
```

### GET /auth/validate

Valida un token JWT.

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "valid": true,
  "user_id": 1,
  "username": "johndoe",
  "role": 1,
  "issued_at": 1704668400,
  "expires_at": 1704672000
}
```

### GET /health

Health check del servicio.

**Response (200):**
```json
{
  "status": "ok",
  "service": "ms-authentication"
}
```

## Ejemplos de uso con rutas protegidas

### GET /api/profile (requiere JWT)

Obtiene el perfil del usuario autenticado.

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "user_id": 1,
  "username": "johndoe",
  "role": 1
}
```

### GET /api/admin/users (requiere JWT y rol admin)

Endpoint de ejemplo que solo pueden acceder usuarios con rol 1 (admin).

**Headers:**
```
Authorization: Bearer <token>
```

**Response (200):**
```json
{
  "message": "admin endpoint - list users"
}
```

## Estructura del proyecto

```
ms-authentication/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── config/
│   └── config.go                   # Configuración
├── internal/
│   ├── application/
│   │   └── usecases/
│   │       └── auth_usecase.go     # Casos de uso
│   ├── domain/
│   │   ├── entities/
│   │   │   └── usuario.go          # Entidades
│   │   ├── repositories/
│   │   │   └── usuario_repository.go  # Interfaces
│   │   └── services/
│   │       └── auth_service.go     # Servicios
│   └── infrastructure/
│       ├── auth/
│       │   └── jwt_auth_service.go # Implementación JWT
│       ├── database/
│       │   └── postgres_usuario_repository.go  # Repository
│       └── http/
│           ├── handlers/
│           │   └── auth_handler.go # HTTP handlers
│           ├── middleware/
│           │   └── jwt_middleware.go  # Middleware JWT
│           └── routes/
│               └── routes.go       # Rutas
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Modelo de datos

### Usuario
```sql
CREATE TABLE "Usuario" (
  id_usuario SERIAL PRIMARY KEY,
  nombre_usuario VARCHAR(100) NOT NULL UNIQUE,
  email VARCHAR(100) NOT NULL UNIQUE,
  id_rol INT NOT NULL DEFAULT 1,
  id_estado_usuario INT NOT NULL DEFAULT 1,
  fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  fecha_modificacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)
```

### Credencial
```sql
CREATE TABLE "Credencial" (
  id_credencial SERIAL PRIMARY KEY,
  id_usuario INT NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  fecha_creacion TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_credencial_usuario 
    FOREIGN KEY (id_usuario) 
    REFERENCES "Usuario"(id_usuario) 
    ON DELETE CASCADE
)
```

## JWT Claims

El token JWT contiene los siguientes claims:

```json
{
  "sub": 1,                    // ID del usuario
  "username": "johndoe",       // Nombre de usuario
  "role": 1,                   // Rol del usuario
  "iat": 1704668400,          // Issued at (timestamp)
  "exp": 1704672000           // Expires at (timestamp)
}
```

## Docker

Construir imagen:
```bash
docker build -t ms-authentication .
```

Ejecutar con Docker Compose:
```bash
docker-compose up
```

## Desarrollo

Ejecutar tests (cuando estén implementados):
```bash
go test ./...
```

Compilar:
```bash
go build -o bin/ms-authentication cmd/api/main.go
```

## Seguridad

- Las contraseñas se hashean usando bcrypt con costo por defecto (10)
- Los tokens JWT se firman con HS256
- El secret debe tener al menos 32 caracteres
- Los tokens expiran en 1 hora
- Se valida el formato y la expiración en cada request

## Mejoras futuras

- [ ] Refresh tokens
- [ ] Revocación de tokens (blacklist)
- [ ] Rate limiting
- [ ] Logs estructurados
- [ ] Métricas
- [ ] Tests unitarios e integración
- [ ] Documentación OpenAPI/Swagger
```
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

### Login
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Respuesta:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Obtener Perfil (Requiere autenticación)
```
GET /api/v1/user/profile
Authorization: Bearer <token>
```

## Variables de Entorno

| Variable | Descripción | Valor por defecto |
|----------|-------------|-------------------|
| SERVER_PORT | Puerto del servidor | 8080 |
| DB_HOST | Host de PostgreSQL | localhost |
| DB_PORT | Puerto de PostgreSQL | 5432 |
| DB_USER | Usuario de PostgreSQL | postgres |
| DB_PASSWORD | Contraseña de PostgreSQL | postgres |
| DB_NAME | Nombre de la base de datos | authentication |
| DB_SSLMODE | Modo SSL de PostgreSQL | disable |
| JWT_SECRET | Clave secreta para JWT | your-secret-key-change-in-production |

## Construcción de la Imagen Docker

Para construir la imagen Docker manualmente:

```bash
docker build -t ms-authentication:latest .
```

Para ejecutar el contenedor:

```bash
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=authentication \
  ms-authentication:latest
```

## Desarrollo

### Agregar Nuevas Dependencias

```bash
go get <package-name>
go mod tidy
```

### Ejecutar Tests

```bash
go test ./...
```

## Tecnologías Utilizadas

- **Go**: Lenguaje de programación
- **Gin**: Framework web
- **PostgreSQL**: Base de datos
- **JWT**: Autenticación basada en tokens
- **Bcrypt**: Hash de contraseñas
- **Docker**: Contenerización
- **UUID**: Generación de identificadores únicos

## Seguridad

- Las contraseñas se almacenan hasheadas usando bcrypt
- Los tokens JWT expiran en 24 horas
- Se implementa validación de entrada en todos los endpoints
- **IMPORTANTE**: Cambiar `JWT_SECRET` en producción

## Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Fork del proyecto
2. Crear una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit de tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abrir un Pull Request

## Licencia

Este proyecto está bajo la licencia MIT.

## Autor

Sistal