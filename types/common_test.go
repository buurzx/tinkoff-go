package types

import (
	"testing"
)

func TestMoneyValue_ToFloat(t *testing.T) {
	tests := []struct {
		name     string
		mv       *MoneyValue
		expected float64
	}{
		{
			name:     "positive value",
			mv:       &MoneyValue{Units: 100, Nano: 500000000, Currency: "rub"},
			expected: 100.5,
		},
		{
			name:     "zero value",
			mv:       &MoneyValue{Units: 0, Nano: 0, Currency: "rub"},
			expected: 0.0,
		},
		{
			name:     "negative value",
			mv:       &MoneyValue{Units: -50, Nano: -250000000, Currency: "rub"},
			expected: -50.25,
		},
		{
			name:     "small decimal",
			mv:       &MoneyValue{Units: 0, Nano: 1, Currency: "rub"},
			expected: 0.000000001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.ToFloat()
			if result != tt.expected {
				t.Errorf("ToFloat() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewMoneyValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		currency string
		expected *MoneyValue
	}{
		{
			name:     "positive value",
			value:    123.45,
			currency: "rub",
			expected: &MoneyValue{Units: 123, Nano: 450000000, Currency: "rub"},
		},
		{
			name:     "zero value",
			value:    0.0,
			currency: "usd",
			expected: &MoneyValue{Units: 0, Nano: 0, Currency: "usd"},
		},
		{
			name:     "negative value",
			value:    -67.89,
			currency: "eur",
			expected: &MoneyValue{Units: -67, Nano: -890000000, Currency: "eur"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMoneyValue(tt.value, tt.currency)
			if result.Units != tt.expected.Units {
				t.Errorf("NewMoneyValue() Units = %v, expected %v", result.Units, tt.expected.Units)
			}
			if result.Currency != tt.expected.Currency {
				t.Errorf("NewMoneyValue() Currency = %v, expected %v", result.Currency, tt.expected.Currency)
			}
			// Allow some tolerance for nano precision
			if abs(result.Nano-tt.expected.Nano) > 1000 {
				t.Errorf("NewMoneyValue() Nano = %v, expected %v", result.Nano, tt.expected.Nano)
			}
		})
	}
}

func TestQuotation_ToFloat(t *testing.T) {
	tests := []struct {
		name     string
		q        *Quotation
		expected float64
	}{
		{
			name:     "positive quotation",
			q:        &Quotation{Units: 250, Nano: 750000000},
			expected: 250.75,
		},
		{
			name:     "zero quotation",
			q:        &Quotation{Units: 0, Nano: 0},
			expected: 0.0,
		},
		{
			name:     "negative quotation",
			q:        &Quotation{Units: -10, Nano: -500000000},
			expected: -10.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.q.ToFloat()
			if result != tt.expected {
				t.Errorf("ToFloat() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestNewQuotation(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected *Quotation
	}{
		{
			name:     "positive value",
			value:    275.25,
			expected: &Quotation{Units: 275, Nano: 250000000},
		},
		{
			name:     "zero value",
			value:    0.0,
			expected: &Quotation{Units: 0, Nano: 0},
		},
		{
			name:     "negative value",
			value:    -15.75,
			expected: &Quotation{Units: -15, Nano: -750000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewQuotation(tt.value)
			if result.Units != tt.expected.Units {
				t.Errorf("NewQuotation() Units = %v, expected %v", result.Units, tt.expected.Units)
			}
			// Allow some tolerance for nano precision
			if abs(result.Nano-tt.expected.Nano) > 1000 {
				t.Errorf("NewQuotation() Nano = %v, expected %v", result.Nano, tt.expected.Nano)
			}
		})
	}
}

func TestMoneyValue_String(t *testing.T) {
	tests := []struct {
		name     string
		mv       *MoneyValue
		expected string
	}{
		{
			name:     "positive value",
			mv:       &MoneyValue{Units: 1000, Nano: 500000000, Currency: "rub"},
			expected: "1000.50 rub",
		},
		{
			name:     "zero value",
			mv:       &MoneyValue{Units: 0, Nano: 0, Currency: "usd"},
			expected: "0.00 usd",
		},
		{
			name:     "negative value",
			mv:       &MoneyValue{Units: -250, Nano: -750000000, Currency: "eur"},
			expected: "-250.75 eur",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mv.String()
			if result != tt.expected {
				t.Errorf("String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestQuotation_String(t *testing.T) {
	tests := []struct {
		name     string
		q        *Quotation
		expected string
	}{
		{
			name:     "positive quotation",
			q:        &Quotation{Units: 275, Nano: 250000000},
			expected: "275.2500",
		},
		{
			name:     "zero quotation",
			q:        &Quotation{Units: 0, Nano: 0},
			expected: "0.0000",
		},
		{
			name:     "negative quotation",
			q:        &Quotation{Units: -15, Nano: -750000000},
			expected: "-15.7500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.q.String()
			if result != tt.expected {
				t.Errorf("String() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:    "INVALID_ARGUMENT",
		Message: "Invalid token provided",
		Details: "Token format is incorrect",
	}

	expected := "tinkoff api error: INVALID_ARGUMENT - Invalid token provided"
	result := err.Error()

	if result != expected {
		t.Errorf("Error() = %v, expected %v", result, expected)
	}
}

// Helper function for absolute value
func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
