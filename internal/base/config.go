package base

import (
	"strconv"
)

type ConfigName string

const (
	ConfigOpenAIAPIKey  ConfigName = "openai.api_key"
	ConfigOpenAIModel   ConfigName = "openai.model"
	ConfigOpenAIBaseURL ConfigName = "openai.base_url"
	ConfigMaxIterations ConfigName = "max_iter"
	ConfigMaxHistory    ConfigName = "max_history"
)

var ConfigKeys = []ConfigName{
	ConfigOpenAIAPIKey,
	ConfigOpenAIModel,
	ConfigOpenAIBaseURL,
	ConfigMaxIterations,
	ConfigMaxHistory,
}

var defaultConfigValues = map[ConfigName]string{
	ConfigOpenAIModel:   "gpt-4o-mini",
	ConfigOpenAIBaseURL: "https://api.openai.com/v1",
	ConfigMaxIterations: "6",
	ConfigMaxHistory:    "10",
}

var configValues = map[ConfigName]string{}

func GetConfig(name ConfigName) string {
	if value, ok := configValues[name]; ok {
		return value
	}

	if value, ok := defaultConfigValues[name]; ok {
		return value
	}

	return ""
}

func GetIntConfig(name ConfigName) (value int, ok bool) {
	str := GetConfig(name)
	if v, err := strconv.Atoi(str); err == nil {
		return v, true
	}
	return 0, false
}

func GetAllConfig() map[ConfigName]string {
	acc := make(map[ConfigName]string, len(configValues))
	for k, v := range defaultConfigValues {
		acc[k] = v
	}
	for k, v := range configValues {
		acc[k] = v
	}
	return acc
}

func SetConfig(name ConfigName, value string) {
	configValues[name] = value
}
