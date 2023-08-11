package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	hcversion "github.com/hashicorp/go-version"
	hcproduct "github.com/hashicorp/hc-install/product"
	tfreleases "github.com/hashicorp/hc-install/releases"
	cp "github.com/otiai10/copy"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/run"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/cdktf-provider-gen/internal/observability"
	"github.com/sourcegraph/cdktf-provider-gen/internal/output"
	"github.com/sourcegraph/cdktf-provider-gen/pkg/cdktf"
	"github.com/sourcegraph/cdktf-provider-gen/pkg/generator"
)

func main() {
	liblog := observability.InitLogs("cdktf-provider-gen", "dev")
	defer liblog.Sync()

	sort.Sort(cli.CommandsByName(gen.Commands))
	sort.Sort(cli.FlagsByName(gen.Flags))

	if err := gen.Run(os.Args); err != nil {
		_ = output.Render(output.FormatText, err)
		os.Exit(1)
	}
}

var (
	//go:embed package.json
	packageJSONTemplateString string
	packageJSONTemplate       = template.Must(template.New("").Parse(packageJSONTemplateString))
)

type projectTemplateData struct {
	Config      generator.Config
	PackageName string
	ModuleName  string

	Deps cdktfDependencies
}

var gen = &cli.App{
	Name: "cdktf-provider-gen",
	Flags: []cli.Flag{
		configFlag,
		cdktfVersionFlag,
	},
	UsageText: `
# Generate the googla provider
cdktf-provider-gen -concifg google.yaml

# Use a specific version of cdktf
cdktf-provider-gen -config google.yaml -cdktf-version 0.17.3
    `,
	Action: func(c *cli.Context) error {
		logger := log.Scoped("gen", "generate cdktf provider code")

		// workarounad for lack of well supported terraform toolchains for bazel
		// so we need to bring our own terraform and configure it in the path
		// so the cdktf-cli npm package can access it
		tfInstallDir, err := os.MkdirTemp("", "tf-bin")
		if err != nil {
			return errors.Wrap(err, "create temp tf-bin dir")
		}
		defer os.RemoveAll(tfInstallDir)
		installer := &tfreleases.ExactVersion{
			Product: hcproduct.Terraform,
			Version: hcversion.Must(hcversion.NewVersion("1.5.5")),
		}
		installer.InstallDir = tfInstallDir
		_, err = installer.Install(c.Context)
		if err != nil {
			return errors.Wrap(err, "install terraform")
		}
		_ = os.Setenv("PATH", tfInstallDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		b, err := os.ReadFile(configFlag.Get(c))
		if err != nil {
			return errors.Wrap(err, "read config file")
		}
		config, err := generator.NewConfig(b)
		if err != nil {
			return errors.Wrapf(err, "parse config file %q", configFlag.Get(c))
		}

		logger = logger.With(log.String("name", config.Name))
		if config.Provider != nil {
			logger = logger.With(
				log.String("name", config.Name),
				log.String("provider.name", config.Provider.Name),
				log.String("provider.version", config.Provider.Version),
			)
		}
		if config.Module != nil {
			logger = logger.With(
				log.String("module.source", config.Module.Source),
				log.String("module.version", config.Module.Version),
			)
		}

		m := cdktf.Manifest{
			Language:         "typescript",
			App:              "echo noop",
			SendCrashReports: false,
			ProjectID:        "noop",
		}
		if config.Provider != nil {
			// this is a special handling for provider name with hyphens
			providerName, ok := Last(strings.Split(config.Provider.Source, "/"))
			if !ok {
				return errors.Newf("provider name not found: %q", config.Provider.Source)
			}
			config.Provider.Name = providerName
			m.TerraformProviders = []cdktf.Source{
				*config.Provider,
			}
		}
		if config.Module != nil {
			config.Module.Name = config.Name
			m.TerraformModules = []cdktf.Source{
				*config.Module,
			}
		}
		var cdktfJSON bytes.Buffer
		enc := json.NewEncoder(&cdktfJSON)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(m); err != nil {
			return errors.Wrap(err, "marshal cdktf.json")
		}

		deps, err := fetchCdktfDependencies(c.Context, cdktfVersionFlag.Get(c))
		if err != nil {
			return errors.Wrap(err, "fetch cdktf dependencies")
		}
		deps.Cdktf = cdktfVersionFlag.Get(c)

		data := projectTemplateData{
			Config:      *config,
			PackageName: config.Target.Go.PackageName,
			ModuleName:  config.Target.Go.ModuleName,
			Deps:        *deps,
		}
		var packageJSON bytes.Buffer
		if err := packageJSONTemplate.Execute(&packageJSON, data); err != nil {
			return errors.Wrap(err, "render package.json")
		}

		tmpDir, err := os.MkdirTemp("", "cdktfprovidergen")
		if err != nil {
			return errors.Wrap(err, "create temp dir")
		}
		defer os.RemoveAll(tmpDir)
		logger = logger.With(log.String("tmpDir", tmpDir))

		logger.Debug("write package.json")
		if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), packageJSON.Bytes(), 0644); err != nil {
			return errors.Wrap(err, "write package.json")
		}
		logger.Debug("write cdktf.json")
		if err := os.WriteFile(filepath.Join(tmpDir, "cdktf.json"), cdktfJSON.Bytes(), 0644); err != nil {
			return errors.Wrap(err, "write cdktf.json")
		}

		logger.Debug("compiling cdktf provider code")
		cmdCtx := observability.LogCommands(c.Context, logger)
		for _, cmd := range []string{
			"npm install --no-save",
			"npm run fetch",
			"npm run compile",
			"npm run pkg:go",
		} {
			if err := run.Cmd(cmdCtx, cmd).Dir(tmpDir).Run().Wait(); err != nil {
				return errors.Wrapf(err, "run: %q", cmd)
			}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "get working dir")
		}
		srcDir := filepath.Join(tmpDir, "dist", "go", config.Target.Go.PackageName)
		outputDir := filepath.Join(cwd, config.Output, config.Target.Go.PackageName)
		logger = logger.With(log.String("srcDir", srcDir), log.String("outputDir", outputDir))
		logger.Debug("ensuring output dir is clean")
		if _, err := os.Stat(outputDir); err == nil {
			if err := os.RemoveAll(outputDir); err != nil {
				return errors.Wrapf(err, "clean output dir %q", outputDir)
			}
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return errors.Wrap(err, "create output dir")
		}
		logger.Debug("copying to output dir")
		if err := cp.Copy(srcDir, outputDir); err != nil {
			return errors.Wrap(err, "copy cdktf.out")
		}

		return nil
	},
}

