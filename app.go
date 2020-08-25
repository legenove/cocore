package cocore

import (
	"github.com/legenove/viper_conf"
	"os"
	"path/filepath"
	"strings"
)

type Application struct {
	DEBUG       bool
	AppENV      string
	ConfigDir   string
	AppConf     *viper_conf.ViperConf
	AppConfName string
}

func (app *Application) GetStringConfig(key, default_value string) string {
	if app.AppConf != nil {
		value, _ := app.AppConf.GetString(key)
		if value != "" {
			return value
		}
	} else {
		appConf, err := Conf.Instance(app.AppConfName, nil, ReloadConfig, nil)
		if err == nil {
			App.AppConf = appConf
		}
	}
	return default_value
}

var App *Application

func InitApp(debug bool, appEnv string, configDir, appConfName string) {
	if appConfName == "" {
		appConfName = "app.toml"
	}
	App = &Application{
		DEBUG:     debug,
		AppENV:    appEnv,
		ConfigDir: configDir,
		AppConfName: appConfName,
	}
	//注册配置信息
	if strings.HasPrefix(App.ConfigDir, "$GOPATH") {
		App.ConfigDir = filepath.Join(os.Getenv("GOPATH"), App.ConfigDir[7:])
	}
	InitConf(App.AppENV, App.ConfigDir)
	appConf, err := Conf.Instance(appConfName, nil, ReloadConfig, nil)
	if err == nil {
		App.AppConf = appConf
	}
	initial()
}

// 初始化信息
func initial() {
	// other initial
}

func ReloadConfig(conf *viper_conf.ViperConf) {
	initial()
}
