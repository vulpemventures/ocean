CREATE TABLE wallet (
    id varchar(255) NOT NULL PRIMARY KEY,
    encrypted_mnemonic VARCHAR(64) NOT NULL,
    password_hash VARCHAR(64) NOT NULL,
    birthday_block_height INTEGER NOT NULL,
    root_path VARCHAR(64) NOT NULL,
    network_name VARCHAR(64) NOT NULL,
    next_account_index INTEGER NOT NULL
);

CREATE TABLE account (
    name VARCHAR(50) NOT NULL PRIMARY KEY,
    index INTEGER NOT NULL,
    xpub VARCHAR(100) NOT NULL,
    derivation_path VARCHAR(100) NOT NULL,
    fk_wallet_id VARCHAR(255) NOT NULL,
    FOREIGN KEY (fk_wallet_id) REFERENCES wallet(id)
);

CREATE TABLE account_script_info (
    id SERIAL PRIMARY KEY,
    script bytea NOT NULL,
    derivation_path VARCHAR(100) NOT NULL,
    fk_account_name VARCHAR(50) NOT NULL,
    FOREIGN KEY (fk_account_name) REFERENCES account(name)
);

CREATE TABLE transaction (
    tx_id VARCHAR(64) NOT NULL PRIMARY KEY,
    tx_hex VARCHAR(1000) NOT NULL,
    block_hash VARCHAR(64) NOT NULL,
    block_height INTEGER NOT NULL
);

CREATE TABLE tx_input_account (
    id SERIAL PRIMARY KEY,
    account_name VARCHAR(50) NOT NULL,
    tx_id VARCHAR(64) NOT NULL,
    FOREIGN KEY (tx_id) REFERENCES transaction(tx_id)
);

CREATE TABLE utxo (
    id SERIAL PRIMARY KEY,
    tx_id VARCHAR(64) NOT NULL,
    vout INTEGER NOT NULL,
    value INTEGER NOT NULL,
    asset VARCHAR(64) NOT NULL,
    value_commitment bytea NOT NULL,
    asset_commitment bytea NOT NULL,
    value_blinder bytea NOT NULL,
    asset_blinder bytea NOT NULL,
    script bytea NOT NULL,
    nonce bytea NOT NULL,
    range_proof bytea NOT NULL,
    surjection_proof bytea NOT NULL,
    account_name varchar(50) NOT NULL,
    lock_timestamp timestamp NOT NULL
);

CREATE TABLE utxo_status (
    id SERIAL PRIMARY KEY,
    block_height INTEGER NOT NULL,
    block_time timestamp NOT NULL,
    block_hash varchar(64) NOT NULL,
    status integer NOT NULL,
    fk_utxo_id varchar(64) NOT NULL,
    FOREIGN KEY (fk_utxo_id) REFERENCES utxo(id)
);