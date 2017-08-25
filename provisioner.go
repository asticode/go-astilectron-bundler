package astibundler

import "github.com/asticode/go-astilectron"

// Constants
const (
	zipNameAstilectron = "astilectron.zip"
	zipNameElectron    = "electron.zip"
)

// NewProvisioner builds the proper disembedder provisioner
func NewProvisioner(disembedFunc func(string) ([]byte, error)) astilectron.Provisioner {
	return astilectron.NewDisembedderProvisioner(disembedFunc, "vendor/"+zipNameAstilectron, "vendor/"+zipNameElectron)
}
