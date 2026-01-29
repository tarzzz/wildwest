package service

import (
	"testing"
)

func TestCalculator_Add(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"positive numbers", 5.0, 3.0, 8.0},
		{"negative numbers", -5.0, -3.0, -8.0},
		{"mixed signs", 5.0, -3.0, 2.0},
		{"with zero", 5.0, 0.0, 5.0},
		{"decimals", 5.5, 3.3, 8.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Subtract(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"positive numbers", 10.0, 4.0, 6.0},
		{"negative numbers", -5.0, -3.0, -2.0},
		{"mixed signs", 5.0, -3.0, 8.0},
		{"with zero", 5.0, 0.0, 5.0},
		{"decimals", 5.5, 3.3, 2.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Subtract(tt.a, tt.b)
			// Use approximate comparison for floating point
			if !floatEquals(result, tt.expected) {
				t.Errorf("Subtract(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Multiply(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		a        float64
		b        float64
		expected float64
	}{
		{"positive numbers", 6.0, 7.0, 42.0},
		{"negative numbers", -5.0, -3.0, 15.0},
		{"mixed signs", 5.0, -3.0, -15.0},
		{"with zero", 5.0, 0.0, 0.0},
		{"decimals", 2.5, 4.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Divide(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name      string
		a         float64
		b         float64
		expected  float64
		expectErr bool
	}{
		{"positive numbers", 15.0, 3.0, 5.0, false},
		{"negative numbers", -15.0, -3.0, 5.0, false},
		{"mixed signs", 15.0, -3.0, -5.0, false},
		{"decimals", 10.0, 4.0, 2.5, false},
		{"division by zero", 10.0, 0.0, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Divide(tt.a, tt.b)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Divide(%v, %v) expected error but got none", tt.a, tt.b)
				}
			} else {
				if err != nil {
					t.Errorf("Divide(%v, %v) unexpected error: %v", tt.a, tt.b, err)
				}
				if result != tt.expected {
					t.Errorf("Divide(%v, %v) = %v, expected %v", tt.a, tt.b, result, tt.expected)
				}
			}
		})
	}
}

// floatEquals checks if two floats are approximately equal
func floatEquals(a, b float64) bool {
	tolerance := 0.00001
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}
