package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// ValidateRUT valida un RUT chileno con su dígito verificador
// Formato esperado: 12345678-9 o 12.345.678-9
func ValidateRUT(rut string) bool {
	// Eliminar puntos y guiones para procesar
	cleanRUT := strings.ReplaceAll(rut, ".", "")
	cleanRUT = strings.ReplaceAll(cleanRUT, "-", "")

	// Verificar formato: al menos 2 caracteres (número + dígito verificador)
	if len(cleanRUT) < 2 {
		return false
	}

	// Separar número y dígito verificador
	rutNumber := cleanRUT[:len(cleanRUT)-1]
	expectedDV := strings.ToUpper(cleanRUT[len(cleanRUT)-1:])

	// Verificar que el número sea numérico
	_, err := strconv.Atoi(rutNumber)
	if err != nil {
		return false
	}

	// Calcular dígito verificador
	calculatedDV := calculateRUTDV(rutNumber)

	return calculatedDV == expectedDV
}

// calculateRUTDV calcula el dígito verificador de un RUT chileno
func calculateRUTDV(rutNumber string) string {
	// Convertir a entero
	num, err := strconv.Atoi(rutNumber)
	if err != nil {
		return ""
	}

	// Algoritmo del módulo 11
	multiplier := 2
	sum := 0

	for num > 0 {
		digit := num % 10
		sum += digit * multiplier
		multiplier++
		if multiplier > 7 {
			multiplier = 2
		}
		num = num / 10
	}

	remainder := sum % 11
	dv := 11 - remainder

	// Convertir el resultado al dígito verificador
	switch dv {
	case 11:
		return "0"
	case 10:
		return "K"
	default:
		return strconv.Itoa(dv)
	}
}

// FormatRUT formatea un RUT al formato estándar con guión (sin puntos)
// Ejemplo: 12345678-9
func FormatRUT(rut string) string {
	// Eliminar puntos y guiones
	cleanRUT := strings.ReplaceAll(rut, ".", "")
	cleanRUT = strings.ReplaceAll(cleanRUT, "-", "")
	cleanRUT = strings.ToUpper(cleanRUT)

	if len(cleanRUT) < 2 {
		return rut
	}

	// Separar número y dígito verificador
	rutNumber := cleanRUT[:len(cleanRUT)-1]
	dv := cleanRUT[len(cleanRUT)-1:]

	return rutNumber + "-" + dv
}

// ValidateRUTFormat valida el formato de un RUT (con o sin validación de DV)
// Acepta formatos: 12345678-9, 12.345.678-9, 123456789
func ValidateRUTFormat(rut string) bool {
	// Patrón para RUT con guión: 12345678-9 o 12.345.678-9
	pattern := `^(\d{1,2}\.?\d{3}\.?\d{3}|\d{7,8})-[\dkK]$`
	matched, _ := regexp.MatchString(pattern, rut)
	if matched {
		return true
	}

	// Patrón para RUT sin guión pero válido numéricamente
	pattern = `^\d{7,9}$`
	matched, _ = regexp.MatchString(pattern, rut)
	return matched
}

// MinPasswordLength constante para la longitud mínima de contraseña según contrato
const MinPasswordLength = 8
