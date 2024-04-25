package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Ip      string
	Port    string
	Root    string
	Route   string
	Timeout int
}

func defaultConfig() Config {
	return Config{
		Ip:      "0.0.0.0",
		Port:    "80",
		Root:    "./symbols",
		Route:   "/download/symbols/",
		Timeout: 300, // Seconds
	}
}

type ConfigParser struct {
}

func newConfingParser() *ConfigParser {
	return &ConfigParser{}
}

func (parser *ConfigParser) loadFile(path string, c interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to load config: ", err)
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		fmt.Println("Failed to parse config: ", err)
		return err
	}

	return nil
}

func LoadConfig(path string) Config {
	c := defaultConfig()
	configParser := newConfingParser()
	configParser.loadFile(path, &c)

	// absolutePath, err := filepath.Abs(c.Root)
	// if err != nil {
	// 	return c
	// }

	// c.Root = absolutePath

	c.Route = mergeSlashes("/" + c.Route + "/")

	return c
}
