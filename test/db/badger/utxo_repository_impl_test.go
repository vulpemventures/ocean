package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (b *BadgerDbTestSuite) TestUtxoRepository() {
	dbtest.TestUtxoRepository(
		b.T(),
		ctx,
		badgerRepoManager.WalletRepository(),
		badgerRepoManager.UtxoRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
