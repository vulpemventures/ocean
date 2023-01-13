ALTER TABLE account_script_info DROP COLUMN fk_account_namespace;

ALTER TABLE tx_input_account DROP COLUMN fk_account_namespace;

ALTER TABLE utxo DROP COLUMN fk_account_namespace;

ALTER TABLE account DROP COLUMN namespace;

ALTER TABLE account DROP COLUMN label;

ALTER TABLE account ADD COLUMN name VARCHAR(50) PRIMARY KEY;

ALTER TABLE account_script_info ADD COLUMN fk_account_name VARCHAR(50);

ALTER TABLE account_script_info ADD FOREIGN KEY (fk_account_name) REFERENCES account(name);

ALTER TABLE tx_input_account ADD COLUMN account_name VARCHAR(50) NOT NULL;

ALTER TABLE utxo ADD COLUMN account_name VARCHAR(50) NOT NULL;