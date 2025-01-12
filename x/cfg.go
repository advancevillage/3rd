package x

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type Cfg struct {
	App   *AppCfg  `yaml:"app"`
	Log   *LogCfg  `yaml:"log"`
	Creds *CredCfg `yaml:"creds"`
}

type AppCfg struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LogCfg struct {
	Level string `yaml:"level"`
}

type CredCfg struct {
	MTls *MTlsCfg `yaml:"mtls"`
}

type MTlsCfg struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

func NewCfg(filepath string) (*Cfg, error) {
	fd, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	c := &Cfg{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
