package cocore

import "github.com/legenove/viper_conf"

var Conf *viper_conf.FileConf

func InitConf(env, confPath string) {
	Conf = viper_conf.NewConf(env, confPath)
}
