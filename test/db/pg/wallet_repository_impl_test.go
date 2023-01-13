package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (p *PgDbTestSuite) TestWalletRepository() {
	dbtest.TestWalletRepository(
		p.T(),
		ctx,
		pgRepoManager.WalletRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
