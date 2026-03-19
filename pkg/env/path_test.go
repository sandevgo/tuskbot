package env

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestBuildServicePATH_CustomPathDeduplicatesAndTrims(t *testing.T) {
	sep := string(filepath.ListSeparator)
	base := t.TempDir()
	p1 := filepath.Join(base, "one")
	p2 := filepath.Join(base, "two")

	customPath := strings.Join([]string{" " + p1 + " ", p2, p1, "", "  "}, sep)
	got := BuildServicePATH(customPath)

	parts := filepath.SplitList(got)
	want := []string{p1, p2}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("BuildServicePATH(custom) = %v, want %v", parts, want)
	}
}

func TestBuildServicePATH_DefaultPathIncludesHomeAndDeduplicatesPATH(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sep := string(filepath.ListSeparator)
	customBin := filepath.Join(t.TempDir(), "custom-bin")
	t.Setenv("PATH", strings.Join([]string{"/usr/bin", customBin, "/bin", customBin}, sep))

	got := BuildServicePATH("")
	parts := filepath.SplitList(got)
	if len(parts) == 0 {
		t.Fatal("BuildServicePATH(\"\") returned empty PATH")
	}

	homeLocalBin := filepath.Join(home, ".local", "bin")
	homeNPMBin := filepath.Join(home, ".npm-global", "bin")

	if parts[0] != homeLocalBin {
		t.Fatalf("first PATH entry = %q, want %q", parts[0], homeLocalBin)
	}
	if !contains(parts, homeNPMBin) {
		t.Fatalf("PATH does not contain %q", homeNPMBin)
	}
	if count(parts, "/usr/bin") != 1 {
		t.Fatalf("expected /usr/bin exactly once, got %d occurrences", count(parts, "/usr/bin"))
	}
	if count(parts, customBin) != 1 {
		t.Fatalf("expected %q exactly once, got %d occurrences", customBin, count(parts, customBin))
	}
}

func TestBuildServicePATH_RespectsMaxLength(t *testing.T) {
	sep := string(filepath.ListSeparator)
	entries := make([]string, 0, 200)
	for i := 0; i < 200; i++ {
		entries = append(entries, filepath.Join("/tusk", strings.Repeat("x", 80), fmt.Sprintf("entry-%03d", i)))
	}

	got := BuildServicePATH(strings.Join(entries, sep))
	if len(got) > maxServicePathLength {
		t.Fatalf("PATH length = %d, exceeds max %d", len(got), maxServicePathLength)
	}

	parts := filepath.SplitList(got)
	if len(parts) == 0 {
		t.Fatal("PATH unexpectedly empty after length limiting")
	}
	if parts[0] != entries[0] {
		t.Fatalf("first retained entry = %q, want %q", parts[0], entries[0])
	}
	if len(parts) >= len(entries) {
		t.Fatalf("expected truncation by max length, kept %d of %d entries", len(parts), len(entries))
	}
}

func TestBuildServicePATH_DefaultPathWithoutHomeDoesNotFail(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("PATH", "")

	got := BuildServicePATH("")
	parts := filepath.SplitList(got)
	if len(parts) == 0 {
		t.Fatal("BuildServicePATH(\"\") returned empty PATH")
	}

	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			t.Fatal("PATH contains empty entry")
		}
	}
	if !contains(parts, "/usr/bin") {
		t.Fatal("PATH does not contain /usr/bin default entry")
	}
	if !contains(parts, "/bin") {
		t.Fatal("PATH does not contain /bin default entry")
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
