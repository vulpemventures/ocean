package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (i *InMemoryDbTestSuite) TestTransactionRepository() {
	dbtest.TestTransactionRepository(
		i.T(),
		ctx,
		inMemoryRepoManager.WalletRepository(),
		inMemoryRepoManager.TransactionRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
