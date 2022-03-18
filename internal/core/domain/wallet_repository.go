package domain

import (
	"context"
)

const (
	WalletCreated WalletEventType = iota
	WalletUnlocked
	WalletPasswordChanged
	WalletAccountCreated
	WalletAccountAddressesDerived
	WalletAccountDeleted
)

var (
	walletTypeString = map[WalletEventType]string{
		WalletCreated:                 "WalletCreated",
		WalletUnlocked:                "WalletUnlocked",
		WalletPasswordChanged:         "WalletPasswordChanged",
		WalletAccountCreated:          "WalletAccountCreated",
		WalletAccountAddressesDerived: "WalletAccountAddressesDerived",
		WalletAccountDeleted:          "WalletAccountDeleted",
	}
)

type WalletEventType int

func (t WalletEventType) String() string {
	return walletTypeString[t]
}

// WalletEvent holds info about an event occured within the repository.
type WalletEvent struct {
	EventType        WalletEventType
	AccountName      string
	AccountAddresses []AddressInfo
}

// WalletRepository is the abstraction for any kind of database intended to
// persist a Wallet.
type WalletRepository interface {
	// CreateWallet stores a new Wallet if not yet existing.
	// Generates a WalletCreated event if successfull.
	CreateWallet(ctx context.Context, wallet *Wallet) error
	// GetWallet returns the stored wallet, if existing.
	GetWallet(ctx context.Context) (*Wallet, error)
	// UnlockWallet attempts to update the status of the Wallet to "unlocked".
	// Generates a WalletUnlocked event if successfull.
	UnlockWallet(ctx context.Context, password string) error
	// UpdateWallet allows to make multiple changes to the Wallet in a
	// transactional way.
	UpdateWallet(
		ctx context.Context, updateFn func(v *Wallet) (*Wallet, error),
	) error
	// CreateAccount creates a new wallet account with the given name and returns
	// its basic info.
	// Generates a WalletAccountCreated event if successfull.
	CreateAccount(ctx context.Context, accountName string) (*AccountInfo, error)
	// DeriveNextExternalAddressesForAccount returns one or more new receiving
	// addresses for the given account.
	// Generates a WalletAccountAddressesDerived event if successfull.
	DeriveNextExternalAddressesForAccount(
		ctx context.Context, accountName string, numOfAddresses uint64,
	) ([]AddressInfo, error)
	// DeriveNextInternalAddressesForAccount returns one or more new change
	// addresses for the given account.
	// Generates a WalletAccountAddressesDerived event if successfull.
	DeriveNextInternalAddressesForAccount(
		ctx context.Context, accountName string, numOfAddresses uint64,
	) ([]AddressInfo, error)
	// DeleteAccount deletes the wallet account with the given name.
	// Generates a WalletAccountDeleted event if successfull.
	DeleteAccount(ctx context.Context, accountName string) error
	// GetEventChannel returns the channel of WalletEvents.
	GetEventChannel() chan WalletEvent
}
