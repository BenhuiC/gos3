package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var Cfg config

type config struct {
	Base     Base         `yaml:"base"`
	Http     HttpCfg      `yaml:"http"`
	Database DatabaseCfg  `yaml:"database"`
	Service  ServiceCfg   `yaml:"services"`
	Hasher   HasherConfig `yaml:"hasher"`
	Ceph     OSSConfig    `yaml:"ceph"`
}

type Base struct {
	Env               string `yaml:"env"`
	GacProjectID      int    `yaml:"gacProjectID"`
	GacProjectName    string `yaml:"gacProjectName"`
	GacWorkspaceName  string `yaml:"gacWorkspaceName"`
	GacWorkspaceID    string `yaml:"gacWorkspaceID"`
	GacWorkspaceRawID uint64 `yaml:"gacWorkspaceRawID"`
}

type HttpCfg struct {
	ListenAddr string `yaml:"listenAddr"`
}

type DatabaseCfg struct {
	DB               string `yaml:"db"`
	WorkerRedisURL   string `yaml:"workerRedisURL"`
	RedisInstanceURL string `yaml:"redisInstanceURL"`
}

type ServiceCfg struct {
	Gac        ServerConfig `yaml:"gac"`
	Annotation ServerConfig `yaml:"annotation"`
	Workspace  ServerConfig `yaml:"workspace"`
	IAM        ServerConfig `yaml:"iam"`
}

func InitConfig(v *viper.Viper) (err error) {
	err = v.Unmarshal(&Cfg, func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})
	return
}

type HasherConfig struct {
	Salt     string `yaml:"salt"`
	MinLen   int    `yaml:"minLen"`
	Alphabet string `yaml:"alphabet"`
}

type OSSConfig struct {
	Endpoint                  string `yaml:"endpoint"`
	AccessKeyID               string `yaml:"accessKeyID"`
	AccessKeySecret           string `yaml:"accessKeySecret"`
	Bucket                    string `yaml:"bucket"`
	S3ForcePathStyle          bool   `yaml:"s3ForcePathStyle"`
	InsecureSkipVerify        bool   `yaml:"insecureSkipVerify"`
	DisableEndpointHostPrefix bool   `yaml:"disableEndpointHostPrefix"`

	SecurityToken string `yaml:"securityToken"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
	AK   string `yaml:"ak"`
	SK   string `yaml:"sk"`
}
