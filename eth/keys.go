package eth

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/pkg/errors"
)

func NewPrivKeyFromKeyStore(keystoreFile, passwordFile string, needPass bool) (*ecdsa.PrivateKey, error) {
	keyJson, err := os.ReadFile(keystoreFile)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read the keyfile at %s", keystoreFile))
	}
	passphrase := ""
	if needPass {
		passphrase, err = getPassphrase(passwordFile)
		if err != nil {
			return nil, err
		}
	}
	key, err := keystore.DecryptKey(keyJson, passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "Error decrypting key")
	}
	return key.PrivateKey, nil
}

func getPassphrase(passwordFile string) (string, error) {
	if passwordFile == "" {
		return utils.GetPassPhrase("", false), nil
	}
	password, err := os.ReadFile(passwordFile)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to read the keyfile at %s", passwordFile))
	}
	return strings.TrimRight(string(password), "\r\n"), nil
}
