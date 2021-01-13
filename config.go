package cocore

import (
	"github.com/legenove/easy-nacos-go/nacos_conf"
	"github.com/legenove/easyconfig/ifacer"
	"github.com/legenove/viper_conf"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"os"
	"path/filepath"
	"strings"
)

var Conf ifacer.ConfigManager

const (
	_                   = iota
	SOURCE_CONFIG_NACOS // nacos配置
	SOURCE_CONFIG_FILE  // 文件配置
)

func InitConf(param ConfigParam) {
	if Conf != nil {
		return
	}
	switch param.Source {
	case SOURCE_CONFIG_FILE:
		Conf = viper_conf.NewConf(param.File.Env, param.File.GetConfPath())
	case SOURCE_CONFIG_NACOS:
		Conf = nacos_conf.NewConfManage(param.Nacos.NameSpace, param.Nacos.Group,
			param.Nacos.DataIdPrefix, param.Nacos.ConfigClient)
	}
}

type ConfigParam struct {
	Source    int
	Name      string
	ParseType string
	options   []ifacer.OptionFunc
	Nacos     *NacosParam
	File      *FileParam
}

type NacosParam struct {
	NameSpace    string
	Group        string
	DataIdPrefix string
	ConfigClient config_client.IConfigClient
}

type FileParam struct {
	Env       string
	ConfigDir string
}

func (fp *FileParam) GetConfPath() string {
	if strings.HasPrefix(fp.ConfigDir, "$GOPATH") {
		fp.ConfigDir = filepath.Join(os.Getenv("GOPATH"), fp.ConfigDir[7:])
	}
	return fp.ConfigDir
}
