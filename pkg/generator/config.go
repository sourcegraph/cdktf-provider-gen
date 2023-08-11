package generator

import (
	"encoding/json"

	"github.com/sourcegraph/cdktf-provider-gen/pkg/cdktf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"sigs.k8s.io/yaml"
)

type Config struct {
	// Name is the name of the provider or module
	// This is used as the output go module suffix
	Name string `json:"name"`

	Provider *cdktf.Source `json:"provider"`
	Module   *cdktf.Source `json:"module"`

	// Target is the config of the target language
	Target *Target `json:"target"`

	// Output is the parent direcotry to write the generated code to.
	// The final output directory will be <output>/<Target.Go.PackageName>
	Output string `json:"output"`
}

type Target struct {
	Go *GoTarget
}

type GoTarget struct {
	// Language of the generated code. Only "go" is supported at the moment.
	Language string `json:"language"`

	// ModuleName is the root module name, e.g., github.com/sourcegraph/controller-cdktf/gen
	ModuleName string `json:"moduleName"`
	// PackagName is the output package under the provided module above, e.g., google
	// If empty, defaults to the provider name.
	// The final full package path will be <moduleName>/<packageName>
	PackageName string `json:"packageName"`
}

func NewConfig(b []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, errors.Wrap(err, "unmarshal config file")
	}
	if c.Provider != nil && c.Module != nil {
		return nil, errors.New("provider and module can't be set at the same time")
	}
	if c.Target == nil {
		return nil, errors.New("language target config is required")
	}
	if c.Target.Go == nil {
		return nil, errors.New("go target config is required")
	}
	if c.Target.Go.PackageName == "" {
		c.Target.Go.PackageName = c.Name
	}
	return &c, nil
}

func (t *Target) UnmarshalJSON(b []byte) error {
	var d struct {
		Language string `json:"language"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		return err
	}
	switch d.Language {
	case "go":
		return json.Unmarshal(b, &t.Go)
	}
	return errors.Newf("unknown target language %q", d.Language)
}

func (t Target) MarshalJSON() ([]byte, error) {
	if t.Go != nil {
		return json.Marshal(t.Go)
	}
	return nil, errors.New("target must have exactly 1 non-nil config")
}
