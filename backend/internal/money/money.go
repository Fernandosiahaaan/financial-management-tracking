package money

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Cents represents monetary amounts stored as integer sen/cents (1 IDR = 100 sen).
// It implements json.Unmarshaler to seamlessly accept either numeric cents (e.g. 10000000)
// or formatted Indonesian strings (e.g. "100.000,00", "100.000", "Rp 100.000,00").
type Cents int64

// Int64 returns the underlying integer value in cents.
func (c Cents) Int64() int64 {
	return int64(c)
}

// UnmarshalJSON unmarshals either a JSON number or formatted string into Cents.
func (c *Cents) UnmarshalJSON(b []byte) error {
	trimmed := strings.TrimSpace(string(b))
	if trimmed == "" || trimmed == "null" {
		*c = 0
		return nil
	}

	// Case 1: Numeric value
	if trimmed[0] != '"' {
		// Could be integer or float
		if strings.Contains(trimmed, ".") {
			f, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return fmt.Errorf("invalid money number: %s", trimmed)
			}
			*c = Cents(int64(f))
			return nil
		}
		val, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid money integer: %s", trimmed)
		}
		*c = Cents(val)
		return nil
	}

	// Case 2: String value
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	cents, err := ParseIDRToCents(s)
	if err != nil {
		return err
	}
	*c = Cents(cents)
	return nil
}

// MarshalJSON marshals Cents as a standard JSON integer.
func (c Cents) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(c), 10)), nil
}

// ParseIDRToCents parses formatted string or plain string into cents (sen).
// Examples:
// - "100.000,00" -> 10000000
// - "100.000,50" -> 10000050
// - "100.000"    -> 10000000
// - "Rp 100.000,00" -> 10000000
// - "100000"     -> 10000000
// - "0"          -> 0
func ParseIDRToCents(s string) (int64, error) {
	clean := strings.TrimSpace(s)
	if clean == "" {
		return 0, nil
	}

	// Strip "Rp", "IDR", currency symbols, and extra spaces
	clean = strings.TrimPrefix(clean, "Rp")
	clean = strings.TrimPrefix(clean, "rp")
	clean = strings.TrimPrefix(clean, "IDR")
	clean = strings.TrimPrefix(clean, "idr")
	clean = strings.TrimSpace(clean)

	if clean == "" {
		return 0, nil
	}

	// Check for negative sign
	isNegative := false
	if strings.HasPrefix(clean, "-") {
		isNegative = true
		clean = strings.TrimPrefix(clean, "-")
		clean = strings.TrimSpace(clean)
	}

	// Check if comma is used as decimal separator (standard Indonesian format: 100.000,00)
	var wholePart string
	var decimalPart string

	if strings.Contains(clean, ",") {
		parts := strings.Split(clean, ",")
		if len(parts) > 2 {
			return 0, fmt.Errorf("invalid money format with multiple commas: %s", s)
		}
		wholePart = parts[0]
		decimalPart = parts[1]
	} else if strings.Contains(clean, ".") && !strings.Contains(clean, ",") {
		// Check if the dot is a decimal separator (e.g. "100000.50") or thousand separator ("100.000")
		dotParts := strings.Split(clean, ".")
		if len(dotParts) == 2 && len(dotParts[1]) <= 2 {
			// e.g. "100000.50" or "50.5" -> dot is decimal
			wholePart = dotParts[0]
			decimalPart = dotParts[1]
		} else {
			// e.g. "100.000" or "1.000.000" -> dot is thousand separator
			wholePart = clean
			decimalPart = "00"
		}
	} else {
		wholePart = clean
		decimalPart = "00"
	}

	// Strip non-digit characters (like thousand dots) from wholePart
	var wholeDigits strings.Builder
	for _, r := range wholePart {
		if unicode.IsDigit(r) {
			wholeDigits.WriteRune(r)
		}
	}

	// Strip non-digits from decimalPart
	var decDigits strings.Builder
	for _, r := range decimalPart {
		if unicode.IsDigit(r) {
			decDigits.WriteRune(r)
		}
	}

	wholeStr := wholeDigits.String()
	if wholeStr == "" {
		wholeStr = "0"
	}

	decStr := decDigits.String()
	if len(decStr) == 0 {
		decStr = "00"
	} else if len(decStr) == 1 {
		decStr = decStr + "0"
	} else if len(decStr) > 2 {
		decStr = decStr[:2]
	}

	whole, err := strconv.ParseInt(wholeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid whole amount in money string %q: %w", s, err)
	}

	dec, err := strconv.ParseInt(decStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid decimal amount in money string %q: %w", s, err)
	}

	totalCents := (whole * 100) + dec
	if isNegative {
		totalCents = -totalCents
	}

	return totalCents, nil
}

// FormatIDR formats cents into standard Indonesian currency with "Rp " prefix and ",00"
// e.g. 10000000 cents -> "Rp 100.000,00"
// e.g. -500000 cents  -> "-Rp 5.000,00"
func FormatIDR(cents int64) string {
	formattedNum := FormatNumberIDR(cents)
	if cents < 0 {
		return "-Rp " + strings.TrimPrefix(formattedNum, "-")
	}
	return "Rp " + formattedNum
}

// FormatNumberIDR formats cents into "xxx.xxx,00" format without currency prefix.
// e.g. 10000000 cents -> "100.000,00"
// e.g. 10000050 cents -> "100.000,50"
func FormatNumberIDR(cents int64) string {
	isNegative := cents < 0
	absCents := cents
	if isNegative {
		absCents = -cents
	}

	whole := absCents / 100
	dec := absCents % 100

	// Format whole part with dots
	wholeStr := strconv.FormatInt(whole, 10)
	var formattedWhole strings.Builder
	n := len(wholeStr)
	for i, r := range wholeStr {
		if i > 0 && (n-i)%3 == 0 {
			formattedWhole.WriteRune('.')
		}
		formattedWhole.WriteRune(r)
	}

	res := fmt.Sprintf("%s,%02d", formattedWhole.String(), dec)
	if isNegative {
		res = "-" + res
	}
	return res
}
