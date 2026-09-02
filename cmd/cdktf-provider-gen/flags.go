package main

import (
	"github.com/urfave/cli/v2"
)

var (
	configFlag = &cli.StringFlag{
		Name:     "config",
		Aliases:  []string{"c"},
		Required: true,
	}
	cdktnVersionFlag = &cli.StringFlag{
		Name:    "cdktn-version",
		Usage:   "The target CDK-Terrain version to use",
		Value:   "0.24.0",
		EnvVars: []string{"CDKTN_VERSION"},
	}
	keepFlag = &cli.BoolFlag{
		Name:  "keep",
		Usage: "Retain the intermediate assets, useful for debugging codegen error",
	}
)
