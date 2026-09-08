package adsefid

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixturesIntegrity guards the golden fixtures, which are byte-identical
// copies of the same tree in the sibling SDK repositories. A drifting fixture
// silently weakens every other test in this file, so verify the manifest.
func TestFixturesIntegrity(t *testing.T) {
	manifest, err := os.ReadFile(filepath.Join("testdata", "CHECKSUMS.txt"))
	if err != nil {
		t.Fatalf("read CHECKSUMS.txt: %v", err)
	}

	listed := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(manifest)), "\n") {
		want, name, ok := strings.Cut(line, "  ")
		if !ok {
			t.Fatalf("malformed manifest line %q", line)
		}
		listed[name] = true

		content, err := os.ReadFile(filepath.Join("testdata", filepath.FromSlash(name)))
		if err != nil {
			t.Errorf("fixture %s listed in the manifest is missing: %v", name, err)
			continue
		}
		if got := hex.EncodeToString(sha256Sum(content)); got != want {
			t.Errorf("fixture %s changed: manifest has %s, file hashes to %s", name, want, got)
		}
	}

	root := filepath.Join("testdata")
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || info.Name() == "CHECKSUMS.txt" {
			return err
		}
		name := filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator)))
		if !listed[name] {
			t.Errorf("fixture %s is not listed in CHECKSUMS.txt", name)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata: %v", err)
	}
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
