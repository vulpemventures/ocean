package ports

import "context"

type Cosigner interface {
	GetXpub(ctx context.Context) (string, error)
	RegisterMultiSig(ctx context.Context, descriptor string) error
	SignTx(ctx context.Context, tx string) (string, error)
}
