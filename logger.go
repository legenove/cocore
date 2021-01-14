package cocore

import (
	"github.com/legenove/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"sync"
	"time"
)

type lwriter struct {
	logger *zap.Logger
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
	LogEnableLevel  = zap.InfoLevel
	LogFormat       string
	TimeDelter      int64 = 86410
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
				if t < nowat-TimeDelter {
					LogPool.mutex.Lock()
					delete(LogPool.registerTime, file)
					l, ok := LogPool.writers[file]
					if ok {
						fdgc := &fdGC{free: l.free, expir: time.Now().Unix()}
						LogPool.free[l.file] = fdgc
						delete(LogPool.writers, file)
					}
					LogPool.mutex.Unlock()
				} else {
					// 防止文件被删除
					if !utils.FileExists(getLogFile(file)) {
						LogPool.mutex.Lock()
						l, ok := LogPool.writers[file]
						if ok {
							fdgc := &fdGC{free: l.free, expir: time.Now().Unix()}
							LogPool.free[l.file] = fdgc
							delete(LogPool.writers, file)
						}
						LogPool.mutex.Unlock()
					}
				}
			}
			time.Sleep(2 * time.Minute)
		}
	}()
}

func initialLog() {
	app := App
	LogPool.LogDir = app.LogDir
	if len(LogPool.LogDir) == 0 {
		LogPool.LogDir = "/tmp/logs/"
	}
	if !strings.HasSuffix(LogPool.LogDir, "/") {
		LogPool.LogDir = LogPool.LogDir + "/"
	}
	LogPool.Debug = app.DEBUG
	switch app.GetStringConfig("LOG_ENABLE_LEVEL", LOG_LEVEL_INFO) {
	case LOG_LEVEL_WARN:
		LogEnableLevel = zap.WarnLevel
	case LOG_LEVEL_ERROR:
		LogEnableLevel = zap.ErrorLevel
	case LOG_LEVEL_DEBUG:
		LogEnableLevel = zap.DebugLevel
	default:
		LogEnableLevel = zap.InfoLevel
	}
	if app.GetStringConfig("LOG_TIME_GROUP", LOG_FORMAT_DAILY) == LOG_FORMAT_HOUR {
		TimeDelter = 3610
		LogFormat = "20060102T15"
	} else {
		TimeDelter = 86410
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

func newLogger(file string) (*zap.Logger, func(), error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendInt64(ts.In(TimeLocation).Unix())
	}

	writer, closeFD, err := zap.Open(file)
	if err != nil {
		return nil, nil, err
	}

	var level zap.AtomicLevel
	if LogPool.Debug {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(LogEnableLevel)
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, level)
	logger := zap.New(core)
	host, e := os.Hostname()
	if e != nil {
		host = ""
	}
	f := []zapcore.Field{
		zap.String("serverHostName", host),
	}
	logger = logger.With(f...)
	return logger, closeFD, nil
}

func (pl *Logger) Instance(k string) (*zap.Logger, error) {
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
