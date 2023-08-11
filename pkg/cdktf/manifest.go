package cdktf

type Manifest struct {
	Language           string   `json:"language"`
	App                string   `json:"app"`
	SendCrashReports   bool     `json:"sendCrashReports"`
	TerraformProviders []Source `json:"terraformProviders,omitempty"`
	TerraformModules   []Source `json:"terraformModules,omitempty"`
	ProjectID          string   `json:"projectId"`
	Comment            string   `json:"//"`
}

type Source struct {
	// Name is the name of the provider or module
	// This field is only used internally to render the `cdktf.json` template
	// It will be ignored in the incoming config file
	Name string `json:"name,omitempty"`

	// Source is the target provider or module to generate
	// e.g., registry.terraform.io/hashicorp/google
	Source string `json:"source"`
	// Version of the target provider or module to generate
	// e.g., "3.19.0"
	Version string `json:"version,omitempty"`
}
