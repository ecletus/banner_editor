package banner_editor

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/qor/assetfs"
)

// SetAssetFS set asset fs for render
func SetAssetFS(assetFS assetfs.Interface) {
	for _, viewPath := range viewPaths {
		assetFS.RegisterPath(viewPath)
	}

	assetFileSystem = assetFS
}

// RegisterViewPath register views directory
func RegisterViewPath(p string) {
	if filepath.IsAbs(p) {
		viewPaths = append(viewPaths, p)
		assetFileSystem.RegisterPath(p)
	} else {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			viewPaths = append(viewPaths, filepath.Join(gopath, "src", p))
			assetFileSystem.RegisterPath(filepath.Join(gopath, "src", p))
		}
	}
}

func render(file string, value interface{}) (template.HTML, error) {
	var (
		err     error
		content []byte
		tmpl    *template.Template
	)

	if content, err = assetFileSystem.Asset(file + ".tmpl"); err == nil {
		if tmpl, err = template.New(filepath.Base(file)).Parse(string(content)); err == nil {
			var result = bytes.NewBufferString("")
			if err = tmpl.Execute(result, value); err == nil {
				return template.HTML(result.String()), nil
			}
		}
	}

	return template.HTML(""), err
}
