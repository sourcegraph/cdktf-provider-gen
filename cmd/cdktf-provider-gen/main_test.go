package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sourcegraph/cdktf-provider-gen/pkg/cdktf"
	"github.com/sourcegraph/cdktf-provider-gen/pkg/generator"
	"github.com/stretchr/testify/require"
)

func TestPackageJSONTemplateUsesCDKTerrain(t *testing.T) {
	data := projectTemplateData{
		Config: generator.Config{
			Name: "random",
			Provider: &cdktf.Source{
				Name: "random",
			},
		},
		PackageName: "random",
		ModuleName:  "github.com/sourcegraph/controller-cdktf/gen",
		Deps: cdktnDependencies{
			Jsii:       "~5.9.0",
			JsiiPacmak: "1.128.0",
			Constructs: "10.6.0",
			Cdktn:      "0.24.0",
		},
	}

	var rendered bytes.Buffer
	require.NoError(t, packageJSONTemplate.Execute(&rendered, data))

	var got struct {
		Name             string            `json:"name"`
		DevDependencies  map[string]string `json:"devDependencies"`
		PeerDependencies map[string]string `json:"peerDependencies"`
		Scripts          map[string]string `json:"scripts"`
	}
	require.NoError(t, json.Unmarshal(rendered.Bytes(), &got))
	require.Equal(t, "@cdktn/provider-random", got.Name)
	require.Equal(t, "0.24.0", got.DevDependencies["@cdktn/provider-generator"])
	require.Equal(t, "0.24.0", got.DevDependencies["cdktn"])
	require.Equal(t, "0.24.0", got.DevDependencies["cdktn-cli"])
	require.Equal(t, "0.24.0", got.PeerDependencies["cdktn"])
	require.Contains(t, got.Scripts["fetch"], "cdktn get")
	require.NotContains(t, rendered.String(), "@cdktf/")
}
