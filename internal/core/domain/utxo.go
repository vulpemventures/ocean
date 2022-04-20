package domain

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil"
)

var (
	ErrUtxoAlreadyLocked = fmt.Errorf("utxo is already locked")
)

// UtxoKey represents the key of an Utxo, composed by its txid and vout.
type UtxoKey struct {
	TxID string
	VOut uint32
}

func (k UtxoKey) Hash() string {
	buf, _ := hex.DecodeString(k.TxID)
	buf = append(buf, byte(k.VOut))
	return hex.EncodeToString(btcutil.Hash160(buf))
}

func (k UtxoKey) String() string {
	return fmt.Sprintf("{%s: %d}", k.TxID, k.VOut)
}

// UtxoInfo holds sensitive info about the utxo. For confidential utxos.
// they must be revealed to return useful UtxoInfo.
type UtxoInfo struct {
	UtxoKey
	Value        uint64
	Asset        string
	Script       []byte
	ValueBlinder []byte
	AssetBlinder []byte
	AccountName  string
}

func (i UtxoInfo) Key() UtxoKey {
	return i.UtxoKey
}

// Balance holds info about the balance of a list of utxos with the same asset.
type Balance struct {
	Confirmed   uint64
	Unconfirmed uint64
	Locked      uint64
}

func (b *Balance) Total() uint64 {
	return b.Confirmed + b.Unconfirmed + b.Locked
}

// Utxo is the data structure representing an Elements UTXO with extra info
// like whether it is spent/utxo, confirmed/unconfirmed or locked/unlocked and
// the name of the account owning it.
type Utxo struct {
	UtxoKey
	Value           uint64
	Asset           string
	ValueCommitment []byte
	AssetCommitment []byte
	ValueBlinder    []byte
	AssetBlinder    []byte
	Script          []byte
	Nonce           []byte
	RangeProof      []byte
	SurjectionProof []byte
	AccountName     string
	LockTimestamp   int64
	Spent           bool
	Confirmed       bool
}

// IsSpent returns whether the utxo have been spent.
func (u *Utxo) IsSpent() bool {
	return u.Spent
}

// IsConfirmed returns whether the utxo is confirmed.
func (u *Utxo) IsConfirmed() bool {
	return u.Confirmed
}

// IsConfidential returns whether the utxo is a confidential one.
func (u *Utxo) IsConfidential() bool {
	return len(u.ValueCommitment) > 0 && len(u.AssetCommitment) > 0
}

// IsRevealed returns whether the utxo is confidential and its blinded data
// (value, asset and relative blinders) have been revealed.
func (u *Utxo) IsRevealed() bool {
	return len(u.ValueBlinder) > 0 && len(u.AssetBlinder) > 0
}

// IsLocked returns whether the utxo is locked.
func (u *Utxo) IsLocked() bool {
	return u.LockTimestamp > 0
}

// CanUnlock reutrns whether a locked utxo can be unlocked.
func (u *Utxo) CanUnlock() bool {
	if !u.IsLocked() {
		return true
	}
	return time.Now().After(time.Unix(u.LockTimestamp, 0))
}

// Key returns the UtxoKey of the current utxo.
func (u *Utxo) Key() UtxoKey {
	return u.UtxoKey
}

// Info returns a light view of the current utxo.
func (u *Utxo) Info() UtxoInfo {
	return UtxoInfo{
		u.Key(), u.Value, u.Asset, u.Script, u.ValueBlinder, u.AssetBlinder,
		u.AccountName,
	}
}

// Spend marks the utxos as spent.
func (u *Utxo) Spend() {
	u.Spent = true
	u.LockTimestamp = 0
}

// Confirm marks the utxos as confirmed.
func (u *Utxo) Confirm() {
	u.Confirmed = true
}

// Lock marks the current utxo as locked.
func (u *Utxo) Lock(timestamp int64) {
	if !u.IsLocked() {
		u.LockTimestamp = timestamp
	}
}

// Unlock marks the current locked utxo as unlocked.
func (u *Utxo) Unlock() {
	u.LockTimestamp = 0
}
