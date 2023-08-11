package main

import (
	"github.com/urfave/cli/v2"
)

var (
	configFlag = &cli.StringFlag{
		Name:     "config",
		Required: true,
	}
	cdktfVersionFlag = &cli.StringFlag{
		Name:    "cdktf-version",
		Usage:   "The target cdktf version to use",
		Value:   "0.16.3",
		EnvVars: []string{"CDKTF_VERSION"},
	}
)
