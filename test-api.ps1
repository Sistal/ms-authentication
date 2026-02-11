# Script de prueba para ms-authentication API
# Uso: .\test-api.ps1

Write-Host "=== Testing ms-authentication API ===" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# 0. Verificar Swagger UI
Write-Host "0. Verificando Swagger UI" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/swagger/index.html" -Method Get -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ Swagger UI disponible en: $baseUrl/swagger/index.html" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠ Swagger UI no disponible (pero esto no afecta la funcionalidad)" -ForegroundColor Yellow
}
Write-Host ""

# 1. Health Check
Write-Host "1. Health Check" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get
    Write-Host "✓ Health check OK" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "✗ Health check failed: $_" -ForegroundColor Red
}
Write-Host ""

# 2. Registro de usuario
Write-Host "2. Registrando nuevo usuario" -ForegroundColor Yellow
$registerBody = @{
    username = "testuser_$(Get-Random -Maximum 10000)"
    email = "test_$(Get-Random -Maximum 10000)@example.com"
    password = "securepass123"
    id_rol = 1
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" `
        -Method Post `
        -ContentType "application/json" `
        -Body $registerBody
    Write-Host "✓ Usuario registrado correctamente" -ForegroundColor Green
    $registerResponse | ConvertTo-Json
    $username = ($registerBody | ConvertFrom-Json).username
} catch {
    Write-Host "✗ Error al registrar usuario: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 3. Login
Write-Host "3. Iniciando sesión" -ForegroundColor Yellow
$loginBody = @{
    username = $username
    password = "securepass123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody
    Write-Host "✓ Login exitoso" -ForegroundColor Green
    Write-Host "Token recibido: $($loginResponse.token.Substring(0, 30))..." -ForegroundColor Cyan
    $token = $loginResponse.token
    $loginResponse | ConvertTo-Json
} catch {
    Write-Host "✗ Error al iniciar sesión: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 4. Validar token
Write-Host "4. Validando token JWT" -ForegroundColor Yellow
$headers = @{
    Authorization = "Bearer $token"
}

try {
    $validateResponse = Invoke-RestMethod -Uri "$baseUrl/auth/validate" `
        -Method Get `
        -Headers $headers
    Write-Host "✓ Token válido" -ForegroundColor Green
    $validateResponse | ConvertTo-Json
} catch {
    Write-Host "✗ Error al validar token: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 5. Acceder a ruta protegida (perfil)
Write-Host "5. Accediendo a ruta protegida (/api/profile)" -ForegroundColor Yellow
try {
    $profileResponse = Invoke-RestMethod -Uri "$baseUrl/api/profile" `
        -Method Get `
        -Headers $headers
    Write-Host "✓ Acceso a ruta protegida exitoso" -ForegroundColor Green
    $profileResponse | ConvertTo-Json
} catch {
    Write-Host "✗ Error al acceder a perfil: $_" -ForegroundColor Red
}
Write-Host ""

# 6. Acceder a ruta de admin
Write-Host "6. Accediendo a ruta de administrador (/api/admin/users)" -ForegroundColor Yellow
try {
    $adminResponse = Invoke-RestMethod -Uri "$baseUrl/api/admin/users" `
        -Method Get `
        -Headers $headers
    Write-Host "✓ Acceso a ruta de admin exitoso" -ForegroundColor Green
    $adminResponse | ConvertTo-Json
} catch {
    Write-Host "✗ Error al acceder a ruta de admin: $_" -ForegroundColor Red
}
Write-Host ""

# 7. Intentar acceder sin token (debe fallar)
Write-Host "7. Intentando acceder sin token (debe fallar)" -ForegroundColor Yellow
try {
    $unauthorizedResponse = Invoke-RestMethod -Uri "$baseUrl/api/profile" -Method Get
    Write-Host "✗ No debería haber accedido sin token" -ForegroundColor Red
} catch {
    Write-Host "✓ Acceso denegado correctamente (sin token)" -ForegroundColor Green
}
Write-Host ""

# 8. Intentar login con credenciales incorrectas (debe fallar)
Write-Host "8. Intentando login con credenciales incorrectas (debe fallar)" -ForegroundColor Yellow
$badLoginBody = @{
    username = $username
    password = "wrongpassword"
} | ConvertTo-Json

try {
    $badLoginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $badLoginBody
    Write-Host "✗ No debería haber iniciado sesión con password incorrecta" -ForegroundColor Red
} catch {
    Write-Host "✓ Login rechazado correctamente (credenciales inválidas)" -ForegroundColor Green
}
Write-Host ""

Write-Host "=== Tests completados ===" -ForegroundColor Cyan
