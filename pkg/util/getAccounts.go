package util

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"main/pkg/types"
)

func privateKeyToAddress(privateKey *ecdsa.PrivateKey) (*common.Address, error) {
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)
	return &address, nil
}

func GetAccounts(inputs []string) ([]types.AccountData, error) {
	var accounts []types.AccountData

	for _, input := range inputs {
		input = RemoveHexPrefix(input)

		sweepedPrivateKey, err := crypto.HexToECDSA(input)

		if err != nil {
			return nil, fmt.Errorf("invalid private key: %s", input)
		}

		sweepedAddress, err := privateKeyToAddress(sweepedPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to derive address: %s", err)
		}

		accounts = append(accounts, types.AccountData{
			AccountKeyHex:  input,
			AccountKey:     sweepedPrivateKey,
			AccountAddress: *sweepedAddress,
		})
	}

	return accounts, nil
}
