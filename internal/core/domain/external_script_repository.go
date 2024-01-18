package domain

import "context"

const (
	ExternalScriptAdded ExternalScriptEventType = iota
	ExternalScriptDeleted
)

var (
	externalScriptTypeString = map[ExternalScriptEventType]string{
		ExternalScriptAdded:   "ExternalScriptAdded",
		ExternalScriptDeleted: "ExternalScriptDeleted",
	}
)

type ExternalScriptEventType int

func (t ExternalScriptEventType) String() string {
	return externalScriptTypeString[t]
}

// ExternalScriptEvent holds info about an event occured within the repository.
type ExternalScriptEvent struct {
	EventType ExternalScriptEventType
	Info      AddressInfo
}

// ExternalScriptRepository is the abstraction for any kind of database intended to
// persist external scripts as AddressInfo.
type ExternalScriptRepository interface {
	// AddScripts persists the given external script by preventing duplicates.
	AddScript(ctx context.Context, script AddressInfo) (bool, error)
	// GetAllScripts returns all the persisted external scripts.
	GetAllScripts(ctx context.Context) ([]AddressInfo, error)
	// DeleteScript removes an external script idenitified by its hash from the store.
	DeleteScript(ctx context.Context, scriptHash string) (bool, error)
}
