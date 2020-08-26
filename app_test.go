package cocore_test

import (
	"fmt"
	"github.com/legenove/cocore"
	"github.com/stretchr/testify/assert"
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
)

func init() {
	cocore.ReloadTime = 3 * time.Second
	cocore.InitApp(true, "", "$GOPATH/src/github.com/legenove/cocore/conf", "")
	filePath = path.Join(cocore.App.ConfigDir, "app.toml")
	backupPath = path.Join(cocore.App.ConfigDir, "app_back.toml")
	updatePath = path.Join(cocore.App.ConfigDir, "update_app.toml")
}

func TestInitApp(t *testing.T) {
	cocore.Reset()
	cocore.InitApp(true, "", "$GOPATH/src/github.com/legenove/cocore/conf", "")
	res := cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "123", res)
}

func TestAutoLoadAppConfig(t *testing.T) {
	cocore.Reset()
	fmt.Println(filePath)
	removeFile(filePath)
	cocore.InitApp(true, "", "$GOPATH/src/github.com/legenove/cocore/conf", "")
	res := cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "abc", res)
	copyFile(backupPath, filePath)
	time.Sleep(5 * time.Second)
	res = cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "123", res)
}

func TestInitFunc(t *testing.T) {
	cocore.Reset()
	var test = 1
	f := func() {
		test++
	}
	cocore.RegisterInitFunc("test", f)
	fmt.Println(filePath)
	cocore.InitApp(true, "", "$GOPATH/src/github.com/legenove/cocore/conf", "")
	res := cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "123", res)
	res = cocore.App.GetStringConfig("bcd", "bcd")
	assert.Equal(t, "bcd", res)
	copyFile(updatePath, filePath)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 3, test)
	res = cocore.App.GetStringConfig("abc", "abc")
	assert.Equal(t, "234", res)
	res = cocore.App.GetStringConfig("bcd", "bcd")
	assert.Equal(t, "345", res)
	copyFile(backupPath, filePath)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 4, test)
}

func removeFile(dst string) {
	os.Remove(dst)
}

func copyFile(src, dst string) {
	exec.Command("cp", src, dst).Run()
}
