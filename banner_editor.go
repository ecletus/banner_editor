package banner_editor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"

	"github.com/jinzhu/gorm"
	"github.com/moisespsena/go-assetfs"
	"github.com/aghape/admin"
	"github.com/aghape/aghape"
	"github.com/aghape/aghape/resource"
	"github.com/aghape/serializable_meta"
)

var (
	registeredElements           []*Element
	registeredExternalStylePaths []string
	viewPaths                    []string
	assetFileSystem              assetfs.Interface
)

type BannerSize struct {
	Width  int
	Height int
}

// BannerEditorConfig configure display elements and setting model
type BannerEditorConfig struct {
	MediaLibrary    *admin.Resource
	BannerSizes     map[string]BannerSize
	Elements        []string
	SettingResource *admin.Resource
}

// QorBannerEditorSettingInterface interface to support customize setting model
type QorBannerEditorSettingInterface interface {
	GetID() uint
	serializable_meta.SerializableMetaInterface
}

// QorBannerEditorSetting default setting model
type QorBannerEditorSetting struct {
	gorm.Model
	serializable_meta.SerializableMeta
}

// Element represent a button/element in banner_editor toolbar
type Element struct {
	Icon     string
	Name     string
	Template string
	Resource *admin.Resource
	Context  func(context *admin.Context, setting interface{}) interface{}
}

// RegisterElement register a element
func RegisterElement(e *Element) {
	registeredElements = append(registeredElements, e)
}

// RegisterExternalStylePath register a asset path
func RegisterExternalStylePath(path string) {
	registeredExternalStylePaths = append(registeredExternalStylePaths, path)
}

// ConfigureQorMeta configure route and funcmap for banner_editor meta
func (config *BannerEditorConfig) ConfigureQorMeta(metaor resource.Metaor) {
	if meta, ok := metaor.(*admin.Meta); ok {
		meta.Type = "banner_editor"
		Admin := meta.GetBaseResource().(*admin.Resource).GetAdmin()

		if config.SettingResource == nil {
			config.SettingResource = Admin.NewResource(&QorBannerEditorSetting{})
		}
		if config.MediaLibrary == nil {
			panic("BannerEditor: MediaLibrary can't be blank.")
		} else {
			urlMeta := config.MediaLibrary.GetMeta("BannerEditorUrl")
			if getMediaLibraryResourceURLMethod(config.MediaLibrary.NewStruct(nil)).IsNil() {
				panic("BannerEditor: MediaLibrary's struct doesn't have any field implement URL method, please refer media_library.MediaLibrary{}.")
			}
			if urlMeta == nil {
				config.MediaLibrary.Meta(&admin.Meta{
					Name: "BannerEditorUrl",
					Type: "hidden",
					Valuer: func(v interface{}, c *qor.Context) interface{} {
						values := getMediaLibraryResourceURLMethod(v).Call([]reflect.Value{})
						if len(values) > 0 {
							return values[0]
						}
						return ""
					},
				})
				config.MediaLibrary.IndexAttrs(config.MediaLibrary.IndexAttrs(), "BannerEditorUrl")
			}
		}

		res := config.SettingResource
		res.Router.Get("/new", admin.NewHandler(New, &admin.RouteConfig{Resource: res}))
		res.Router.Post("/", admin.NewHandler(Create, &admin.RouteConfig{Resource: res}))
		res.ObjectRouter.Put("/", admin.NewHandler(Update, &admin.RouteConfig{Resource: res}))
		res.RegisterDefaultRouters("read", "update")

		Admin.RegisterFuncMap("banner_editor_configure", func(config *BannerEditorConfig) string {
			type element struct {
				Name      string
				CreateURL string
				Icon      string
			}
			var (
				selectedElements = registeredElements
				elements         = []element{}
				newElementURL    = res.GetAdmin().Router.Prefix() + fmt.Sprintf("/%v/new", res.ToParam())
			)
			if len(config.Elements) != 0 {
				selectedElements = []*Element{}
				for _, name := range config.Elements {
					if e := GetElement(name); e != nil {
						selectedElements = append(selectedElements, e)
					}
				}
			}
			for _, e := range selectedElements {
				elements = append(elements, element{Icon: e.Icon, Name: e.Name, CreateURL: fmt.Sprintf("%v?kind=%v", newElementURL, template.URLQueryEscaper(e.Name))})
			}
			results, err := json.Marshal(struct {
				Elements          []element
				ExternalStylePath []string
				EditURL           string
				BannerSizes       map[string]BannerSize
			}{
				Elements:          elements,
				ExternalStylePath: registeredExternalStylePaths,
				EditURL:           fmt.Sprintf("%v/%v/:id/edit", res.GetAdmin().Router.Prefix(), res.ToParam()),
				BannerSizes:       config.BannerSizes,
			})
			if err != nil {
				return err.Error()
			}
			return string(results)
		})
	}
}

// GetElement returnn element struct by name
func GetElement(name string) *Element {
	for _, e := range registeredElements {
		if e.Name == name {
			return e
		}
	}
	return nil
}

// GetID return setting ID
func (setting QorBannerEditorSetting) GetID() uint {
	return setting.ID
}

// GetSerializableArgumentResource return setting's resource
func (setting QorBannerEditorSetting) GetSerializableArgumentResource() *admin.Resource {
	element := GetElement(setting.Kind)
	if element != nil {
		return element.Resource
	}
	return nil
}

func getMediaLibraryResourceURLMethod(i interface{}) reflect.Value {
	value := reflect.Indirect(reflect.ValueOf(i))
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if urlMethod := field.MethodByName("URL"); urlMethod.IsValid() {
			return urlMethod
		}
	}
	return reflect.Value{}
}
