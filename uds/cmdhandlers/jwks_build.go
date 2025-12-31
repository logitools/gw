package cmdhandlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/security"
)

type JwksBuild struct {
	AppProvider framework.AppProviderFunc
}

func (*JwksBuild) GroupName() string {
	return "jwks"
}

func (h *JwksBuild) Command() string {
	return "jwks-build"
}

func (h *JwksBuild) Desc() string {
	return "Rebuild jwks.json file from key files"
}

func (h *JwksBuild) Usage() string {
	return h.Command() + " key"
}

func (h *JwksBuild) HandleCommand(_ []string, w io.Writer) error {
	appCore := h.AppProvider().AppCore()
	// Load Pem Files into JWKS
	jwks, err := security.LoadPublicPEMKeysAsJWKS(appCore.StorageConf.KeyStoreConf.PublicKeyDir) // *security.Jwks
	if err != nil {
		return fmt.Errorf("failed to load public keys. %v", err)
	}
	// Write jwks.json.tmp file
	// NOTE: `gen` directory must be prepared and writable
	tmpFilePath := filepath.Join(appCore.AppRoot, "gen", "jwks.json.tmp")
	if err = jwks.CreateJSONFile(tmpFilePath); err != nil {
		return fmt.Errorf("failed to build jwks.json.tmp file %v", err)
	}
	// Rename (=atomic) to the real one
	filePath := filepath.Join(appCore.AppRoot, "gen", "jwks.json")
	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return fmt.Errorf("failed to write jwks.json file %v", err)
	}
	_, _ = fmt.Fprintln(w, "jwks.json created")
	return nil
}