func Last[E any](s []E) (E, bool) {
	if len(s) == 0 {
		var zero E
		return zero, false
	}
	return s[len(s)-1], true
}

type cdktfDependencies struct {
	Jsii       string
	JsiiPacmak string
	Constructs string
	Cdktf      string
}

func fetchCdktfDependencies(ctx context.Context, version string) (*cdktfDependencies, error) {
	npmAPIURL := fmt.Sprintf("https://registry.npmjs.org/cdktf/%s", version)

	response, err := http.Get(npmAPIURL)
	if err != nil {
		return nil, errors.Wrap(err, "fetch cdktf version from registry")
	}
	defer response.Body.Close()

	var resp struct {
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, errors.Wrap(err, "decode cdktf version response")
	}

	deps := &cdktfDependencies{}
	if v, ok := resp.DevDependencies["jsii"]; ok {
		deps.Jsii = v
	} else {
		return nil, errors.New("jsii version not found")
	}
	if v, ok := resp.DevDependencies["jsii-pacmak"]; ok {
		deps.JsiiPacmak = v
	} else {
		return nil, errors.New("jsii-pacmak version not found")
	}
	if v, ok := resp.DevDependencies["constructs"]; ok {
		deps.Constructs = v
	} else {
		return nil, errors.New("constructs version not found")
	}
	return deps, nil
}
