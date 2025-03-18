package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

func loadConfig(c *config) error {
	file, err := os.Open("./config/base.yaml")
	if err != nil {
		return err
	}
	err = yaml.NewDecoder(file).Decode(c)
	return err
}
