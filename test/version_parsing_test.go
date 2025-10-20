package test

import (
	"strconv"
	"strings"
	"testing"
)

func TestVersionParsingWithPrerelease(t *testing.T) {
	testCases := []struct {
		version       string
		expectedMajor int
		expectedMinor int
		expectedPatch int
	}{
		{"25.1.0-prerelease", 25, 1, 0},
		{"24.10.5", 24, 10, 5},
		{"23.0.1-rc1", 23, 0, 1},
		{"1.2.3-alpha.1", 1, 2, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			parts := strings.Split(tc.version, ".")

			var major, minor, patch int

			for idx, value := range parts {
				if idx > 2 {
					break
				}

				if hyphenIdx := strings.Index(value, "-"); hyphenIdx != -1 {
					value = value[:hyphenIdx]
				}

				num, err := strconv.Atoi(value)
				if err != nil {
					t.Fatalf("Failed to parse version component %q: %v", value, err)
				}

				switch idx {
				case 0:
					major = num
				case 1:
					minor = num
				default:
					patch = num
				}
			}

			if major != tc.expectedMajor {
				t.Errorf("Expected major=%d, got %d", tc.expectedMajor, major)
			}
			if minor != tc.expectedMinor {
				t.Errorf("Expected minor=%d, got %d", tc.expectedMinor, minor)
			}
			if patch != tc.expectedPatch {
				t.Errorf("Expected patch=%d, got %d", tc.expectedPatch, patch)
			}
		})
	}
}
