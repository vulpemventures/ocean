package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (i *InMemoryDbTestSuite) TestUtxoRepository() {
	dbtest.TestUtxoRepository(
		i.T(),
		ctx,
		inMemoryRepoManager.WalletRepository(),
		inMemoryRepoManager.UtxoRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
