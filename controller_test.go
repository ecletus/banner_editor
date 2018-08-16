package banner_editor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/aghape/admin"
	"github.com/aghape/banner_editor/test/config/bindatafs"
	"github.com/aghape/media"
	"github.com/aghape/media/media_library"
	"github.com/aghape/aghape"
	"github.com/aghape/aghape/test/utils"
	qor_utils "github.com/aghape/aghape/utils"
)

var (
	mux                  = http.NewServeMux()
	Server               = httptest.NewServer(mux)
	db                   = utils.TestDB()
	Admin                = admin.New(&qor.NewConfig(db))
	assetManagerResource *admin.Resource
)

type bannerEditorArgument struct {
	gorm.Model
	Value string `gorm:"size:4294967295";`
}

func init() {
	// Migrate database
	if err := db.DropTableIfExists(&QorBannerEditorSetting{}, &bannerEditorArgument{}, &media_library.QorMediaLibrary{}).Error; err != nil {
		panic(err)
	}
	media.RegisterCallbacks(db)
	db.AutoMigrate(&QorBannerEditorSetting{}, &bannerEditorArgument{}, &media_library.QorMediaLibrary{})

	// Banner Editor
	type subHeaderSetting struct {
		Text  string
		Color string
	}
	type buttonSetting struct {
		Text  string
		Link  string
		Color string
	}
	subHeaderRes := Admin.NewResource(&subHeaderSetting{})
	subHeaderRes.Meta(&admin.Meta{Name: "Text"})
	subHeaderRes.Meta(&admin.Meta{Name: "Color"})

	buttonRes := Admin.NewResource(&buttonSetting{})
	buttonRes.Meta(&admin.Meta{Name: "Text"})
	buttonRes.Meta(&admin.Meta{Name: "Link"})
	RegisterViewPath("github.com/aghape/banner_editor/test/views")

	RegisterElement(&Element{
		Name:     "Sub Header",
		Template: "sub_header",
		Resource: subHeaderRes,
		Context: func(c *admin.Context, r interface{}) interface{} {
			return r.(QorBannerEditorSettingInterface).GetSerializableArgument(r.(QorBannerEditorSettingInterface))
		},
	})
	RegisterElement(&Element{
		Name:     "Button",
		Template: "button",
		Resource: buttonRes,
		Context: func(c *admin.Context, r interface{}) interface{} {
			setting := r.(QorBannerEditorSettingInterface).GetSerializableArgument(r.(QorBannerEditorSettingInterface)).(*buttonSetting)
			setting.Color = "Red"
			return setting
		},
	})

	// Add asset resource
	assetManagerResource = Admin.AddResource(&media_library.QorMediaLibrary{})
	assetManagerResource.IndexAttrs("Title", "File")

	bannerEditorResource := Admin.AddResource(&bannerEditorArgument{}, &admin.Config{Name: "Banner"})
	bannerEditorResource.Meta(&admin.Meta{Name: "Value", Config: &BannerEditorConfig{
		MediaLibrary: assetManagerResource,
	}})

	Admin.MountTo("/admin", mux)
	mux.Handle("/system/", qor_utils.FileServer(http.Dir("public")))

	// Add dummy background image
	image := media_library.QorMediaLibrary{}
	file, err := os.Open("test/views/images/background.jpg")
	if err != nil {
		panic(err)
	}
	image.File.Scan(file)
	db.Create(&image)

	if os.Getenv("MODE") == "server" {
		db.Create(&bannerEditorArgument{
			Value: `<span id="qor-bannereditor__i9mt1" class="qor-bannereditor__draggable" data-edit-id="1" data-position-left="202" data-position-top="152" style="position: absolute; left: 16.8896%; top: 50.6667%;"><em style="color: #ff0000;">Hello World!</em>
</span>`,
		})
		fmt.Printf("Test Server URL: %v\n", Server.URL+"/admin")
		time.Sleep(time.Second * 3000)
	}
}

func TestGetConfig(t *testing.T) {
	otherBannerEditorResource := Admin.AddResource(&bannerEditorArgument{}, &admin.Config{Name: "other_banner_editor_argument"})
	otherBannerEditorResource.Meta(&admin.Meta{Name: "Value", Config: &BannerEditorConfig{
		Elements:     []string{"Sub Header"},
		MediaLibrary: assetManagerResource,
	}})

	anotherBannerEditorResource := Admin.AddResource(&bannerEditorArgument{}, &admin.Config{Name: "another_banner_editor_argument"})
	anotherBannerEditorResource.Meta(&admin.Meta{Name: "Value", Config: &BannerEditorConfig{
		Elements:     []string{"Button"},
		MediaLibrary: assetManagerResource,
	}})

	assertConfigIncludeElements(t, "banners", []string{"Sub Header", "Button"})
	assertConfigIncludeElements(t, "other_banner_editor_arguments", []string{"Sub Header"})
	assertConfigIncludeElements(t, "another_banner_editor_arguments", []string{"Button"})
}

