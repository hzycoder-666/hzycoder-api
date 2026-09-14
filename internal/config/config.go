package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`

	Database struct {
		DSN             string `mapstructure:"dsn"`
		MaxOpenConns    int    `mapstructure:"max_open_conns"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns"`
		ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
		ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
	} `mapstructure:"database"`

	Log struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`

	JWT struct {
		Secret string `mapstructure:"secret"`
		Expire int    `mapstructure:"expire"`
	} `mapstructure:"jwt"`

	Auth struct {
		AllowAdminRegister bool `mapstructure:"allow_admin_register"`
		LoginRateLimit     int  `mapstructure:"login_rate_limit"`
		RegisterRateLimit  int  `mapstructure:"register_rate_limit"`
	} `mapstructure:"auth"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)
	bindEnv(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.max_open_conns", 20)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 3600)
	v.SetDefault("database.conn_max_idle_time", 1800)
	v.SetDefault("log.level", "info")
	v.SetDefault("jwt.expire", 7200)
	v.SetDefault("auth.allow_admin_register", false)
	v.SetDefault("auth.login_rate_limit", 10)
	v.SetDefault("auth.register_rate_limit", 5)
}

func bindEnv(v *viper.Viper) {
	keys := []string{
		"server.port",
		"database.dsn",
		"database.max_open_conns",
		"database.max_idle_conns",
		"database.conn_max_lifetime",
		"database.conn_max_idle_time",
		"log.level",
		"jwt.secret",
		"jwt.expire",
		"auth.allow_admin_register",
		"auth.login_rate_limit",
		"auth.register_rate_limit",
	}

	for _, key := range keys {
		_ = v.BindEnv(key)
	}
}
