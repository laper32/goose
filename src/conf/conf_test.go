package conf

import (
	// "github.com/laper32/goose/conf"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/laper32/goose/db/pgsql"
	"github.com/laper32/goose/logging"

	goredis "github.com/redis/go-redis/v9"
)

type Services struct {
	Ban string
}

type Config struct {
	Log     *logging.Config
	Pgsql   *pgsql.Config
	Redis   *goredis.Options
	Service *Services
}

func New() (c *Config) {
	c = &Config{
		&logging.Config{},
		&pgsql.Config{},
		&goredis.Options{},
		&Services{},
	}
	Load(ConfigData{
		Name: "mach-auth",
		Type: "toml",
		Path: []string{"."},
		Data: c,
	})
	return
}

func TestConf(t *testing.T) {
	c := New()
	data, err := json.Marshal(c.Log)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	data, err = json.Marshal(c.Pgsql)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	fmt.Println(c.Redis)
	fmt.Println(c.Service)
}
