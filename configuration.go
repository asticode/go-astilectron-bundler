package astibundler

// Configuration represents the bundle configuration
type Configuration struct {
	AppName            string                     `json:"app_name"`
	AppIconDarwinPath  string                     `json:"app_icon_darwin_path"` // Darwin systems requires a specific .icns file
	AppIconDefaultPath string                     `json:"app_icon_default_path"`
	Environments       []ConfigurationEnvironment `json:"environments"`
	InputPath          string                     `json:"input_path"`
	OutputPath         string                     `json:"output_path"`
}

// ConfigurationEnvironment represents the bundle configuration environment
type ConfigurationEnvironment struct {
	Arch string `json:"arch"`
	OS   string `json:"os"`
}
