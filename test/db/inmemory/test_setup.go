package pgtest

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vulpemventures/go-elements/network"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/core/ports"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/inmemory"
	"github.com/vulpemventures/ocean/test/testutil"
	"strings"
)

var (
	mnemonic = []string{
		"leave", "dice", "fine", "decrease", "dune", "ribbon", "ocean", "earn",
		"lunar", "account", "silver", "admit", "cheap", "fringe", "disorder", "trade",
		"because", "trade", "steak", "clock", "grace", "video", "jacket", "equal",
	}
	encryptedMnemonic = "8f29524ee5995c838ca6f28c7ded7da6dc51de804fd2703775989e65ddc1bb3b60122bf0f430bb3b7a267449aaeee103375737d679bfdabf172c3842048925e6f8952e214f6b900435d24cff938be78ad3bb303d305702fbf168534a45a57ac98ca940d4c3319f14d0c97a20b5bcb456d72857d48d0b4f0e0dcf71d1965b6a42aca8d84fcb66aadeabc812a9994cf66e7a75f8718a031418468f023c560312a02f46ec8e65d5dd65c968ddb93e10950e96c8e730ce7a74d33c6ddad9e12f45e534879f1605eb07fe90432f6592f7996091bbb3e3b2"
	password          = "password"
	newPassword       = "newPassword"
	ctx               = context.Background()
	rootPath          = "m/84'/1'"
	regtest           = network.Regtest.Name
	birthdayBlock     = uint32(1)

	inMemoryRepoManager ports.RepoManager
)

type InMemoryDbTestSuite struct {
	suite.Suite
}

func (i *InMemoryDbTestSuite) SetupSuite() {
	mockedMnemonicCypher := &testutil.MockMnemonicCypher{}
	mockedMnemonicCypher.On("Encrypt", mock.Anything, mock.Anything).Return(testutil.H2b(encryptedMnemonic), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, []byte(password)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, []byte(newPassword)).Return([]byte(strings.Join(mnemonic, " ")), nil)
	mockedMnemonicCypher.On("Decrypt", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("invalid password"))
	domain.MnemonicCypher = mockedMnemonicCypher
	domain.MnemonicStore = testutil.NewInMemoryMnemonicStore()
}

func (i *InMemoryDbTestSuite) TearDownSuite() {
}

func (i *InMemoryDbTestSuite) BeforeTest(suiteName, testName string) {
	domain.MnemonicStore = testutil.NewInMemoryMnemonicStore()
	inMemoryRepoManager = inmemory.NewRepoManager()
	handler := testutil.HandlerFactory(i.T(), "inmemory")
	inMemoryRepoManager.RegisterHandlerForWalletEvent(domain.WalletCreated, handler)
	inMemoryRepoManager.RegisterHandlerForWalletEvent(domain.WalletUnlocked, handler)
	inMemoryRepoManager.RegisterHandlerForWalletEvent(domain.WalletAccountCreated, handler)
	inMemoryRepoManager.RegisterHandlerForWalletEvent(domain.WalletAccountAddressesDerived, handler)
	inMemoryRepoManager.RegisterHandlerForWalletEvent(domain.WalletAccountDeleted, handler)
}

func (i *InMemoryDbTestSuite) AfterTest(suiteName, testName string) {
}
