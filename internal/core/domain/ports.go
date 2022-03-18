package domain

// IMnemonicStore defines the methods a store storing a mnemonic in plaintext
// must implement to either set, unset or get it.
type IMnemonicStore interface {
	Set(mnemonic string)
	Unset()
	IsSet() bool
	Get() []string
}

// IMnemonicCipher defines the methods a cypher must implement to encrypt or
// decrypt a mnemonic with a password.
type IMnemonicCypher interface {
	Encrypt(mnemonic, password []byte) ([]byte, error)
	Decrypt(encryptedMnemonic, password []byte) ([]byte, error)
}

var MnemonicStore IMnemonicStore
var MnemonicCypher IMnemonicCypher
