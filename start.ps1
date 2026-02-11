#!/usr/bin/env pwsh
# Script para iniciar el microservicio de autenticación

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MS-Authentication - Sistal" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Verificar que Go esté instalado
Write-Host "Verificando instalación de Go..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go no está instalado o no está en el PATH" -ForegroundColor Red
    Write-Host "Por favor, instala Go desde: https://golang.org/dl/" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Go instalado: $goVersion" -ForegroundColor Green
Write-Host ""

# Verificar que el archivo .env existe
Write-Host "Verificando archivo de configuración..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    Write-Host "⚠ Archivo .env no encontrado" -ForegroundColor Yellow
    Write-Host "Usando valores predeterminados de configuración" -ForegroundColor Yellow
} else {
    Write-Host "✓ Archivo .env encontrado" -ForegroundColor Green
}
Write-Host ""

# Instalar/actualizar dependencias
Write-Host "Instalando dependencias..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Falló la descarga de dependencias" -ForegroundColor Red
    exit 1
}
go mod tidy
Write-Host "✓ Dependencias instaladas correctamente" -ForegroundColor Green
Write-Host ""

# Compilar la aplicación
Write-Host "Compilando aplicación..." -ForegroundColor Yellow
go build -o bin/ms-authentication.exe cmd/api/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Falló la compilación" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Aplicación compilada correctamente" -ForegroundColor Green
Write-Host ""

# Iniciar la aplicación
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Iniciando Servidor..." -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Servidor escuchando en: http://localhost:8080" -ForegroundColor Green
Write-Host "Swagger UI disponible en: http://localhost:8080/swagger/index.html" -ForegroundColor Green
Write-Host ""
Write-Host "Presiona Ctrl+C para detener el servidor" -ForegroundColor Yellow
Write-Host ""

./bin/ms-authentication.exe
