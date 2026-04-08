package deploy

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateData is passed to every text/template.
type TemplateData struct {
	Port        string
	MainPackage string
	HealthPath  string
	Module      string
	APIBaseURL  string
}

// DetectModule reads the `module` directive from go.mod in dir.
func DetectModule(dir string) (string, error) {
	path := filepath.Join(dir, "go.mod")
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("go.mod okunamadı (%s): %w", path, err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if mod == "" {
				continue
			}
			// strip // comments
			if i := strings.Index(mod, "//"); i >= 0 {
				mod = strings.TrimSpace(mod[:i])
			}
			return mod, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("go.mod içinde module satırı yok: %s", path)
}

func mergeTemplateData(s *Schema, opts InitOptions) (TemplateData, error) {
	d := TemplateData{
		Port:        s.Defaults.Port,
		MainPackage: s.Defaults.MainPackage,
		HealthPath:  s.Defaults.HealthPath,
		Module:      s.Defaults.Module,
		APIBaseURL:  s.Defaults.APIBaseURL,
	}
	if opts.Port != "" {
		d.Port = opts.Port
	}
	if opts.MainPackage != "" {
		d.MainPackage = opts.MainPackage
	}
	if opts.HealthPath != "" {
		d.HealthPath = opts.HealthPath
	}
	if opts.APIBaseURL != "" {
		d.APIBaseURL = opts.APIBaseURL
	}
	if opts.Module != "" {
		d.Module = opts.Module
	} else if mod, err := DetectModule(opts.OutDir); err == nil {
		d.Module = mod
	}
	if d.Module == "" {
		return TemplateData{}, fmt.Errorf("Go modülü yok; repoda go.mod olduğundan emin olun veya --module kullanın")
	}
	if d.APIBaseURL == "" {
		d.APIBaseURL = "https://YOUR_RAILWAY_OR_FLY_API_URL"
	}
	return d, nil
}
