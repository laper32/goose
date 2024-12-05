package conf

import (
	"fmt"

	"github.com/spf13/viper"
)

// /etc/mach
// ~/.mach/config

type ConfigData struct {
	Name string
	Type string
	Path []string
	Data interface{}
}

func Load(cfg ConfigData) {
	viper.SetConfigName(cfg.Name)
	viper.SetConfigType(cfg.Type)
	for _, path := range cfg.Path {
		viper.AddConfigPath(path)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err = viper.SafeWriteConfig(); err != nil {
				fmt.Println(err)
			}
		} else {
			panic(err)
		}
	}
	if err := viper.Unmarshal(cfg.Data); err != nil {
		panic(err)
	}

}
