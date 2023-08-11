package generator

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/cdktf-provider-gen/pkg/cdktf"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		b       []byte
		want    autogold.Value
		wantErr autogold.Value
	}{
		{
			name: "valid provider",
			b: []byte(`
name: google
provider:
  source: registry.terraform.io/hashicorp/google
  version: 4.69.1
target:
  language: go
  moduleName: github.com/sourcegraph/controller-cdktf/gen
  packageName: google
output: gen            
`),
			want: autogold.Expect(&Config{
				Name: "google", Provider: &cdktf.Source{
					Source:  "registry.terraform.io/hashicorp/google",
					Version: "4.69.1",
				},
				Target: &Target{Go: &GoTarget{
					Language:    "go",
					ModuleName:  "github.com/sourcegraph/controller-cdktf/gen",
					PackageName: "google",
				}},
				Output: "gen",
			}),
		},
		{
			name: "valid module",
			b: []byte(`
name: gkeprivate
module:
  source: terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster
  version: "24.0.0"
target:
  language: go
  moduleName: github.com/sourcegraph/controller-cdktf/gen
output: gen           
`),
			want: autogold.Expect(&Config{
				Name: "gkeprivate", Module: &cdktf.Source{
					Source:  "terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster",
					Version: "24.0.0",
				},
				Target: &Target{Go: &GoTarget{
					Language:    "go",
					ModuleName:  "github.com/sourcegraph/controller-cdktf/gen",
					PackageName: "gkeprivate",
				}},
				Output: "gen",
			}),
		},
		{
			name: "invalid: both provider and module",
			b: []byte(`
name: gkeprivate
provider:
  source: registry.terraform.io/hashicorp/google
  version: 4.69.1
module:
  source: terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster
  version: "24.0.0"
target:
  language: go
  moduleName: github.com/sourcegraph/controller-cdktf/gen
output: gen           
`),
			wantErr: autogold.Expect("provider and module can't be set at the same time"),
		},
		{
			name: "invalid - unsupported target",
			b: []byte(`
name: gkeprivate
module:
  source: terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster
  version: "24.0.0"
target:
  language: javascript
  moduleName: github.com/sourcegraph/controller-cdktf/gen
output: gen            
            `),
			wantErr: autogold.Expect(`unmarshal config file: error unmarshaling JSON: while decoding JSON: unknown target language "javascript"`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewConfig(tc.b)
			if tc.wantErr != nil {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
				return
			}
			require.NoError(t, err)
			tc.want.Equal(t, got)
		})
	}
}
