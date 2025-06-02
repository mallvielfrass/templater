package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	HttpPort int
	// DBUrl        string
	// DBName       string
	// JWTSecret    string
	BadgerDBPath string
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LookupAndParseEnvInt(envName string) (int, bool) {
	env, exists := os.LookupEnv(envName)
	if !exists {
		//fmc.Printfln("#rbtError#wbt: #bbt%s", fmt.Errorf("env '%s' not found", envName).Error())
		fmt.Printf("Error: env '%s' not found", envName)
		return 0, false
	}
	parsedInt, err := strconv.Atoi(env)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return 0, false
	}
	return parsedInt, true
}
func NewConfig(configPath string) (*Config, error) {
	var conf Config
	if fileExists(configPath) {
		err := godotenv.Load(configPath)
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("config file not found")
	}
	//get Config from environment
	httpPort, _ := LookupAndParseEnvInt("HttpPort")

	conf.HttpPort = httpPort

	conf.BadgerDBPath = os.Getenv("BadgerDBPath")

	return &conf, nil
}
func (c *Config) Validate() error {

	if c.HttpPort == 0 {
		return fmt.Errorf("http port is not set")
	}

	if c.BadgerDBPath == "" {
		return fmt.Errorf("badger db path is not set")
	}

	return nil
}
