package bootstrap

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"
)

type Config map[string]any

func (c Config) GetString(name string, defaultVal ...string) string {
	if v, ok := c[name]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

func extractConfigIfSet(name string, config Config) (Config, bool) {
	if subconfig, ok := config[name]; ok {
		switch result := subconfig.(type) {
		case Config:
			return result, true
		case map[string]any:
			return result, true
		case string:
			conf := make(Config)
			if err := json.Unmarshal([]byte(result), &conf); err == nil {
				return conf, true
			}
		}
	}
	return config, false
}

func decode(in, out any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: false,
		DecodeHook:  decodeHook,
		Result:      out,
		MatchName: func(mapKey, fieldName string) bool {
			if mapKey == fieldName {
				return true
			}
			if strings.EqualFold(mapKey, fieldName) {
				return true
			}
			return strings.EqualFold(mapKey, strcase.ToSnake(fieldName))
		},
	})
	if err != nil {
		return err
	}
	return decoder.Decode(in)
}

func decodeHook(from, to reflect.Type, data any) (any, error) {
	if to != reflect.TypeOf(time.Duration(0)) {
		return data, nil
	}
	switch from.Kind() {
	case reflect.String:
		return time.ParseDuration(data.(string))
	case reflect.Int:
		return time.Second * time.Duration(data.(int)), nil
	}
	return data, nil
}
