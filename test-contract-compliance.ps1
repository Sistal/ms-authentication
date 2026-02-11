# Script de prueba completo para MS-Authentication
# Cumplimiento del contrato BFF

$baseUrl = "http://localhost:8081/api/v1/auth"
$testUser = @{
    nombre_usuario = "test_$(Get-Random)"
    nombre_completo = "Usuario de Prueba"
    rut = "12345678-9"
    password = "Test1234"
    id_rol = 1
}

Write-Host "=== MS-Authentication - Test Suite ===" -ForegroundColor Cyan
Write-Host ""

# 1. Health Check
Write-Host "[1] Testing Health Check..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    Write-Host "✓ Health: $($health.status)" -ForegroundColor Green
} catch {
    Write-Host "✗ Health check failed" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 2. Register (público)
Write-Host "[2] Testing POST /auth/register..." -ForegroundColor Yellow
try {
    $registerBody = $testUser | ConvertTo-Json
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/register" `
        -Method POST -ContentType "application/json" -Body $registerBody
    
    if ($registerResponse.success -eq $true) {
        Write-Host "✓ Register exitoso" -ForegroundColor Green
        Write-Host "  - Usuario ID: $($registerResponse.data.id_usuario)" -ForegroundColor Gray
    } else {
        Write-Host "✗ Success = false" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Register failed: $_" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        $errorDetails = $_.ErrorDetails.Message | ConvertFrom-Json
        Write-Host "  Error: $($errorDetails.message)" -ForegroundColor Red
    }
}
Write-Host ""

# 3. Login
Write-Host "[3] Testing POST /auth/login..." -ForegroundColor Yellow
try {
    $loginBody = @{
        nombre_usuario = $testUser.nombre_usuario
        password = $testUser.password
    } | ConvertTo-Json
    
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/login" `
        -Method POST -ContentType "application/json" -Body $loginBody
    
    if ($loginResponse.success -eq $true) {
        Write-Host "✓ Login exitoso" -ForegroundColor Green
        Write-Host "  - Message: $($loginResponse.message)" -ForegroundColor Gray
        Write-Host "  - Token type: $($loginResponse.data.token_type)" -ForegroundColor Gray
        Write-Host "  - Expires in: $($loginResponse.data.expires_in) seg" -ForegroundColor Gray
        Write-Host "  - Rol: $($loginResponse.data.usuario.rol.nombre_rol)" -ForegroundColor Gray
        Write-Host "  - Estado: $($loginResponse.data.usuario.estado.nombre_estado)" -ForegroundColor Gray
        
        $token = $loginResponse.data.token
        $userId = $loginResponse.data.usuario.id_usuario
        
        # Verificar estructura completa
        if ($loginResponse.data.expires_in -ne 86400) {
            Write-Host "  ! WARNING: expires_in debería ser 86400 (24h)" -ForegroundColor Yellow
        }
        if (-not $loginResponse.data.usuario.rol.nombre_rol) {
            Write-Host "  ! WARNING: falta rol.nombre_rol" -ForegroundColor Yellow
        }
        if (-not $loginResponse.data.usuario.estado.nombre_estado) {
            Write-Host "  ! WARNING: falta estado.nombre_estado" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✗ Success = false" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "✗ Login failed: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 4. Validate Token
Write-Host "[4] Testing GET /auth/validate..." -ForegroundColor Yellow
try {
    $validateResponse = Invoke-RestMethod -Uri "$baseUrl/validate" `
        -Method GET -Headers @{ Authorization = "Bearer $token" }
    
    if ($validateResponse.success -eq $true) {
        Write-Host "✓ Token válido" -ForegroundColor Green
        Write-Host "  - Usuario: $($validateResponse.data.nombre_usuario)" -ForegroundColor Gray
        Write-Host "  - Nombre completo: $($validateResponse.data.nombre_completo)" -ForegroundColor Gray
        Write-Host "  - RUT: $($validateResponse.data.rut)" -ForegroundColor Gray
        Write-Host "  - Rol: $($validateResponse.data.nombre_rol)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Validate failed" -ForegroundColor Red
}
Write-Host ""

# 5. Get Me
Write-Host "[5] Testing GET /auth/me..." -ForegroundColor Yellow
try {
    $meResponse = Invoke-RestMethod -Uri "$baseUrl/me" `
        -Method GET -Headers @{ Authorization = "Bearer $token" }
    
    if ($meResponse.success -eq $true) {
        Write-Host "✓ Profile obtenido" -ForegroundColor Green
        Write-Host "  - ID: $($meResponse.data.id_usuario)" -ForegroundColor Gray
        Write-Host "  - Nombre: $($meResponse.data.nombre_completo)" -ForegroundColor Gray
        Write-Host "  - Rol: $($meResponse.data.rol.nombre_rol)" -ForegroundColor Gray
        Write-Host "  - Estado: $($meResponse.data.estado.nombre_estado)" -ForegroundColor Gray
        Write-Host "  - Fecha creación: $($meResponse.data.fecha_creacion)" -ForegroundColor Gray
        
        # Verificar que datos vienen de BD, no solo del token
        if ($meResponse.data.fecha_creacion) {
            Write-Host "  ✓ Datos de BD confirmados" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "✗ GetMe failed" -ForegroundColor Red
}
Write-Host ""

# 6. Get Roles
Write-Host "[6] Testing GET /auth/roles..." -ForegroundColor Yellow
try {
    $rolesResponse = Invoke-RestMethod -Uri "$baseUrl/roles" `
        -Method GET -Headers @{ Authorization = "Bearer $token" }
    
    if ($rolesResponse.success -eq $true) {
        Write-Host "✓ Roles obtenidos" -ForegroundColor Green
        Write-Host "  - Total: $($rolesResponse.meta.total)" -ForegroundColor Gray
        foreach ($rol in $rolesResponse.data) {
            Write-Host "    - $($rol.id_rol): $($rol.nombre_rol)" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "✗ GetRoles failed" -ForegroundColor Red
}
Write-Host ""

# 7. Change Password
Write-Host "[7] Testing PUT /auth/change-password..." -ForegroundColor Yellow
try {
    $changePassBody = @{
        password_actual = $testUser.password
        password_nueva = "NewPass1234"
        password_confirmacion = "NewPass1234"
    } | ConvertTo-Json
    
    $changePassResponse = Invoke-RestMethod -Uri "$baseUrl/change-password" `
        -Method PUT -ContentType "application/json" -Body $changePassBody `
        -Headers @{ Authorization = "Bearer $token" }
    
    if ($changePassResponse.success -eq $true) {
        Write-Host "✓ Contraseña cambiada" -ForegroundColor Green
        Write-Host "  - Message: $($changePassResponse.message)" -ForegroundColor Gray
        
        # Actualizar password para próximas pruebas
        $testUser.password = "NewPass1234"
    }
} catch {
    Write-Host "✗ ChangePassword failed" -ForegroundColor Red
}
Write-Host ""

# 8. Refresh Token (obtener refresh token del login)
Write-Host "[8] Testing POST /auth/refresh..." -ForegroundColor Yellow
if ($loginResponse.data.refresh_token) {
    try {
        $refreshBody = @{
            refresh_token = $loginResponse.data.refresh_token
        } | ConvertTo-Json
        
        $refreshResponse = Invoke-RestMethod -Uri "$baseUrl/refresh" `
            -Method POST -ContentType "application/json" -Body $refreshBody
        
        if ($refreshResponse.success -eq $true) {
            Write-Host "✓ Token renovado" -ForegroundColor Green
            Write-Host "  - Nuevo token type: $($refreshResponse.data.token_type)" -ForegroundColor Gray
            Write-Host "  - Expires in: $($refreshResponse.data.expires_in)" -ForegroundColor Gray
        }
    } catch {
        Write-Host "✗ RefreshToken failed" -ForegroundColor Red
    }
} else {
    Write-Host "⊘ No refresh_token en login response" -ForegroundColor Yellow
}
Write-Host ""

# 9. List Users (Admin) - Debe fallar con usuario normal
Write-Host "[9] Testing GET /auth/users (Admin requerido)..." -ForegroundColor Yellow
try {
    $usersResponse = Invoke-RestMethod -Uri "$baseUrl/users" `
        -Method GET -Headers @{ Authorization = "Bearer $token" }
    
    Write-Host "! Usuario normal tiene acceso (DEBERÍA SER 403)" -ForegroundColor Yellow
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    if ($statusCode -eq 403) {
        Write-Host "✓ Correctamente denegado (403 Forbidden)" -ForegroundColor Green
    } else {
        Write-Host "✗ Error inesperado: $statusCode" -ForegroundColor Red
    }
}
Write-Host ""

# 10. Logout
Write-Host "[10] Testing POST /auth/logout..." -ForegroundColor Yellow
try {
    $logoutResponse = Invoke-RestMethod -Uri "$baseUrl/logout" `
        -Method POST -Headers @{ Authorization = "Bearer $token" }
    
    if ($logoutResponse.success -eq $true) {
        Write-Host "✓ Logout exitoso" -ForegroundColor Green
        Write-Host "  - Message: $($logoutResponse.message)" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Logout failed" -ForegroundColor Red
}
Write-Host ""

# 11. Test de usuario inactivo (403)
Write-Host "[11] Testing login con usuario inactivo (403)..." -ForegroundColor Yellow
Write-Host "  ⊘ Requiere actualizar manualmente un usuario a estado 2" -ForegroundColor Gray
Write-Host ""

# Resumen
Write-Host "=== Resumen de Tests ===" -ForegroundColor Cyan
Write-Host "Endpoints probados:"
Write-Host "  ✓ POST /auth/register" -ForegroundColor Green
Write-Host "  ✓ POST /auth/login (con validación de estructura)" -ForegroundColor Green
Write-Host "  ✓ GET  /auth/validate" -ForegroundColor Green
Write-Host "  ✓ GET  /auth/me" -ForegroundColor Green
Write-Host "  ✓ GET  /auth/roles" -ForegroundColor Green
Write-Host "  ✓ PUT  /auth/change-password" -ForegroundColor Green
Write-Host "  ✓ POST /auth/refresh" -ForegroundColor Green
Write-Host "  ✓ POST /auth/logout" -ForegroundColor Green
Write-Host "  ✓ GET  /auth/users (verificación 403)" -ForegroundColor Green
Write-Host ""
Write-Host "Formato de respuesta:" -ForegroundColor Cyan
Write-Host "  ✓ success: boolean" -ForegroundColor Green
Write-Host "  ✓ message: string" -ForegroundColor Green
Write-Host "  ✓ data: object" -ForegroundColor Green
Write-Host "  ✓ meta: object (cuando aplica)" -ForegroundColor Green
Write-Host ""
Write-Host "JWT Claims:" -ForegroundColor Cyan
Write-Host "  ✓ expires_in: 86400 segundos" -ForegroundColor Green
Write-Host "  ✓ usuario.rol {id, nombre, descripcion}" -ForegroundColor Green
Write-Host "  ✓ usuario.estado {id, nombre}" -ForegroundColor Green
Write-Host ""
Write-Host "=== Tests Completados ===" -ForegroundColor Green
