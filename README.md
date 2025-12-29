# ms-authentication

Microservicio de autenticación desarrollado en Go utilizando Gin framework, PostgreSQL y arquitectura hexagonal.

## Características

- 🏗️ **Arquitectura Hexagonal**: Separación clara entre dominio, aplicación e infraestructura
- 🚀 **Gin Framework**: Framework web rápido y minimalista para Go
- 🐘 **PostgreSQL**: Base de datos relacional robusta
- 🔐 **JWT Authentication**: Autenticación basada en tokens JWT
- 🔒 **Bcrypt**: Hash seguro de contraseñas
- 🐳 **Docker**: Contenerización completa con Docker y Docker Compose

## Estructura del Proyecto

```
ms-authentication/
├── cmd/
│   └── api/
│       └── main.go              # Punto de entrada de la aplicación
├── internal/
│   ├── domain/                  # Capa de dominio
│   │   ├── entities/            # Entidades del dominio
│   │   └── ports/               # Interfaces (puertos)
│   ├── application/             # Capa de aplicación
│   │   └── usecases/            # Casos de uso
│   └── infrastructure/          # Capa de infraestructura
│       └── adapters/
│           ├── database/        # Adaptador de PostgreSQL
│           └── http/            # Adaptador HTTP (Gin)
├── config/                      # Configuración
├── Dockerfile                   # Dockerfile para la aplicación
├── docker-compose.yml           # Configuración de Docker Compose
└── .env.example                 # Ejemplo de variables de entorno
```

## Arquitectura Hexagonal

Este proyecto implementa una arquitectura hexagonal (también conocida como arquitectura de puertos y adaptadores):

- **Domain Layer**: Contiene las entidades y las interfaces (puertos) que definen el comportamiento
- **Application Layer**: Implementa la lógica de negocio a través de casos de uso
- **Infrastructure Layer**: Contiene las implementaciones concretas (adaptadores) como la base de datos y HTTP

## Requisitos Previos

- Go 1.21 o superior
- Docker y Docker Compose (para ejecución con contenedores)
- PostgreSQL 15 (si se ejecuta sin Docker)

## Instalación y Ejecución

### Opción 1: Usando Docker Compose (Recomendado)

1. Clonar el repositorio:
```bash
git clone https://github.com/Sistal/ms-authentication.git
cd ms-authentication
```

2. Construir y ejecutar los contenedores:
```bash
docker-compose up --build
```

La aplicación estará disponible en `http://localhost:8080`

### Opción 2: Ejecución Local

1. Clonar el repositorio:
```bash
git clone https://github.com/Sistal/ms-authentication.git
cd ms-authentication
```

2. Copiar el archivo de ejemplo de variables de entorno:
```bash
cp .env.example .env
```

3. Configurar las variables de entorno en `.env` según tu configuración

4. Instalar las dependencias:
```bash
go mod download
```

5. Ejecutar la aplicación:
```bash
go run cmd/api/main.go
```

## Endpoints de la API

### Health Check
```
GET /health
```

### Registro de Usuario
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