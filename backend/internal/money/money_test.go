package money

import (
	"encoding/json"
	"testing"
)

func TestParseIDRToCents(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"100.000,00", 10000000},
		{"100.000,50", 10000050},
		{"100.000", 10000000},
		{"Rp 100.000,00", 10000000},
		{"Rp. 100.000,00", 10000000},
		{"100000", 10000000},
		{"100000.50", 10000050},
		{"0", 0},
		{"", 0},
		{"-50.000,00", -5000000},
		{"1.500.000,00", 150000000},
	}

	for _, tc := range tests {
		got, err := ParseIDRToCents(tc.input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.input, err)
		}
		if got != tc.expected {
			t.Errorf("for input %q, expected %d, got %d", tc.input, tc.expected, got)
		}
	}
}

func TestFormatIDR(t *testing.T) {
	tests := []struct {
		cents    int64
		expected string
	}{
		{10000000, "Rp 100.000,00"},
		{10000050, "Rp 100.000,50"},
		{0, "Rp 0,00"},
		{50000, "Rp 500,00"},
		{150000000, "Rp 1.500.000,00"},
		{-5000000, "-Rp 50.000,00"},
	}

	for _, tc := range tests {
		got := FormatIDR(tc.cents)
		if got != tc.expected {
			t.Errorf("for cents %d, expected %q, got %q", tc.cents, tc.expected, got)
		}
	}
}

func TestCentsJSONUnmarshal(t *testing.T) {
	type Payload struct {
		Amount Cents `json:"amount"`
	}

	// Test numeric integer cents
	var p1 Payload
	if err := json.Unmarshal([]byte(`{"amount": 10000000}`), &p1); err != nil {
		t.Fatal(err)
	}
	if p1.Amount != 10000000 {
		t.Errorf("expected 10000000, got %d", p1.Amount)
	}

	// Test formatted Indonesian string
	var p2 Payload
	if err := json.Unmarshal([]byte(`{"amount": "100.000,00"}`), &p2); err != nil {
		t.Fatal(err)
	}
	if p2.Amount != 10000000 {
		t.Errorf("expected 10000000, got %d", p2.Amount)
	}

	// Test formatted string with Rp
	var p3 Payload
	if err := json.Unmarshal([]byte(`{"amount": "Rp 100.000,00"}`), &p3); err != nil {
		t.Fatal(err)
	}
	if p3.Amount != 10000000 {
		t.Errorf("expected 10000000, got %d", p3.Amount)
	}
}
