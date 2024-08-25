package utils

import (
	"encoding/json"
	"os"

	"github.com/jessevdk/go-flags"
)

type ConfigFile struct {
	Path string `short:"c" long:"config" env:"CONFIG" default:"" description:"Path to the configuration file"`
}

func FillFromFile(cfg interface{}) error {
	var cf ConfigFile

	parser := flags.NewParser(&cf, flags.IgnoreUnknown)

	_, err := parser.Parse()
	if err != nil {
		return err
	}

	if cf.Path == "" {
		return err
	}

	data, err := os.ReadFile(cf.Path)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}
	return nil
}
