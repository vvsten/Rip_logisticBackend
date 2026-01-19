package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	// PublicBaseURL - внешний URL этого сервиса, который будет использоваться вторым сервисом для callback.
	// Пример: http://localhost:8083 или http://192.168.1.64:8083
	PublicBaseURL string

	// HTTPS Configuration
	EnableHTTPS bool
	CertFile    string
	KeyFile     string

	// JWT Configuration
	JWTSecret             string
	JWTAccessTokenExpire  int
	JWTRefreshTokenExpire int

	// Redis Configuration
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// Lab8: async сервис (второй сервис) + псевдо-авторизация токеном
	AsyncServiceURL   string
	AsyncServiceToken string
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}           // создаем объект конфига
	err = viper.Unmarshal(cfg) // читаем информацию из файла,
	// конвертируем и затем кладем в нашу переменную cfg
	if err != nil {
		return nil, err
	}

	log.Info("config parsed")

	return cfg, nil
}
