package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (p *PgDbTestSuite) TestTransactionRepository() {
	dbtest.TestTransactionRepository(
		p.T(),
		ctx,
		pgRepoManager.WalletRepository(),
		pgRepoManager.TransactionRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
