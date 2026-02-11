package entities

// APIResponse estructura genérica para todas las respuestas de la API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorDetail estructura para errores con código específico
type ErrorDetail struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// ValidationError estructura para errores de validación por campo
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse crea una respuesta exitosa estándar
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// SuccessResponseWithMeta crea una respuesta exitosa con metadatos
func SuccessResponseWithMeta(message string, data interface{}, meta interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
}

// ErrorResponse crea una respuesta de error con código específico
func ErrorResponse(message string, errorDetail ErrorDetail) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Errors:  errorDetail,
	}
}

// ErrorResponseSimple crea una respuesta de error simple
func ErrorResponseSimple(message string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
	}
}

// ValidationErrorResponse crea una respuesta de error de validación
func ValidationErrorResponse(message string, errors []ValidationError) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}
