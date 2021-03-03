package cocore

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/legenove/easyconfig/ifacer"
)

var App *Application
var ReloadTime = 60 * time.Second
var appInitFunc map[string]func()
var resetChan chan struct{}
var listenFuncTime int32

func init() {
	appInitFunc = make(map[string]func())
	resetChan = make(chan struct{})
}

type Application struct {
	sync.Mutex
	DEBUG      bool
	LogDir     string
	LogType    string
	AppENV     string
	AppConf    ifacer.Configer
	AppOptions AppParam
}

func RegisterInitFunc(name string, f func()) {
	appInitFunc[name] = f
}

func (app *Application) initAppConf() error {
	app.Lock()
	defer app.Unlock()
	appConf, err := Conf.Instance(app.AppOptions.Name, app.AppOptions.ParseType, nil)
	if err == nil {
		app.AppConf = appConf
	}
	return err
}
func (app *Application) listenAppConfChange() {
	if app.AppConf != nil {
		if atomic.CompareAndSwapInt32(&listenFuncTime, 0, 1) {
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
	}
}

func (app *Application) loadAppConf() {
	app.Lock()
	defer app.Unlock()
	if app.AppConf == nil {
		appConf, err := Conf.Instance(app.AppOptions.Name, app.AppOptions.ParseType, nil)
		if err == nil {
			app.AppConf = appConf
			app.listenAppConfChange()
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

func InitApp(debug bool, appEnv string, params AppParam) {
	if App != nil {
		return
	}
	var logType string
	switch params.LogType {
	case LOG_TYPE_CONSOLE:
		logType = LOG_TYPE_CONSOLE
	case LOG_TYPE_FILE:
		logType = LOG_TYPE_FILE
	default:
		logType = LOG_TYPE_FILE
	}
	App = &Application{
		DEBUG:      debug,
		AppENV:     appEnv,
		LogType:    logType,
		AppOptions: params,
	}
	InitConf(App.AppOptions)
	err := App.initAppConf()
	if err == nil {
		select {
		case <-time.After(3 * time.Second):
			fmt.Println("Conf Load Error: init app conf error in 3 secend")
		case <-App.AppConf.OnChangeChan():
		}
		App.listenAppConfChange()
	} else {
		fmt.Println("Conf Load Error: init error", err)
	}
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
	initialLog()
	initial()
	RegisterInitFunc("cocoreInitLog", initialLog)
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
	atomic.StoreInt32(&listenFuncTime, 0)
}

// 初始化信息
func initial() {
	// other initial
	for _, f := range appInitFunc {
		f()
	}
}
