package env

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildServicePATH(t *testing.T) {
	sep := string(filepath.ListSeparator)

	type expectation struct {
		nonEmpty         bool
		exactParts       []string
		first            string
		contains         []string
		counts           map[string]int
		noEmptyParts     bool
		maxLen           int
		maxPartsLessThan int
	}

	type testCase struct {
		name  string
		setup func(t *testing.T) (customPath string, want expectation)
	}

	tests := []testCase{
		{
			name: "custom path deduplicates and trims",
			setup: func(t *testing.T) (string, expectation) {
				base := t.TempDir()
				p1 := filepath.Join(base, "one")
				p2 := filepath.Join(base, "two")

				customPath := strings.Join([]string{" " + p1 + " ", p2, p1, "", "  "}, sep)
				return customPath, expectation{
					nonEmpty:   true,
					exactParts: []string{p1, p2},
				}
			},
		},
		{
			name: "custom path deduplicates normalized equivalents",
			setup: func(t *testing.T) (string, expectation) {
				base := t.TempDir()
				p := filepath.Join(base, "bin")
				pWithTrailingSep := p + string(filepath.Separator)
				pWithDot := p + string(filepath.Separator) + "."

				customPath := strings.Join([]string{p, pWithTrailingSep, pWithDot}, sep)
				return customPath, expectation{
					nonEmpty:   true,
					exactParts: []string{p},
				}
			},
		},
		{
			name: "default path includes home and deduplicates env PATH",
			setup: func(t *testing.T) (string, expectation) {
				home := t.TempDir()
				t.Setenv("HOME", home)

				customBin := filepath.Join(t.TempDir(), "custom-bin")
				customBinTrailingSep := customBin + string(filepath.Separator)
				customBinWithDot := customBin + string(filepath.Separator) + "."
				t.Setenv("PATH", strings.Join([]string{"/usr/bin", customBin, "/bin", customBinTrailingSep, customBinWithDot}, sep))

				return "", expectation{
					nonEmpty: true,
					first:    filepath.Join(home, ".local", "bin"),
					contains: []string{filepath.Join(home, ".npm-global", "bin"), customBin},
					counts: map[string]int{
						"/usr/bin": 1,
						customBin:  1,
					},
				}
			},
		},
		{
			name: "respects max path length",
			setup: func(t *testing.T) (string, expectation) {
				entries := make([]string, 0, 200)
				for i := 0; i < 200; i++ {
					entries = append(entries, filepath.Join("/tusk", strings.Repeat("x", 80), fmt.Sprintf("entry-%03d", i)))
				}

				return strings.Join(entries, sep), expectation{
					nonEmpty:         true,
					first:            entries[0],
					maxLen:           maxServicePathLength,
					maxPartsLessThan: len(entries),
				}
			},
		},
		{
			name: "default path still valid when HOME and PATH are empty",
			setup: func(t *testing.T) (string, expectation) {
				t.Setenv("HOME", "")
				t.Setenv("PATH", "")

				return "", expectation{
					nonEmpty:     true,
					contains:     []string{"/usr/bin", "/bin"},
					noEmptyParts: true,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customPath, want := tt.setup(t)
			got := BuildServicePATH(customPath)
			parts := filepath.SplitList(got)

			if want.nonEmpty && len(parts) == 0 {
				t.Fatal("BuildServicePATH returned empty PATH")
			}

			if want.exactParts != nil && !reflect.DeepEqual(parts, want.exactParts) {
				t.Fatalf("BuildServicePATH parts = %v, want %v", parts, want.exactParts)
			}

			if want.first != "" && len(parts) > 0 && parts[0] != want.first {
				t.Fatalf("first PATH entry = %q, want %q", parts[0], want.first)
			}

			for _, p := range want.contains {
				if !contains(parts, p) {
					t.Fatalf("PATH does not contain %q", p)
				}
			}

			for p, expectedCount := range want.counts {
				if gotCount := count(parts, p); gotCount != expectedCount {
					t.Fatalf("expected %q exactly %d times, got %d", p, expectedCount, gotCount)
				}
			}

			if want.noEmptyParts {
				for _, p := range parts {
					if strings.TrimSpace(p) == "" {
						t.Fatal("PATH contains empty entry")
					}
				}
			}

			if want.maxLen > 0 && len(got) > want.maxLen {
				t.Fatalf("PATH length = %d, exceeds max %d", len(got), want.maxLen)
			}

			if want.maxPartsLessThan > 0 && len(parts) >= want.maxPartsLessThan {
				t.Fatalf("expected truncation by max length, kept %d entries of %d", len(parts), want.maxPartsLessThan)
			}
		})
	}
}

func contains(parts []string, value string) bool {
	for _, p := range parts {
		if p == value {
			return true
		}
	}
	return false
}

func count(parts []string, value string) int {
	n := 0
	for _, p := range parts {
		if p == value {
			n++
		}
	}
	return n
}
