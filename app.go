package cocore

import (
	"github.com/legenove/easyconfig/ifacer"
	"sync"
	"time"
)

var App *Application
var ReloadTime = 60 * time.Second
var appInitFunc map[string]func()
var resetChan chan struct{}

func init() {
	appInitFunc = make(map[string]func())
	resetChan = make(chan struct{})
}

type Application struct {
	sync.Mutex
	DEBUG         bool
	AppENV        string
	AppConf       ifacer.Configer
	AppConfParams ConfigParam
}

func RegisterInitFunc(name string, f func()) {
	appInitFunc[name] = f
}

func (app *Application) loadAppConf() {
	app.Lock()
	defer app.Unlock()
	if app.AppConf == nil {
		appConf, err := Conf.Instance(app.AppConfParams.Name, app.AppConfParams.ParseType, nil)
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

func InitApp(debug bool, appEnv string, configParams ConfigParam) {
	if App != nil {
		return
	}
	App = &Application{
		DEBUG:         debug,
		AppENV:        appEnv,
		AppConfParams: configParams,
	}
	InitConf(App.AppConfParams)
	App.loadAppConf()
	go func() {
		for {
			if App.AppConf != nil {
				break
			}
			App.loadAppConf()
			time.Sleep(ReloadTime)
		}
	}()
	// 初始化log
	initialLog(App)
	initial()
}

// for test
func Reset() {
	if len(resetChan) == 0 {
		resetChan <- struct{}{}
	}
	App = nil
	Conf = nil
	appInitFunc = make(map[string]func())
	resetChan = make(chan struct{})
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
			case <-resetChan:
				return
			case <-App.AppConf.OnChangeChan():
				initial()
			}
		}
	}()
}
