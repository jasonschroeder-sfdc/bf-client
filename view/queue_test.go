package view

import (
	"testing"
)

func TestNumValueString(t *testing.T) {
	testCases := []struct {
		name           string
		numValue       numValue
		expectedOutput string
	}{
		{
			name:           "simple integer value",
			numValue:       numValue{fmt: "Value: %v", value: 42},
			expectedOutput: "Value: 42",
		},
		{
			name:           "zero value",
			numValue:       numValue{fmt: "Count: %v", value: 0},
			expectedOutput: "Count: 0",
		},
		{
			name:           "negative value",
			numValue:       numValue{fmt: "Items: %v", value: -10},
			expectedOutput: "Items: -10",
		},
		{
			name:           "format with no placeholders",
			numValue:       numValue{fmt: "No placeholders", value: 5},
			expectedOutput: "No placeholders",
		},
		{
			name:           "label with value",
			numValue:       numValue{fmt: "Workers: %v", value: 25},
			expectedOutput: "Workers: 25",
		},
		{
			name:           "with mode and parent",
			numValue:       numValue{fmt: "Queue: %v", value: 15, mode: 2, parent: &numValue{fmt: "Parent: %v", value: 100}},
			expectedOutput: "Queue: 15",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.numValue.String()
			if result != tc.expectedOutput {
				t.Errorf("Expected output %q, got %q", tc.expectedOutput, result)
			}
		})
	}
}

// TestNumValueStringEdgeCases tests edge cases for the String method
func TestNumValueStringEdgeCases(t *testing.T) {
	// Test with multiple format specifiers (only the first should be used)
	nv := numValue{
		fmt:   "Values: %v and %v",
		value: 10,
	}
	result := nv.String()
	expected := "Values: 10 and %!v(MISSING)"
	if result != expected {
		t.Errorf("Multiple format test: expected %q, got %q", expected, result)
	}

	// Test with empty format string
	nv = numValue{
		fmt:   "",
		value: 42,
	}
	result = nv.String()
	expected = ""
	if result != expected {
		t.Errorf("Empty format test: expected %q, got %q", expected, result)
	}

	// Test with different format specifier
	nv = numValue{
		fmt:   "Hex: %x",
		value: 255,
	}
	result = nv.String()
	expected = "Hex: ff"
	if result != expected {
		t.Errorf("Hex format test: expected %q, got %q", expected, result)
	}
} 