func TestControllerCRUD(t *testing.T) {
	resp, _ := http.Get(Server.URL + "/admin/qor_banner_editor_settings/new?kind=Sub%20Header")
	assetPageHaveAttributes(t, resp, "Text", "Color")

	resp, _ = http.Get(Server.URL + "/admin/qor_banner_editor_settings/new?kind=Button")
	assetPageHaveAttributes(t, resp, "Text", "Link")

	// Test create setting via HTML request
	resp, _ = http.PostForm(Server.URL+"/admin/qor_banner_editor_settings?kind=Button", url.Values{
		"QorResource.Kind":                  {"Button"},
		"QorResource.SerializableMeta.Text": {"Search by Google"},
		"QorResource.SerializableMeta.Link": {"http://www.google.com"},
	})
	body, _ := ioutil.ReadAll(resp.Body)
	assetPageHaveText(t, string(body), "Search by Google")
	assetPageHaveText(t, string(body), "http://www.google.com")

	resp, _ = http.Get(Server.URL + "/admin/qor_banner_editor_settings/1/edit")
	body, _ = ioutil.ReadAll(resp.Body)
	assetPageHaveText(t, string(body), "Search by Google")
	assetPageHaveText(t, string(body), "http://www.google.com")

	// Test create setting via JSON request
	resp, _ = http.PostForm(Server.URL+"/admin/qor_banner_editor_settings.json?kind=Button", url.Values{
		"QorResource.Kind":                  {"Button"},
		"QorResource.SerializableMeta.Text": {"Search by Yahoo"},
		"QorResource.SerializableMeta.Link": {"http://www.yahoo.com"},
	})
	body, _ = ioutil.ReadAll(resp.Body)
	assetPageHaveText(t, string(body), `{"ID":2,"Template":"<a style='color:Red' href='http://www.yahoo.com'>Search by Yahoo</a>\n"`)

	// Test update setting via JSON request
	resp, _ = http.PostForm(Server.URL+"/admin/qor_banner_editor_settings/2.json?kind=Button", url.Values{
		"_method":                           {"PUT"},
		"QorResource.Kind":                  {"Button"},
		"QorResource.SerializableMeta.Text": {"Search by Bing"},
		"QorResource.SerializableMeta.Link": {"http://www.bing.com"},
	})
	body, _ = ioutil.ReadAll(resp.Body)
	assetPageHaveText(t, string(body), `{"ID":2,"Template":"<a style='color:Red' href='http://www.bing.com'>Search by Bing</a>\n"`)

	// Test Customize AssetFS
	SetAssetFS(bindatafs.AssetFS)
	resp, _ = http.PostForm(Server.URL+"/admin/qor_banner_editor_settings.json?kind=Button", url.Values{
		"QorResource.Kind":                  {"Button"},
		"QorResource.SerializableMeta.Text": {"Search by Baidu"},
		"QorResource.SerializableMeta.Link": {"http://www.baidu.com"},
	})
	body, _ = ioutil.ReadAll(resp.Body)
	assetPageHaveText(t, string(body), `{"ID":3,"Template":"<a style='color:Red' href='http://www.baidu.com'>Search by Baidu</a>\n"`)
}

func TestMediaLibraryURL(t *testing.T) {
	resp, _ := http.Get(Server.URL + "/admin/media_libraries")
	body, _ := ioutil.ReadAll(resp.Body)
	assetPageHaveText(t, string(body), "/system/media_libraries/1/file.jpg")
}

func assetPageHaveText(t *testing.T, body string, text string) {
	if !strings.Contains(body, text) {
		t.Error(color.RedString("PageHaveText: expect page have text %v, but got %v", text, body))
	}
}

func assetPageHaveAttributes(t *testing.T, resp *http.Response, attributes ...string) {
	body, _ := ioutil.ReadAll(resp.Body)
	for _, attr := range attributes {
		if !strings.Contains(string(body), fmt.Sprintf("QorResource.SerializableMeta.%v", attr)) {
			t.Error(color.RedString("PageHaveAttrributes: expect page have attributes %v, but got %v", attr, string(body)))
		}
	}
}

func assertConfigIncludeElements(t *testing.T, resourceName string, elements []string) {
	resp, _ := http.Get(fmt.Sprintf("%v/admin/%v/new", Server.URL, resourceName))
	body, _ := ioutil.ReadAll(resp.Body)
	results := []string{}
	for _, elm := range elements {
		urlParam := strings.Replace(elm, " ", "&#43;", -1)
		data := fmt.Sprintf("{\"Name\":\"%v\",\"CreateURL\":\"/admin/qor_banner_editor_settings/new?kind=%v\",\"Icon\":\"\"}", elm, urlParam)
		results = append(results, data)
	}
	resultStr := strings.Join(results, ",")
	expectedConfig := fmt.Sprintf("data-configure='{\"Elements\":[%v],\"ExternalStylePath\":null,\"EditURL\":\"/admin/qor_banner_editor_settings/:id/edit\",\"BannerSizes\":null}'", resultStr)
	expectedConfig = strings.Replace(expectedConfig, "\"", "&#34;", -1)
	expectedConfig = strings.Replace(expectedConfig, "'", "\"", -1)
	assetPageHaveText(t, string(body), expectedConfig)
}
