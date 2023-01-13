package pgtest

import dbtest "github.com/vulpemventures/ocean/test/db"

func (p *PgDbTestSuite) TestUtxoRepository() {
	dbtest.TestUtxoRepository(
		p.T(),
		ctx,
		pgRepoManager.WalletRepository(),
		pgRepoManager.UtxoRepository(),
		mnemonic,
		password,
		newPassword,
		rootPath,
		regtest,
		birthdayBlock,
	)
}
