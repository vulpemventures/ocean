package pgtest

import (
	pb "github.com/vulpemventures/ocean/api-spec/protobuf/gen/go/ocean/v1"
	"github.com/vulpemventures/ocean/test/testutil"
	"io"
	"time"
)

var (
	address = "localhost:18000"
)

type UtxoEvent struct {
	EvenType   string
	Err        error
	ConnClosed bool
}

func (g *GrpcDbTestSuite) TestOceanSanity() {
	walletClient, cancel, err := testutil.GetWalletClient(address, true, "")
	g.NoError(err)
	defer cancel()

	_, err = walletClient.CreateWallet(ctx, &pb.CreateWalletRequest{
		Mnemonic: mnemonic,
		Password: password,
	})
	g.NoError(err)

	_, err = walletClient.Unlock(ctx, &pb.UnlockRequest{
		Password: password,
	})
	g.NoError(err)

	accountClient, cancel, err := testutil.GetAccountClient(address, true, "")
	g.NoError(err)
	defer cancel()

	accountInfo, err := accountClient.CreateAccountBIP44(ctx, &pb.CreateAccountBIP44Request{
		Label:      "myAccount",
		ExtraXpubs: nil,
	})
	g.NoError(err)
	g.Equal("bip84-account0", accountInfo.GetAccountInfo().GetNamespace())
	g.Equal("myAccount", accountInfo.GetAccountInfo().GetLabel())

	deriveAddressResp, err := accountClient.DeriveAddresses(ctx, &pb.DeriveAddressesRequest{
		Name:           accountInfo.GetAccountInfo().GetNamespace(),
		NumOfAddresses: 1,
	})
	g.NoError(err)
	g.Equal(1, len(deriveAddressResp.GetAddresses()))

	notificationClient, cancel, err := testutil.GetNotificationsClient(
		address,
		true, "",
	)
	g.NoError(err)
	defer cancel()

	stream, err := notificationClient.UtxosNotifications(ctx, &pb.UtxosNotificationsRequest{})
	g.NoError(err)

	utxoEvent := make(chan UtxoEvent)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				utxoEvent <- UtxoEvent{
					ConnClosed: true,
				}

				return
			}
			if err != nil {
				utxoEvent <- UtxoEvent{
					Err: err,
				}
			}

			if len(resp.GetUtxos()) > 0 {
				utxoEvent <- UtxoEvent{
					EvenType: resp.GetEventType().String(),
				}

				time.Sleep(1 * time.Second)

				if err := stream.CloseSend(); err != nil {
					utxoEvent <- UtxoEvent{
						Err: err,
					}
				}

				return
			}
		}
	}()

	addr := deriveAddressResp.GetAddresses()[0]
	txId, err := testutil.Faucet(addr)
	g.NoError(err)

	utxoEventReceived := false
loop:
	for {
		select {
		case <-time.After(5 * time.Second):
			break loop
		case e := <-utxoEvent:
			if e.ConnClosed {
				break loop
			}
			if e.Err != nil {
				g.T().Log(e.Err.Error())
			}
			if e.EvenType != "" {
				g.T().Logf("receivet utxo event: %v", e.EvenType)
				utxoEventReceived = true
				break loop
			}
		}
	}

	if !utxoEventReceived {
		g.FailNow("utxo event not received")
	}

	time.Sleep(3 * time.Second)

	txClient, cancel, err := testutil.GetTransactionClient(address, true, "")
	g.NoError(err)
	defer cancel()

	tx, err := txClient.GetTransaction(ctx, &pb.GetTransactionRequest{
		Txid: txId,
	})
	g.NoError(err)
	g.NotNil(tx.GetBlockDetails())
}
