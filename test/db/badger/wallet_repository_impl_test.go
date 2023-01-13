package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (b *BadgerDbTestSuite) TestWalletRepository() {
	dbtest.TestWalletRepository(
		b.T(),
		ctx,
		badgerRepoManager.WalletRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
