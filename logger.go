package cocore

import (
	"github.com/legenove/utils"
	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"os"
	"strings"
	"sync"
	"time"
)

type lwriter struct {
	logger *zerolog.Logger
	free   func()
	file   string
}

type fdGC struct {
	free  func()
	expir int64
}

type Logger struct {
	writers      map[string]*lwriter
	registerTime map[string]int64
	mutex        sync.RWMutex
	LogDir       string
	Debug        bool
	free         map[string]*fdGC
}

var (
	LogPool         *Logger
	LogHost         string
	TimeLocation, _ = utils.TimeLoadLocation()
	LogEnableLevel  = zerolog.InfoLevel
	LogFormat       string
)

const (
	LOG_LEVEL_DEBUG = "debug"
	LOG_LEVEL_INFO  = "info"
	LOG_LEVEL_WARN  = "warn"
	LOG_LEVEL_ERROR = "error"
)

const (
	LOG_FORMAT_DAILY = "daily"
	LOG_FORMAT_HOUR  = "hour"
)

func init() {
	LogPool = &Logger{
		writers:      make(map[string]*lwriter),
		free:         make(map[string]*fdGC),
		registerTime: make(map[string]int64),
	}
	go func() {
		for {
			for file, gc := range LogPool.free {
				LogPool.mutex.Lock()
				gc.free()
				delete(LogPool.free, file)
				LogPool.mutex.Unlock()
			}
			time.Sleep(10 * time.Second)
		}
	}()
	go func() {
		for {
			nowat := time.Now().Unix()
			for file, t := range LogPool.registerTime {
				if t < nowat-86410 {
					LogPool.mutex.Lock()
					delete(LogPool.registerTime, file)
					delete(LogPool.writers, file)
					LogPool.mutex.Unlock()
				} else {
					// 防止文件被删除
					if !utils.FileExists(getLogFile(file)) {
						LogPool.mutex.Lock()
						delete(LogPool.writers, file)
						LogPool.mutex.Unlock()
					}
				}
			}
			time.Sleep(2 * time.Minute)
		}
	}()
}

func initialLog(app *Application) {
	LogPool.LogDir = app.GetStringConfig("LOG_DIR", "/data/logs")
	if !strings.HasSuffix(LogPool.LogDir, "/") {
		LogPool.LogDir = LogPool.LogDir + "/"
	}
	LogPool.Debug = app.DEBUG
	switch app.GetStringConfig("LOG_ENABLE_LEVEL", LOG_LEVEL_INFO) {
	case LOG_LEVEL_WARN:
		LogEnableLevel = zerolog.WarnLevel
	case LOG_LEVEL_ERROR:
		LogEnableLevel = zerolog.ErrorLevel
	case LOG_LEVEL_DEBUG:
		LogEnableLevel = zerolog.DebugLevel
	default:
		LogEnableLevel = zerolog.InfoLevel
	}
	if app.GetStringConfig("LOG_TIME_GROUP", LOG_FORMAT_DAILY) == LOG_FORMAT_HOUR {
		LogFormat = "20060102T15"
	} else {
		LogFormat = "20060102"
	}
	host, e := os.Hostname()
	if e != nil {
		host = ""
	}
	if host != "" && !strings.HasSuffix(host, "/") {
		host += "/"
	}
	LogHost = "/" + host
}

func newLogger(filePath string) (*zerolog.Logger, func(), error) {
	writer, closeFD, err := zap.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	host, e := os.Hostname()
	if e != nil {
		host = ""
	}
	logger := zerolog.New(writer).With().Timestamp().Str("host", host).Logger()
	if LogPool.Debug {
		logger.Level(zerolog.DebugLevel)
	} else {
		logger.Level(LogEnableLevel)
	}
	return &logger, closeFD, nil
}

func (pl *Logger) Instance(k string) (*zerolog.Logger, error) {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()
	l, ok := pl.writers[k]
	if ok {
		// 不需要判断，因为输入的时候，会优先创建文件
		//f := getLogFile(k)
		//if utils.FileExists(f) {
		return l.logger, nil
		//}
		//fdgc := &fdGC{free: l.free, expir: time.Now().In(TimeLocation).Unix()}
		//pl.free[l.file] = fdgc
		//delete(pl.writers, k)
	}
	var err error
	file, err := initLogFile(k)
	if err != nil {
		return nil, err
	}

	logger, free, err := newLogger(file)
	if err != nil {
		return nil, err
	}

	pl.writers[k] = &lwriter{logger: logger, free: free, file: file}
	pl.registerTime[k] = time.Now().Unix()
	return logger, nil
}

func getLogFile(file string) string {
	return utils.ConcatenateStrings(LogPool.LogDir, time.Now().In(TimeLocation).Format(LogFormat), LogHost, file, ".log")
}

func getLogPath() string {
	return utils.ConcatenateStrings(LogPool.LogDir, time.Now().In(TimeLocation).Format(LogFormat), LogHost)
}

func initLogFile(file string) (string, error) {
	path := getLogPath()
	var err error
	exists := utils.PathExists(path)
	if exists == false {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return "", err
		}
	}

	logFile := getLogFile(file)
	if utils.FileExists(logFile) {
		return logFile, nil
	}

	fp, err := os.Create(logFile)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	if err := fp.Chmod(os.ModePerm); err != nil {
		return "", err
	}
	return logFile, nil
}
