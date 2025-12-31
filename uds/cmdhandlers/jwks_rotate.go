package cmdhandlers

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/security"
)

type JwksRotate struct {
	AppProvider framework.AppProviderFunc
}

func (h *JwksRotate) GroupName() string {
	return "jwks"
}

func (h *JwksRotate) Command() string {
	return "jwks-rotate"
}

func (h *JwksRotate) Desc() string {
	return "Rotate RSA key pair and build jwks"
}

func (h *JwksRotate) Usage() string {
	return h.Command()
}

func (h *JwksRotate) HandleCommand(_ []string, w io.Writer) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048) // *rsa.PrivateKey
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %v", err)
	}
	publicKey := &privateKey.PublicKey                 // *rsa.PublicKey
	keyId, err := security.GenerateKeyID(publicKey, 8) // string
	if err != nil {
		return fmt.Errorf("failed to generate a key id. %v", err)
	}
	appCore := h.AppProvider().AppCore()
	// Store PEM keys
	privateKeyStorePath := appCore.StorageConf.KeyStoreConf.PrivateKeyDir + "/" + keyId + "_private.pem"
	if err = security.SavePrivatePEMKeyLocal(privateKeyStorePath, privateKey); err != nil {
		return fmt.Errorf("failed to save private key. %v", err)
	}
	publicKeyStorePath := appCore.StorageConf.KeyStoreConf.PublicKeyDir + "/" + keyId + "_public.pem"
	if err = security.SavePublicPEMKeyLocal(publicKeyStorePath, publicKey); err != nil {
		return fmt.Errorf("failed to save public key. %v", err)
	}
	// Update Current Key id
	if err = appCore.KVDBClient.Set(appCore.RootCtx, appCore.AppName+":kid", keyId, 0); err != nil {
		return fmt.Errorf("failed to save current key. %v", err)
	}
	// Load Pem Files into JWKS
	jwks, err := security.LoadPublicPEMKeysAsJWKS(appCore.StorageConf.KeyStoreConf.PublicKeyDir) // *security.Jwks
	if err != nil {
		return fmt.Errorf("failed to load public keys. %v", err)
	}
	// Write jwks.json.tmp file
	tmpFilePath := filepath.Join(appCore.AppRoot, "gen", "jwks.json.tmp")
	if err = jwks.CreateJSONFile(tmpFilePath); err != nil {
		return fmt.Errorf("failed to build jwks.json.tmp file %v", err)
	}
	// Rename (=atomic) to the real one
	filePath := filepath.Join(appCore.AppRoot, "gen", "jwks.json")
	if err = os.Rename(tmpFilePath, filePath); err != nil {
		return fmt.Errorf("failed to write jwks.json file %v", err)
	}
	_, _ = fmt.Fprintf(w, "jwks.json created\npublic key: %v\nkid: %s\n", publicKey, keyId)
	return nil
}
