package pgsql

import (
	"fmt"
	"time"

	"github.com/laper32/goose/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host         string
	User         string
	Password     string
	DBName       string
	Debug        bool
	MaxOpenConns int
	MaxIdleConns int
}

var client *gorm.DB

func (conf *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=Asia/Shanghai", conf.Host, conf.User, conf.Password, conf.DBName)
}

func (conf *Config) Init() {
	if conf.Host == "" {
		conf.Host = "127.0.0.1"
	}
	if conf.User == "" {
		conf.User = "postgres"
	}
	if conf.Password == "" {
		conf.Password = "password"
	}
	if conf.DBName == "" {
		conf.DBName = "postgres"
	}
}

func Init(conf *Config) *gorm.DB {
	var err error
	conf.Init()
	conn := func() (*gorm.DB, error) {
		return gorm.Open(postgres.Open(conf.GetDSN()), &gorm.Config{})
	}
	// Retry
	for i := 0; i < 3; i++ {
		client, err = conn()
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	// Set conns pool
	if err == nil {
		sqlDB, err := client.DB()
		if err == nil {
			sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
			sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
		}
	}

	if err != nil {
		panic(err)
	}
	return client
}

func Healthy() bool {
	if client != nil {
		db, err := client.DB()
		if err == nil {
			err = db.Ping()
			if err == nil {
				return true
			}
		}
		logging.Error(err)
	}
	return true
}

func Close() {

}
