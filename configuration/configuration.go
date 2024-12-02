package configuration

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/joho/godotenv"
)

type Configuration struct {
	Configurations *map[string]any
}

func NewConfiguration(envPath string) *Configuration {
	app := &Configuration{}
	configurations := make(map[string]any)
	app.Configurations = &configurations
	if err := godotenv.Load(envPath); err != nil {
		color.Redln("Invalid Configuration error: " + err.Error())
		os.Exit(0)
	}
	return app
}

// Env Get Configuration from env.
func (config *Configuration) Env(envName string, defaultValue interface{}) interface{} {
	envValue := os.Getenv(envName)
	if envValue == "" {
		return defaultValue
	}
	if defaultValue == nil {
		return envValue
	}

	defaultType := reflect.TypeOf(defaultValue)
	if defaultType.Kind() == reflect.Ptr {
		defaultType = defaultType.Elem()
	}

	switch defaultType.Kind() {
	case reflect.Bool:
		if envValue == "true" {
			return true
		} else if envValue == "false" {
			return false
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intValue, err := strconv.ParseInt(envValue, 10, 64); err == nil {
			if reflect.TypeOf(defaultValue).Kind() == reflect.Ptr {
				val := reflect.New(defaultType).Elem()
				val.SetInt(intValue)
				return val.Addr().Interface()
			}
			return reflect.ValueOf(intValue).Convert(defaultType).Interface()
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintValue, err := strconv.ParseUint(envValue, 10, 64); err == nil {
			if reflect.TypeOf(defaultValue).Kind() == reflect.Ptr {
				val := reflect.New(defaultType).Elem()
				val.SetUint(uintValue)
				return val.Addr().Interface()
			}
			return reflect.ValueOf(uintValue).Convert(defaultType).Interface()
		}
	case reflect.Float32, reflect.Float64:
		if floatValue, err := strconv.ParseFloat(envValue, 64); err == nil {
			if reflect.TypeOf(defaultValue).Kind() == reflect.Ptr {
				val := reflect.New(defaultType).Elem()
				val.SetFloat(floatValue)
				return val.Addr().Interface()
			}
			return reflect.ValueOf(floatValue).Convert(defaultType).Interface()
		}
	case reflect.String:
		return envValue
	}
	return defaultValue
}

func (config *Configuration) Add(name string, configuration any) {
	(*config.Configurations)[name] = configuration
}

// Get Configuration from Configurationapplication.
func (config *Configuration) Get(path string, defaultValue any) any {
	return getConfigValue(*config.Configurations, strings.Split(path, "."), defaultValue)
}

func getConfigValue(config map[string]any, keys []string, defaultValue any) any {
	if len(keys) == 0 {
		return defaultValue
	}
	currentKey := keys[0]
	remainingKeys := keys[1:]
	if len(remainingKeys) == 0 {
		return config[currentKey]
	} else if valueMap, ok := config[currentKey].(map[string]any); ok {
		return getConfigValue(valueMap, remainingKeys, defaultValue)
	}
	return defaultValue
}
