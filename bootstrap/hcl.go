package bootstrap

import (
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

var configParser = HCLConfigParser{}

func ParseConfig(filename string) (*RootConfig, error) {
	return configParser.Parse(filename)
}

type ConfigParser interface {
	Parse(filename string) (*RootConfig, error)
}

type HCLConfigParser struct {
}

func (p HCLConfigParser) Parse(filename string) (*RootConfig, error) {
	cfg := &RootConfig{}

	if err := hclsimple.DecodeFile(filename, evalContext, cfg); err != nil {
		return nil, errors.Wrapf(
			err,
			"cannot load configuration file %s",
			filename,
		)
	}
	return cfg, nil
}

type RootConfig struct {
	LogLevel string          `hcl:"log_level,optional"`
	Servers  []*ServerConfig `hcl:"server,block"`
}

type ServerConfig struct {
	Name           string                 `hcl:"name,label"`
	ListenAddress  string                 `hcl:"listen"`
	ReadTimeout    time.Duration          `hcl:"read_timeout,optional"`
	WriteTimeout   time.Duration          `hcl:"write_timeout,optional"`
	IdleTimeout    time.Duration          `hcl:"idle_timeout,optional"`
	ProxyServices  []*ProxyServiceConfig  `hcl:"proxy_service,block"`
	StaticServices []*StaticServiceConfig `hcl:"static_service,block"`
}

type ProxyServiceConfig struct {
	Method        string            `hcl:"method,optional"`
	PathTemplate  string            `hcl:"path_template,label"`
	Backend       BackendConfig     `hcl:"backend,block"`
	FlushInterval time.Duration     `hcl:"flush_interval,optional"`
	SetHeader     map[string]string `hcl:"set_header,optional"`
	Transformer   DynamicConfig     `hcl:"transformer,block"`
}

type StaticServiceConfig struct {
	Method       string            `hcl:"method,optional"`
	PathTemplate string            `hcl:"path_template,label"`
	SetHeader    map[string]string `hcl:"set_header,optional"`
	ResponseCode int               `hcl:"response_code"`
	Content      string            `hcl:"content,optional"`
}

type BackendConfig struct {
	TargetURL              string `hcl:"target_url"`
	SuccessHTTPStatusCodes []int  `hcl:"success_http_status_codes,optional"`
}

type DynamicConfig struct {
	Name       string               `hcl:"name,label"`
	Attributes map[string]cty.Value `hcl:",remain"`
}

func (c *DynamicConfig) ToConfig() (Config, error) {
	cfg := make(Config)
	for name, v := range c.Attributes {
		goV, err := extractGoValues(v, v.Type())
		if err != nil {
			return nil, errors.Wrap(err, "cannot extract value")
		}
		cfg[name] = goV
	}
	return cfg, nil
}
