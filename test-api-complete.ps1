# Script de Prueba de API - MS Authentication
# Asegúrate de que la aplicación esté corriendo en http://localhost:8080

$baseUrl = "http://localhost:8080"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Pruebas de API - MS Authentication" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Función para hacer requests
function Invoke-APIRequest {
    param (
        [string]$Method,
        [string]$Endpoint,
        [object]$Body,
        [string]$Token
    )
    
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }
    
    try {
        $params = @{
            Uri = "$baseUrl$Endpoint"
            Method = $Method
            Headers = $headers
        }
        
        if ($Body) {
            $params["Body"] = ($Body | ConvertTo-Json)
        }
        
        $response = Invoke-RestMethod @params
        return $response
    }
    catch {
        Write-Host "Error: $_" -ForegroundColor Red
        return $null
    }
}

# 1. Verificar Health Check
Write-Host "1. Verificando Health Check..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "✓ Health Check OK: $($health | ConvertTo-Json -Compress)" -ForegroundColor Green
}
catch {
    Write-Host "✗ Health Check falló" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 2. Registrar un nuevo usuario
Write-Host "2. Registrando nuevo usuario..." -ForegroundColor Yellow
$registerData = @{
    username = "admin"
    email = "admin@sistal.com"
    password = "Admin123!"
    rut = "12345678-9"
    id_rol = 1
}

$registerResponse = Invoke-APIRequest -Method POST -Endpoint "/auth/register" -Body $registerData
if ($registerResponse) {
    Write-Host "✓ Usuario registrado exitosamente:" -ForegroundColor Green
    Write-Host ($registerResponse | ConvertTo-Json) -ForegroundColor Gray
    $userId = $registerResponse.user_id
}
else {
    Write-Host "⚠ El usuario ya existe o hubo un error" -ForegroundColor Yellow
}
Write-Host ""

# 3. Login con el usuario
Write-Host "3. Realizando login..." -ForegroundColor Yellow
$loginData = @{
    identifier = "admin"
    password = "Admin123!"
}

$loginResponse = Invoke-APIRequest -Method POST -Endpoint "/auth/login" -Body $loginData
if ($loginResponse) {
    Write-Host "✓ Login exitoso:" -ForegroundColor Green
    Write-Host "Token: $($loginResponse.token.Substring(0, 50))..." -ForegroundColor Gray
    Write-Host "Usuario: $($loginResponse.user.username)" -ForegroundColor Gray
    Write-Host "Email: $($loginResponse.user.email)" -ForegroundColor Gray
    Write-Host "Rol: $($loginResponse.user.id_rol)" -ForegroundColor Gray
    $token = $loginResponse.token
}
else {
    Write-Host "✗ Login falló" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 4. Validar token
Write-Host "4. Validando token..." -ForegroundColor Yellow
try {
    $validateResponse = Invoke-RestMethod -Uri "$baseUrl/auth/validate" -Method GET -Headers @{
        "Authorization" = "Bearer $token"
    }
    Write-Host "✓ Token válido:" -ForegroundColor Green
    Write-Host ($validateResponse | ConvertTo-Json) -ForegroundColor Gray
}
catch {
    Write-Host "✗ Validación de token falló" -ForegroundColor Red
}
Write-Host ""

# 5. Obtener información del usuario actual
Write-Host "5. Obteniendo información del usuario actual..." -ForegroundColor Yellow
try {
    $meResponse = Invoke-RestMethod -Uri "$baseUrl/auth/me" -Method GET -Headers @{
        "Authorization" = "Bearer $token"
    }
    Write-Host "✓ Información del usuario obtenida:" -ForegroundColor Green
    Write-Host ($meResponse | ConvertTo-Json) -ForegroundColor Gray
}
catch {
    Write-Host "✗ Error al obtener información del usuario" -ForegroundColor Red
}
Write-Host ""

# 6. Acceder a endpoint protegido
Write-Host "6. Accediendo a endpoint protegido (/api/profile)..." -ForegroundColor Yellow
try {
    $profileResponse = Invoke-RestMethod -Uri "$baseUrl/api/profile" -Method GET -Headers @{
        "Authorization" = "Bearer $token"
    }
    Write-Host "✓ Acceso al perfil exitoso:" -ForegroundColor Green
    Write-Host ($profileResponse | ConvertTo-Json) -ForegroundColor Gray
}
catch {
    Write-Host "✗ Error al acceder al perfil" -ForegroundColor Red
}
Write-Host ""

# 7. Probar login con email
Write-Host "7. Probando login con email..." -ForegroundColor Yellow
$loginEmailData = @{
    identifier = "admin@sistal.com"
    password = "Admin123!"
}

$loginEmailResponse = Invoke-APIRequest -Method POST -Endpoint "/auth/login" -Body $loginEmailData
if ($loginEmailResponse) {
    Write-Host "✓ Login con email exitoso" -ForegroundColor Green
}
else {
    Write-Host "✗ Login con email falló" -ForegroundColor Red
}
Write-Host ""

# 8. Probar login con RUT
Write-Host "8. Probando login con RUT..." -ForegroundColor Yellow
$loginRutData = @{
    identifier = "12345678-9"
    password = "Admin123!"
}

$loginRutResponse = Invoke-APIRequest -Method POST -Endpoint "/auth/login" -Body $loginRutData
if ($loginRutResponse) {
    Write-Host "✓ Login con RUT exitoso" -ForegroundColor Green
}
else {
    Write-Host "✗ Login con RUT falló" -ForegroundColor Red
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Pruebas Completadas" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Para ver la documentación Swagger, abre:" -ForegroundColor Yellow
Write-Host "http://localhost:8080/swagger/index.html" -ForegroundColor Cyan
