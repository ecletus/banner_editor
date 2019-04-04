package banner_editor

import (
	"bytes"
	"path/filepath"
	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena/template/html/template"
)

// SetAssetFS set asset fs for render
func SetAssetFS(assetFS assetfs.Interface) {
	for _, viewPath := range viewPaths {
		assetFS.RegisterPath(viewPath)
	}

	assetFileSystem = assetFS
}

func render(file string, value interface{}) (template.HTML, error) {
	var (
		err   error
		asset assetfs.AssetInterface
		tmpl  *template.Template
	)

	if asset, err = assetFileSystem.Asset(file + ".tmpl"); err == nil {
		if tmpl, err = template.New(filepath.Base(file)).SetPath(asset.GetName()).Parse(asset.GetString()); err == nil {
			var result = bytes.NewBufferString("")
			if err = tmpl.Execute(result, value); err == nil {
				return template.HTML(result.String()), nil
			}
		}
	}

	return template.HTML(""), err
}
