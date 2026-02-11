# Script de inicio rápido con Swagger UI
# Uso: .\start-with-swagger.ps1

Write-Host "=== Iniciando ms-authentication con Swagger UI ===" -ForegroundColor Cyan
Write-Host ""

# Verificar que estamos en el directorio correcto
if (-not (Test-Path "cmd/api/main.go")) {
    Write-Host "Error: Debes ejecutar este script desde el directorio raíz del proyecto" -ForegroundColor Red
    exit 1
}

Write-Host "1. Verificando Go..." -ForegroundColor Yellow
$goVersion = go version
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Go instalado: $goVersion" -ForegroundColor Green
} else {
    Write-Host "✗ Go no está instalado o no está en el PATH" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "2. Verificando dependencias..." -ForegroundColor Yellow
go mod download
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Dependencias descargadas" -ForegroundColor Green
} else {
    Write-Host "✗ Error al descargar dependencias" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "3. Generando documentación Swagger..." -ForegroundColor Yellow
$env:Path += ";$env:GOPATH\bin"
swag init -g cmd/api/main.go --output docs 2>&1 | Out-Null
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Documentación Swagger generada" -ForegroundColor Green
} else {
    Write-Host "⚠ Swagger CLI no encontrado, intentando instalar..." -ForegroundColor Yellow
    go install github.com/swaggo/swag/cmd/swag@latest
    $env:Path += ";$env:GOPATH\bin"
    swag init -g cmd/api/main.go --output docs 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Documentación Swagger generada" -ForegroundColor Green
    } else {
        Write-Host "⚠ No se pudo generar documentación Swagger (continuando sin ella)" -ForegroundColor Yellow
    }
}
Write-Host ""

Write-Host "4. Iniciando servidor..." -ForegroundColor Yellow
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "  Servidor iniciándose..." -ForegroundColor White
Write-Host "  Espera unos segundos..." -ForegroundColor White
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Iniciar servidor en background
$serverJob = Start-Job -ScriptBlock {
    param($dir)
    Set-Location $dir
    go run cmd/api/main.go
} -ArgumentList $PWD

# Esperar a que el servidor esté listo
Write-Host "Esperando a que el servidor esté disponible..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$serverReady = $false

while ($attempt -lt $maxAttempts -and -not $serverReady) {
    Start-Sleep -Seconds 1
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 1 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            $serverReady = $true
        }
    } catch {
        $attempt++
    }
}

if ($serverReady) {
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
    Write-Host "  ✓ Servidor iniciado correctamente!" -ForegroundColor Green
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
    Write-Host ""
    Write-Host "URLs disponibles:" -ForegroundColor Cyan
    Write-Host "  • Swagger UI:    " -NoNewline -ForegroundColor White
    Write-Host "http://localhost:8080/swagger/index.html" -ForegroundColor Yellow
    Write-Host "  • Health Check:  " -NoNewline -ForegroundColor White
    Write-Host "http://localhost:8080/health" -ForegroundColor Yellow
    Write-Host "  • API Base:      " -NoNewline -ForegroundColor White
    Write-Host "http://localhost:8080" -ForegroundColor Yellow
    Write-Host ""
    
    Write-Host "Endpoints principales:" -ForegroundColor Cyan
    Write-Host "  • POST /auth/register - Registrar usuario" -ForegroundColor White
    Write-Host "  • POST /auth/login    - Iniciar sesión" -ForegroundColor White
    Write-Host "  • GET  /auth/validate - Validar token" -ForegroundColor White
    Write-Host "  • GET  /api/profile   - Perfil (requiere JWT)" -ForegroundColor White
    Write-Host ""
    
    Write-Host "Abriendo Swagger UI en el navegador..." -ForegroundColor Yellow
    Start-Sleep -Seconds 1
    Start-Process "http://localhost:8080/swagger/index.html"
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "  Presiona Ctrl+C para detener el servidor" -ForegroundColor White
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host ""
    
    # Mostrar logs del servidor
    Receive-Job -Job $serverJob -Wait
} else {
    Write-Host ""
    Write-Host "✗ El servidor no pudo iniciarse" -ForegroundColor Red
    Write-Host "Revisa los logs para más información" -ForegroundColor Yellow
    Stop-Job -Job $serverJob
    Remove-Job -Job $serverJob
    exit 1
}

# Cleanup
Stop-Job -Job $serverJob
Remove-Job -Job $serverJob
