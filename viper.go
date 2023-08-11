package helper

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"
)

var (
	viperHelper     *ViperHelper
	viperHelperOnce sync.Once
)

type ViperHelper struct {
	V           *viper.Viper
	ConfigFile  string
	TagName     string
	EnableEnv   bool
	DecodeHooks []mapstructure.DecodeHookFunc
}

func Viper(options ...func(*ViperHelper)) *ViperHelper {
	viperHelperOnce.Do(func() {
		viperHelper = &ViperHelper{
			V:          viper.New(),
			ConfigFile: "etc/config.yaml",
			TagName:    "yaml",
			EnableEnv:  true,
			DecodeHooks: []mapstructure.DecodeHookFunc{
				mapstructure.StringToTimeDurationHookFunc(),
				StringToSliceHookFunc(","),
				mapstructure.OrComposeDecodeHookFunc(
					mapstructure.StringToTimeHookFunc(time.RFC3339),
					mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
				),
				UnmarshalToStructHookFunc(yaml.Unmarshal),
				UnmarshalToMapHookFunc(yaml.Unmarshal),
				UnmarshalToSliceHookFunc(yaml.Unmarshal),
			},
		}
	})
	for _, opt := range options {
		opt(viperHelper)
	}
	return viperHelper
}

func (h *ViperHelper) Unmarshal(conf any) error {
	if h.ConfigFile != "" {
		h.V.SetConfigFile(h.ConfigFile)
		err := h.V.ReadInConfig()
		if err != nil {
			return errors.Wrap(err, "read config failed")
		}
	}

	if h.EnableEnv {
		for _, env := range os.Environ() {
			key, _, ok := strings.Cut(env, "=")
			if !ok {
				return errors.New("cut env failed")
			}
			if strings.HasPrefix(key, "_") {
				continue
			}
			err := h.V.BindEnv(strings.ReplaceAll(key, "_", "."), key)
			if err != nil {
				return errors.Wrap(err, "bind env failed")
			}
		}
	}

	err := h.V.Unmarshal(
		conf,
		func(cfg *mapstructure.DecoderConfig) {
			cfg.TagName = h.TagName
			cfg.DecodeHook = mapstructure.ComposeDecodeHookFunc(h.DecodeHooks...)
		},
	)
	return errors.Wrap(err, "unmarshal config failed")
}
