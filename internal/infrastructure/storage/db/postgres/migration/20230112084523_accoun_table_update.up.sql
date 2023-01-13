ALTER TABLE account_script_info DROP COLUMN fk_account_name;

ALTER TABLE account DROP COLUMN name;

ALTER TABLE account ADD COLUMN namespace varchar(500) PRIMARY KEY;

ALTER TABLE account ADD COLUMN label varchar(500);

ALTER TABLE account_script_info ADD COLUMN fk_account_namespace varchar(500);

ALTER TABLE account_script_info ADD FOREIGN KEY (fk_account_namespace) REFERENCES account(namespace);

ALTER TABLE tx_input_account DROP COLUMN account_name;

ALTER TABLE tx_input_account ADD COLUMN fk_account_namespace varchar(500);

ALTER TABLE tx_input_account ADD FOREIGN KEY (fk_account_namespace) REFERENCES account(namespace);

ALTER TABLE utxo DROP COLUMN account_name;

ALTER TABLE utxo ADD COLUMN fk_account_namespace varchar(500);

ALTER TABLE utxo ADD FOREIGN KEY (fk_account_namespace) REFERENCES account(namespace);
