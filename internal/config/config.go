package config

import "os"

// Config содержит переменные окружения для запуска сервера.
type Config struct {
	IP           string
	Port         string
	ErrorMessage string
}

// FromEnv собирает конфигурацию из переменных окружения.
func FromEnv() Config {
	return Config{
		IP:           os.Getenv("IP"),
		Port:         os.Getenv("PORT"),
		ErrorMessage: os.Getenv("ERROR"),
	}
}

// Address возвращает адрес в формате host:port.
func (c Config) Address() string {
	return c.IP + ":" + c.Port
}
