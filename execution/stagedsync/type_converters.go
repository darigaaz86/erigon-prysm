package stagedsync

import (
	"math/big"

	"github.com/erigontech/erigon/common"
	coreTypes "github.com/erigontech/erigon/core/types"
	"github.com/erigontech/erigon/execution/types/accounts"
	"github.com/holiman/uint256"
)

// Type converters between core/types and execution/types

// coreAccountToExecAccount converts core/types.Account to execution/types/accounts.Account
func coreAccountToExecAccount(acc *coreTypes.Account) *accounts.Account {
	if acc == nil {
		return nil
	}

	execAcc := &accounts.Account{
		Nonce:       acc.Nonce,
		Balance:     acc.Balance, // Both use uint256.Int
		Incarnation: 0,           // Default incarnation
	}

	// Copy code hash
	copy(execAcc.CodeHash[:], acc.CodeHash[:])

	return execAcc
}

// execAccountToCoreAccount converts execution/types/accounts.Account to core/types.Account
func execAccountToCoreAccount(acc *accounts.Account) *coreTypes.Account {
	if acc == nil {
		return nil
	}

	coreAcc := &coreTypes.Account{
		Nonce:   acc.Nonce,
		Balance: acc.Balance, // Both use uint256.Int
	}

	// Copy code hash
	copy(coreAcc.CodeHash[:], acc.CodeHash[:])

	return coreAcc
}

// uint256ToBigInt converts uint256.Int to *big.Int
func uint256ToBigInt(u *uint256.Int) *big.Int {
	if u == nil {
		return big.NewInt(0)
	}
	return u.ToBig()
}

// bigIntToUint256 converts *big.Int to uint256.Int
func bigIntToUint256(b *big.Int) *uint256.Int {
	if b == nil {
		return uint256.NewInt(0)
	}
	u := uint256.NewInt(0)
	u.SetFromBig(b)
	return u
}

// hashToBytes converts common.Hash to []byte
func hashToBytes(h common.Hash) []byte {
	return h.Bytes()
}

// bytesToHash converts []byte to common.Hash
func bytesToHash(b []byte) common.Hash {
	return common.BytesToHash(b)
}
