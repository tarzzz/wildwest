package handler

import (
	"bytes"
	"calculator-api/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *CalculatorHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	calculator := service.NewCalculator()
	handler := NewCalculatorHandler(calculator)
	return router, handler
}

func TestCalculatorHandler_Calculate_Add(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	reqBody := CalculateRequest{
		Operation: "add",
		Operand1:  5.0,
		Operand2:  3.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response CalculateResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Result != 8.0 {
		t.Errorf("Expected result 8.0, got %v", response.Result)
	}
	if response.Operation != "add" {
		t.Errorf("Expected operation 'add', got %v", response.Operation)
	}
}

func TestCalculatorHandler_Calculate_Subtract(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	reqBody := CalculateRequest{
		Operation: "subtract",
		Operand1:  10.0,
		Operand2:  4.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response CalculateResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Result != 6.0 {
		t.Errorf("Expected result 6.0, got %v", response.Result)
	}
}

func TestCalculatorHandler_Calculate_Multiply(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	reqBody := CalculateRequest{
		Operation: "multiply",
		Operand1:  6.0,
		Operand2:  7.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response CalculateResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Result != 42.0 {
		t.Errorf("Expected result 42.0, got %v", response.Result)
	}
}

func TestCalculatorHandler_Calculate_Divide(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	reqBody := CalculateRequest{
		Operation: "divide",
		Operand1:  15.0,
		Operand2:  3.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response CalculateResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Result != 5.0 {
		t.Errorf("Expected result 5.0, got %v", response.Result)
	}
}

func TestCalculatorHandler_Calculate_DivisionByZero(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	reqBody := CalculateRequest{
		Operation: "divide",
		Operand1:  10.0,
		Operand2:  0.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "DIVISION_BY_ZERO" {
		t.Errorf("Expected error code 'DIVISION_BY_ZERO', got %v", response.Error.Code)
	}
}

func TestCalculatorHandler_Calculate_InvalidOperation(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	reqBody := CalculateRequest{
		Operation: "power",
		Operand1:  2.0,
		Operand2:  3.0,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "INVALID_OPERATION" {
		t.Errorf("Expected error code 'INVALID_OPERATION', got %v", response.Error.Code)
	}
}

func TestCalculatorHandler_Calculate_InvalidInput(t *testing.T) {
	router, handler := setupTestRouter()
	router.POST("/calculate", handler.Calculate)

	// Missing required fields
	reqBody := map[string]interface{}{
		"operation": "add",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Error.Code != "INVALID_INPUT" {
		t.Errorf("Expected error code 'INVALID_INPUT', got %v", response.Error.Code)
	}
}

func TestHealthHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHealthHandler()
	router.GET("/health", handler.Health)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response.Status)
	}
	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %v", response.Version)
	}
}
