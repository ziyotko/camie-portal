package generator

import (
	"camie-portal/backend/internal/config"
	"fmt"
	"html"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func writeFile(root, name string, data []byte) error {
	target := filepath.Join(root, filepath.FromSlash(name))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("output escapes staging")
	}
	if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0644)
}
func copyAssets(source, target string) error {
	for _, name := range []string{"styles.css", "script.js", "favicon.ico", "assets", "slice"} {
		start := filepath.Join(source, name)
		err := filepath.WalkDir(start, func(file string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("asset symlinks are not supported: %s", file)
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf("not a regular asset: %s", file)
			}
			relative, err := filepath.Rel(source, file)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			return writeFile(target, relative, data)
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func publishDirectory(source, stage, target string) error {
	if err := config.ValidateOutput(source, target); err != nil {
		return err
	}
	if err := config.ValidateOutput(source, stage); err != nil {
		return err
	}
	previous := target + ".previous"
	if err := config.ValidateOutput(source, previous); err != nil {
		return err
	}
	if err := os.Chmod(stage, 0755); err != nil {
		return err
	}
	if err := os.RemoveAll(previous); err != nil {
		return err
	}
	hadTarget := false
	if _, err := os.Stat(target); err == nil {
		if err = os.Rename(target, previous); err != nil {
			return err
		}
		hadTarget = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(stage, target); err != nil {
		if hadTarget {
			if rollback := os.Rename(previous, target); rollback != nil {
				return fmt.Errorf("publish failed: %v; rollback failed: %w", err, rollback)
			}
		}
		return err
	}
	return nil
}

var references = regexp.MustCompile(`(?i)(?:href|src|poster)="([^"]+)"`)

func validateSite(root string) error {
	for _, name := range []string{"index.html", "news.html", "videos.html", "detail.html", "video-detail.html", "search.html", "styles.css", "script.js", "favicon.ico", "generated-content.js"} {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			return err
		}
		if info.Size() == 0 {
			return fmt.Errorf("empty output %s", name)
		}
	}
	return filepath.WalkDir(root, func(file string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(file) != ".html" {
			return err
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), "#ZgotmplZ") {
			return fmt.Errorf("unsafe template URL in %s", file)
		}
		for _, match := range references.FindAllSubmatch(data, -1) {
			raw := html.UnescapeString(string(match[1]))
			if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "/") {
				continue
			}
			u, err := url.Parse(raw)
			if err != nil {
				return err
			}
			if u.IsAbs() {
				continue
			}
			destination := filepath.Clean(filepath.Join(filepath.Dir(file), filepath.FromSlash(u.Path)))
			rel, err := filepath.Rel(root, destination)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return fmt.Errorf("link escapes site: %s", raw)
			}
			slash := filepath.ToSlash(rel)
			managed := strings.HasSuffix(slash, ".html") || strings.HasPrefix(slash, "assets/") || strings.HasPrefix(slash, "slice/") || slash == "styles.css" || slash == "script.js" || slash == "favicon.ico" || slash == "generated-content.js"
			if managed {
				if _, err = os.Stat(destination); err != nil {
					return fmt.Errorf("broken link %s in %s", raw, file)
				}
			}
		}
		return nil
	})
}
