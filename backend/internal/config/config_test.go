package config

import (
	"path/filepath"
	"testing"
)

func TestOutputConfinedToDist(t *testing.T) {
	source := t.TempDir()
	for _, target := range []string{source, filepath.Dir(source), filepath.Join(source, "dist"), filepath.Join(source, "backend"), filepath.Join(source, "dist", "..", "assets")} {
		if ValidateOutput(source, target) == nil {
			t.Fatalf("unsafe output accepted: %s", target)
		}
	}
	if err := ValidateOutput(source, filepath.Join(source, "dist", "camie-preview")); err != nil {
		t.Fatal(err)
	}
}
