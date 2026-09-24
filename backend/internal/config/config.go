package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Database struct {
		DSNEnv string `yaml:"dsn_env"`
	} `yaml:"database"`
	Server struct {
		Addr           string `yaml:"addr"`
		TokenEnv       string `yaml:"token_env"`
		RequestTimeout string `yaml:"request_timeout"`
	} `yaml:"server"`
	Site struct {
		SourceRoot    string `yaml:"source_root"`
		DistRoot      string `yaml:"dist_root"`
		PreviewRoot   string `yaml:"preview_root"`
		TemplateRoot  string `yaml:"template_root"`
		PageName      string `yaml:"page_name"`
		PageSize      int    `yaml:"page_size"`
		Timezone      string `yaml:"timezone"`
		MediaBaseURL  string `yaml:"media_base_url"`
		FallbackCover string `yaml:"fallback_cover"`
		HeroColumn    string `yaml:"hero_column"`
	} `yaml:"site"`
}

func Load(name string) (Config, error) {
	var c Config
	data, err := os.ReadFile(name)
	if err != nil {
		return c, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err = decoder.Decode(&c); err != nil {
		return c, err
	}
	if c.Database.DSNEnv == "" {
		c.Database.DSNEnv = "CAMIE_DB_DSN"
	}
	if c.Server.Addr == "" {
		c.Server.Addr = "127.0.0.1:9144"
	}
	if c.Server.TokenEnv == "" {
		c.Server.TokenEnv = "CAMIE_STATIC_TOKEN"
	}
	if c.Server.RequestTimeout == "" {
		c.Server.RequestTimeout = "10m"
	}
	if d, e := time.ParseDuration(c.Server.RequestTimeout); e != nil || d <= 0 {
		return c, fmt.Errorf("invalid request_timeout")
	}
	if c.Site.PageSize < 1 || c.Site.PageSize > 100 {
		return c, fmt.Errorf("page_size must be between 1 and 100")
	}
	if strings.TrimSpace(c.Site.PageName) == "" {
		return c, fmt.Errorf("site.page_name is required")
	}
	if _, err = time.LoadLocation(c.Site.Timezone); err != nil {
		return c, err
	}
	base, _ := filepath.Abs(filepath.Dir(name))
	for _, p := range []*string{&c.Site.SourceRoot, &c.Site.DistRoot, &c.Site.PreviewRoot, &c.Site.TemplateRoot} {
		if *p == "" {
			return c, fmt.Errorf("site paths cannot be empty")
		}
		if !filepath.IsAbs(*p) {
			*p = filepath.Join(base, *p)
		}
		*p = filepath.Clean(*p)
	}
	for _, target := range []string{c.Site.DistRoot, c.Site.PreviewRoot} {
		if err = ValidateOutput(c.Site.SourceRoot, target); err != nil {
			return c, err
		}
	}
	if strings.EqualFold(c.Site.DistRoot, c.Site.PreviewRoot) {
		return c, fmt.Errorf("preview_root must differ from dist_root")
	}
	return c, nil
}

// All generated and recursively managed files stay within this project's dist directory.
func ValidateOutput(source, target string) error {
	base := filepath.Join(source, "dist")
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("output must be a child of %s", base)
	}
	for p := target; ; p = filepath.Dir(p) {
		info, e := os.Lstat(p)
		if e == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output path contains a symlink: %s", p)
		}
		if e != nil && !os.IsNotExist(e) {
			return e
		}
		if strings.EqualFold(p, source) || filepath.Dir(p) == p {
			break
		}
	}
	return nil
}
