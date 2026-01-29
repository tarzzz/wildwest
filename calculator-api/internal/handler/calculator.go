package handler

import (
	"calculator-api/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CalculateRequest represents the request body for calculation
type CalculateRequest struct {
	Operation string  `json:"operation" binding:"required"`
	Operand1  float64 `json:"operand1" binding:"required"`
	Operand2  float64 `json:"operand2" binding:"required"`
}

// CalculateResponse represents the response body for calculation
type CalculateResponse struct {
	Result    float64   `json:"result"`
	Operation string    `json:"operation"`
	Operands  []float64 `json:"operands"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CalculatorHandler handles calculator operations
type CalculatorHandler struct {
	calculator *service.Calculator
}

// NewCalculatorHandler creates a new CalculatorHandler
func NewCalculatorHandler(calculator *service.Calculator) *CalculatorHandler {
	return &CalculatorHandler{
		calculator: calculator,
	}
}

// Calculate handles POST /api/v1/calculate requests
func (h *CalculatorHandler) Calculate(c *gin.Context) {
	var req CalculateRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_INPUT",
				Message: "Missing or invalid fields in request body",
			},
		})
		return
	}

	var result float64
	var err error

	// Execute the appropriate operation
	switch req.Operation {
	case "add":
		result = h.calculator.Add(req.Operand1, req.Operand2)
	case "subtract":
		result = h.calculator.Subtract(req.Operand1, req.Operand2)
	case "multiply":
		result = h.calculator.Multiply(req.Operand1, req.Operand2)
	case "divide":
		result, err = h.calculator.Divide(req.Operand1, req.Operand2)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: ErrorDetail{
					Code:    "DIVISION_BY_ZERO",
					Message: "Cannot divide by zero",
				},
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: ErrorDetail{
				Code:    "INVALID_OPERATION",
				Message: "Operation must be one of: add, subtract, multiply, divide",
			},
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, CalculateResponse{
		Result:    result,
		Operation: req.Operation,
		Operands:  []float64{req.Operand1, req.Operand2},
	})
}
