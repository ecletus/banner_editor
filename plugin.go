package banner_editor

import (
	"github.com/moisespsena/go-assetfs/api"
	"github.com/aghape/assets"
	"github.com/aghape/db"
	"github.com/aghape/plug"
)

type Plugin struct {
	db.DisDBNames
	AssetFSKey string
}

func (p *Plugin) OnRegister() {
	p.DBOnMigrateGorm(func(e *db.GormDBEvent) error {
		return e.DB.AutoMigrate(&QorBannerEditorSetting{}).Error
	})
}

func (p *Plugin) Init(options *plug.Options) error {
	SetAssetFS(assets.TemplateFS(options.GetInterface(p.AssetFSKey).(api.Interface)).NameSpace("banner_editor"))
	return nil
}
