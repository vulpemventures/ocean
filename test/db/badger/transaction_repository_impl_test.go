package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (b *BadgerDbTestSuite) TestTransactionRepository() {
	dbtest.TestTransactionRepository(
		b.T(),
		ctx,
		badgerRepoManager.WalletRepository(),
		badgerRepoManager.TransactionRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
