package proxy

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

type PorxyConfig struct {
	Listen string `yaml:"listen"`
	Mport string `yaml:"mport"`
	ConfigToken string `yaml:"configtoken"`
	ProxyList []string `yaml:"proxylist"`
	BackendCfg BackendCfg `yaml:"backendcfg"`
}

type BackendCfg struct {
	CheckTimeout int `yaml:"checktimeout"`
	CheckFail int `yaml:"checkfail"`
	CheckInter time.Duration `yaml:"checkinter"`
	Backends []Backend `yaml:"backends"`
}

type Backend struct {
	Addr string `yaml:"addr"`
	Weight int `yaml:"weight"`
	MaxConn int `yaml:"maxconn"`
}


func (PorxyConfig *PorxyConfig) GetConf(configFile string) *PorxyConfig {
    yamlFile, err := ioutil.ReadFile(configFile)
    if err != nil {
        log.Println(err.Error())
    }

    err = yaml.UnmarshalStrict(yamlFile, PorxyConfig)

    if err != nil {
        log.Println(err.Error())
    }

    return PorxyConfig
}

