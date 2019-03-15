package banner_editor

import (
	"github.com/ecletus/assets"
	"github.com/ecletus/db"
	"github.com/ecletus/plug"
	"github.com/moisespsena/go-assetfs/api"
)

type Plugin struct {
	db.DBNames
	plug.EventDispatcher
	AssetFSKey string
}

func (p *Plugin) OnRegister() {
	db.Events(p).DBOnMigrate(func(e *db.DBEvent) error {
		return e.AutoMigrate(&QorBannerEditorSetting{}).Error
	})
}

func (p *Plugin) Init(options *plug.Options) error {
	SetAssetFS(assets.TemplateFS(options.GetInterface(p.AssetFSKey).(assetfsapi.Interface)).NameSpace("banner_editor"))
	return nil
}
