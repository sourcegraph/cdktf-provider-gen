package observability

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/run"

	"github.com/sourcegraph/cdktf-provider-gen/internal/observability/internal/resource"
)

// InitLogs initializes the logger with the provided resource name
// and returns the logger
// Make sure to call logger.Sync() when the program exits
// This should be called in the entrypoint (e.g. main function of the executable) of the program
func InitLogs(name string, buildCommit string) *log.PostInitCallbacks {
	if _, set := os.LookupEnv(log.EnvLogFormat); !set {
		os.Setenv(log.EnvLogFormat, "console")
		if _, set := os.LookupEnv(log.EnvDevelopment); !set {
			os.Setenv(log.EnvDevelopment, "true")
		}
	}
	if _, set := os.LookupEnv(log.EnvLogLevel); !set {
		os.Setenv(log.EnvLogLevel, "info")
	}

	return log.Init(resource.BuildLogResource(name, buildCommit))

}

// LogCommands logs unredacted command args to the provided logger
func LogCommands(ctx context.Context, logger log.Logger) context.Context {
	return run.LogCommands(ctx, func(command run.ExecutedCommand) {
		logger.Debug("running", log.String("cmd", strings.Join(command.Args, " ")))
	})
}

// TerraformPrintfer implements the `terraform-exec` logger interface,
// so we can print out `terraform-exec` internal logs
type TerraformPrintfer struct {
	Logger log.Logger
}

func (t TerraformPrintfer) Printf(format string, v ...any) {
	t.Logger.Debug(fmt.Sprintf(format, v...))
}

func buildLogResource(name string, buildCommit string) log.Resource {
	return log.Resource{
		Name:    name,
		Version: buildCommit,
		InstanceID: func() string {
			if envHostname := os.Getenv("HOSTNAME"); envHostname != "" {
				return envHostname
			}
			h, _ := os.Hostname()
			return h
		}(),
	}
}
