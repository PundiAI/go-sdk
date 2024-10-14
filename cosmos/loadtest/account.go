package loadtest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
)

type Account struct {
	PrivKey       cryptotypes.PrivKey `json:"-"`
	PrivateKey    string              `json:"private_key"`
	KeyType       string              `json:"key_type"`
	Mnemonic      string              `json:"mnemonic"`
	Address       string              `json:"address"`
	AccountNumber uint64              `json:"-"`
	Sequence      uint64              `json:"-"`
}

var _ sort.Interface = (*Accounts)(nil)

type Accounts struct {
	accounts []*Account
	sync.RWMutex
	cacheIndex int
}

func (a *Accounts) Len() int {
	return len(a.accounts)
}

func (a *Accounts) Less(i, j int) bool {
	return a.accounts[i].AccountNumber < a.accounts[j].AccountNumber
}

func (a *Accounts) Swap(i, j int) {
	a.accounts[i], a.accounts[j] = a.accounts[j], a.accounts[i]
}

func (a *Accounts) NextAccount() *Account {
	a.RLock()
	defer a.RUnlock()
	account := a.accounts[a.cacheIndex]
	a.cacheIndex++
	if a.cacheIndex >= len(a.accounts) {
		a.cacheIndex = 0
	}
	return account
}

func (a *Accounts) IsFistAccount() bool {
	return a.cacheIndex == 1
}

func NewAccounts(client RPCClient, keyDir string) (*Accounts, error) {
	prefix, err := client.GetAddressPrefix()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get address prefix")
	}
	entries, err := os.ReadDir(keyDir)
	if err != nil {
		return nil, err
	}
	accountChan := make(chan *Account, len(entries))
	wg := sync.WaitGroup{}
	jobs := make(chan int, 64)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		jobs <- 1
		wg.Add(1)
		go func(client RPCClient, keyFile string) {
			defer func() {
				wg.Done()
				<-jobs
			}()
			var account *Account
			account, err = newAccount(keyFile, prefix)
			if err != nil {
				panic(err)
			}
			var accountI types.AccountI
			accountI, err = client.QueryAccount(account.Address)
			if err != nil {
				panic(err)
			}
			account.AccountNumber = accountI.GetAccountNumber()
			account.Sequence = accountI.GetSequence()
			accountChan <- account
		}(client, filepath.Join(keyDir, entry.Name()))
	}
	wg.Wait()
	close(jobs)

	accounts := make([]*Account, len(accountChan))
	for i := 0; i < len(accounts); i++ {
		accounts[i] = <-accountChan
	}
	close(accountChan)

	return &Accounts{
		accounts: accounts,
		RWMutex:  sync.RWMutex{},
	}, nil
}

func NewAccountFromGenesis(genesisFilePath, keyDir string) (*Accounts, error) {
	genesisFile, err := os.ReadFile(genesisFilePath)
	if err != nil {
		return nil, err
	}
	var genesis struct {
		AppState struct {
			Auth struct {
				Accounts []struct {
					Address       string `json:"address"`
					AccountNumber string `json:"account_number"`
					Sequence      string `json:"sequence"`
				} `json:"accounts"`
			} `json:"auth"`
		} `json:"app_state"`
	}
	if err = json.Unmarshal(genesisFile, &genesis); err != nil {
		return nil, err
	}
	var prefix string
	adminAddrStr := genesis.AppState.Auth.Accounts[0].Address
	if !strings.HasPrefix(adminAddrStr, "0x") {
		prefix, _, err = bech32.DecodeAndConvert(adminAddrStr)
		if err != nil {
			return nil, err
		}
	}

	entries, err := os.ReadDir(keyDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, errors.New("no key files found")
	}
	accounts := make([]*Account, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var account *Account
		account, err = newAccount(filepath.Join(keyDir, entry.Name()), prefix)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	for _, account := range genesis.AppState.Auth.Accounts {
		for _, acc := range accounts {
			if account.Address == acc.Address {
				acc.AccountNumber, err = strconv.ParseUint(account.AccountNumber, 10, 64)
				if err != nil {
					return nil, err
				}
				acc.Sequence, err = strconv.ParseUint(account.Sequence, 10, 64)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &Accounts{
		accounts: accounts,
		RWMutex:  sync.RWMutex{},
	}, nil
}

func newAccount(keyFile string, prefix string) (*Account, error) {
	bz, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	var account Account
	if err = json.Unmarshal(bz, &account); err != nil {
		return nil, err
	}

	var privBz []byte
	if account.PrivateKey != "" {
		privBz = common.FromHex(account.PrivateKey)
	} else {
		if account.Mnemonic == "" {
			return nil, errors.New("private key and mnemonic are both empty")
		}
		coinType := uint32(118)
		if account.KeyType == ethsecp256k1.KeyType {
			coinType = 60
		}
		privBz, err = hd.Secp256k1.Derive()(account.Mnemonic, "", hd.CreateHDPath(coinType, 0, 0).String())
		if err != nil {
			return nil, err
		}
	}

	if prefix == "0x" || account.KeyType == ethsecp256k1.KeyType {
		account.PrivKey = &ethsecp256k1.PrivKey{Key: privBz}
		accAddress := common.BytesToAddress(account.PrivKey.PubKey().Address())
		if account.Address != "" && account.Address != accAddress.String() {
			return nil, errors.Errorf("address not match expected: %s actual: %s", account.Address, accAddress.String())
		}
		account.Address = accAddress.String()
	} else {
		account.PrivKey = &secp256k1.PrivKey{Key: privBz}
		accAddress, err := bech32.ConvertAndEncode(prefix, account.PrivKey.PubKey().Address())
		if err != nil {
			return nil, err
		}
		if account.Address != "" && account.Address != accAddress {
			return nil, errors.Errorf("address not match expected: %s actual: %s", account.Address, accAddress)
		}
		account.Address = accAddress
	}
	return &account, nil
}

func CreateGenesisAccounts(addrPrefix string, number int, outDir string) error {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil
	}
	for i := 0; i < number; i++ {
		bip44Params := hd.NewFundraiserParams(118, 0, 0)
		var privBz []byte
		privBz, err = hd.Secp256k1.Derive()(mnemonic, "", bip44Params.String())
		if err != nil {
			return err
		}

		var privKey cryptotypes.PrivKey
		var address string
		if addrPrefix == "0x" {
			privKey = &ethsecp256k1.PrivKey{Key: privBz}
			address = common.BytesToAddress(privKey.PubKey().Address()).String()
		} else {
			privKey = &secp256k1.PrivKey{Key: privBz}
			address, err = bech32.ConvertAndEncode(addrPrefix, privKey.PubKey().Address())
			if err != nil {
				return err
			}
		}

		var data []byte
		data, err = json.Marshal(map[string]string{
			"mnemonic": mnemonic,
			"address":  address,
		})
		if err != nil {
			return err
		}

		fileName := filepath.Join(outDir, "test_"+strconv.Itoa(i))
		if err = os.WriteFile(fileName, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}
