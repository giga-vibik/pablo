package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var cfg Config

type Config struct {
	Postgres  DatabaseConfig `yaml:"DB"`
	DebugFlag DebugFlag      `yaml:"Debug_flag"`
	S3        S3             `yaml:"S3"`
	Zernio    Zernio         `yaml:"Zernio"`
	Auth      Auth           `yaml:"Auth"`
	Publisher Publisher      `yaml:"Publisher"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBname   string `yaml:"dbname"`
}

type DebugFlag struct {
	Flag bool `yaml:"flag"`
}

// S3 — хранилище видео. Бакет должен отдавать объекты публично: zernio
// забирает медиа по URL, а не принимает файл.
type S3 struct {
	Endpoint   string `yaml:"endpoint"`
	Region     string `yaml:"region"`
	Bucket     string `yaml:"bucket"`
	AccessKey  string `yaml:"accessKey"`
	SecretKey  string `yaml:"secretKey"`
	PublicBase string `yaml:"publicBase"` // напр. https://pablo.storage.yandexcloud.net
}

// Zernio — внешний агрегатор публикаций (getlate.dev).
// Профиль клиент находит сам, поэтому в конфиге нужен только ключ.
type Zernio struct {
	APIKey  string `yaml:"apiKey"`
	BaseURL string `yaml:"baseURL"` // пусто → https://zernio.com/api/v1
}

type Auth struct {
	Secret string `yaml:"secret"`
	// Единственный логин/пароль — сервис внутренний, регистрации нет.
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

type Publisher struct {
	// Как часто воркер проверяет посты, которым подошло время публикации.
	TickSeconds int `yaml:"tickSeconds"`
	// Сколько постов забирать за тик.
	BatchSize int `yaml:"batchSize"`
}

func InitConfig() (*Config, error) {
	configPath := "cmd/config.yaml"

	clean := filepath.Clean(configPath)

	file, err := os.Open(clean)
	if err != nil {
		return nil, fmt.Errorf("error: fail to open config file in path %q with error %w", configPath, err)
	}
	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error: fail to parse config %w", err)
	}

	return &cfg, nil
}

func GetConfig() *Config { return &cfg }
