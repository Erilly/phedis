package config

import (
	"github.com/naoina/toml"
	"io/ioutil"
	"time"
)

type Options struct {
	Listen             string        `toml:"listen"`
	LogFile            string        `toml:"logfile"`
	PhredisProxyKey    string        `toml:"phredis_proxy_key"`
	Timeout            time.Duration `toml:"timeout"`
	MinProxyPoolLength int `toml:"min_proxy_pool_length"`
	MaxProxyPoolLength int `toml:"max_proxy_pool_length"`
}

var Configs *Options

func initConfig() {
	by, err := ioutil.ReadFile(conf)
	if err != nil {
		panic(err)
	}
	Configs = &Options{}
	err = toml.Unmarshal(by, Configs)

	if err != nil {
		panic(err)
	}
}

func Start() {
	initFlag()
	initConfig()
}
