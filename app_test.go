package cocore_test

import (
	"github.com/legenove/cocore"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

var (
	filePath   string
	backupPath string
	updatePath string
	loggerName string
)

func init() {
	cocore.ReloadTime = 3 * time.Second
	cocore.InitApp(true, "", cocore.AppParam{
		LogType:   cocore.LOG_TYPE_CONSOLE,
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.toml",
		ParseType: "toml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/cocore/conf",
		},
	})
	configDir := cocore.App.AppOptions.File.GetConfPath()
	filePath = path.Join(configDir, "app.toml")
	backupPath = path.Join(configDir, "app_back.toml")
	updatePath = path.Join(configDir, "update_app.toml")
	loggerName = "loggerTest"
}

func TestInitApp(t *testing.T) {
	cocore.Reset()
	cocore.InitApp(true, "", cocore.AppParam{
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.toml",
		ParseType: "toml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/cocore/conf",
		},
	})
	res := cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "abc", res)
	res = cocore.App.GetStringConfig("LOG_ENABLE_LEVEL", "debug")
	assert.Equal(t, "info", res)
}

func TestInitAppByYaml(t *testing.T) {
	cocore.Reset()
	cocore.InitApp(true, "", cocore.AppParam{
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.yaml",
		ParseType: "yaml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/cocore/conf",
		},
	})
	res := cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "abc", res)
	res = cocore.App.GetStringConfig("LOG_ENABLE_LEVEL", "debug")
	assert.Equal(t, "info", res)
}

func TestAutoLoadAppConfig(t *testing.T) {
	cocore.Reset()
	removeFile(filePath)
	cocore.InitApp(true, "", cocore.AppParam{
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.toml",
		ParseType: "toml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/cocore/conf",
		},
	})
	res := cocore.App.GetStringConfig("LOG_ENABLE_LEVEL", "debug")
	assert.Equal(t, "debug", res)
	copyFile(backupPath, filePath)
	time.Sleep(5 * time.Second)
	res = cocore.App.GetStringConfig("LOG_ENABLE_LEVEL", "debug")
	assert.Equal(t, "info", res)
}

func TestInitFunc(t *testing.T) {
	cocore.Reset()
	var test = 1
	f := func() {
		test++
	}
	cocore.RegisterInitFunc("test", f)
	var res string
	cocore.InitApp(true, "", cocore.AppParam{
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.toml",
		ParseType: "toml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/cocore/conf",
		},
	})
	res = cocore.App.GetStringConfig("LOG_ENABLE_LEVEL", "debug")
	assert.Equal(t, "info", res)
	res = cocore.App.GetStringConfig("update_val", "none")
	assert.Equal(t, "none", res)
	copyFile(updatePath, filePath)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 3, test)
	res = cocore.App.GetStringConfig("LOG_ENABLE_LEVEL", "debug")
	assert.Equal(t, "debug", res)
	res = cocore.App.GetStringConfig("update_val", "none")
	assert.Equal(t, "update", res)
	copyFile(backupPath, filePath)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, test)
}

func TestLogger_Instance(t *testing.T) {
	cocore.InitApp(true, "", cocore.AppParam{
		LogType:   cocore.LOG_TYPE_CONSOLE,
		Source:    cocore.SOURCE_CONFIG_FILE,
		Name:      "app.toml",
		ParseType: "toml",
		Nacos:     nil,
		File: &cocore.FileParam{
			Env:       "",
			ConfigDir: "$GOPATH/src/github.com/legenove/cocore/conf",
		},
	})
	log, err := cocore.LogPool.Instance(loggerName)
	if err != nil {
		t.Error(err.Error())
	}
	log.Info("msg", zap.String("test1", "123"))

	log, err = cocore.LogPool.Instance(loggerName+"1")
	if err != nil {
		t.Error(err.Error())
	}
	log.Info("msg", zap.String("test1", "123"))

	//os.RemoveAll("/tmp/cocore")
}

func removeFile(dst string) {
	os.Remove(dst)
}

func copyFile(src, dst string) {
	exec.Command("cp", src, dst).Run()
}
