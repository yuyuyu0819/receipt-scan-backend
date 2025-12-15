package db

import (
	"testing"
	"time"
)

func TestParsePurchaseDate(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "date only",
			input:    "2025-12-09",
			expected: time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "datetime without timezone",
			input:    "2025-12-09T19:52:00",
			expected: time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "datetime with timezone",
			input:    "2025-12-09T19:52:00Z",
			expected: time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid format",
			input:   "12/09/2025",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePurchaseDate(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !got.Equal(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}
