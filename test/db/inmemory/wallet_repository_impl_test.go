package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (i *InMemoryDbTestSuite) TestWalletRepository() {
	dbtest.TestWalletRepository(
		i.T(),
		ctx,
		inMemoryRepoManager.WalletRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
