package cocore

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/legenove/viper_conf"
)

var App *Application

var appInitFunc map[string]func()

func init() {
	appInitFunc = make(map[string]func())
}

type Application struct {
	sync.Mutex
	DEBUG       bool
	AppENV      string
	ConfigDir   string
	AppConf     *viper_conf.ViperConf
	AppConfName string
}

func RegisterInitFunc(name string, f func()) {
	appInitFunc[name] = f
}

func (app *Application) loadAppConf() {
	app.Lock()
	defer app.Unlock()
	if app.AppConf == nil {
		appConf, err := Conf.Instance(app.AppConfName, nil)
		if err == nil {
			App.AppConf = appConf
			listenAppConfChange()
		}
	}
}

func (app *Application) GetStringConfig(key, default_value string) string {
	if app.AppConf == nil {
		app.loadAppConf()
	}
	if app.AppConf != nil {
		value, _ := app.AppConf.GetString(key)
		if value != "" {
			return value
		}
	}
	return default_value
}

func InitApp(debug bool, appEnv string, configDir, appConfName string) {
	if App != nil {
		return
	}
	if appConfName == "" {
		appConfName = "app.toml"
	}
	App = &Application{
		DEBUG:       debug,
		AppENV:      appEnv,
		ConfigDir:   configDir,
		AppConfName: appConfName,
	}
	//注册配置信息
	if strings.HasPrefix(App.ConfigDir, "$GOPATH") {
		App.ConfigDir = filepath.Join(os.Getenv("GOPATH"), App.ConfigDir[7:])
	}
	InitConf(App.AppENV, App.ConfigDir)
	App.loadAppConf()
	go func() {
		for {
			if App.AppConf != nil {
				break
			}
			App.loadAppConf()
			time.Sleep(60 * time.Second)
		}
	}()
	initial()
}

// 初始化信息
func initial() {
	// other initial
	for _, f := range appInitFunc {
		f()
	}
}

func listenAppConfChange() {
	go func() {
		for {
			select {
			case <- App.AppConf.OnChange:
				initial()
			}
		}
	}()
}

