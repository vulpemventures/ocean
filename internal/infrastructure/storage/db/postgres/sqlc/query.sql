/* WALLET & ACCOUNT */
-- name: InsertWallet :one
INSERT INTO wallet(id, encrypted_mnemonic,password_hash,birthday_block_height,root_path,network_name,next_account_index)
VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: GetWalletAccountsAndScripts :many
SELECT w.id as walletId,w.encrypted_mnemonic,w.password_hash,w.birthday_block_height,w.root_path,w.network_name,w.next_account_index, a.name,a.index,a.xpub,a.derivation_path as account_derivation_path,a.next_external_index,a.next_internal_index,a.fk_wallet_id,asi.script,asi.derivation_path as script_derivation_path,asi.fk_account_name FROM
wallet w LEFT JOIN account a ON w.id = a.fk_wallet_id
LEFT JOIN account_script_info asi on a.name = asi.fk_account_name
WHERE w.id = $1;

-- name: UpdateWallet :one
UPDATE wallet SET encrypted_mnemonic = $2, password_hash = $3, birthday_block_height = $4, root_path = $5, network_name = $6, next_account_index = $7 WHERE id = $1 RETURNING *;

-- name: GetAccount :one
SELECT * FROM account WHERE name = $1;

-- name: InsertAccount :one
INSERT INTO account(name,index,xpub,derivation_path,next_external_index,next_internal_index,fk_wallet_id)
VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: UpdateAccountIndexes :one
UPDATE account SET next_external_index = $1, next_internal_index = $2 WHERE name = $3 RETURNING *;

-- name: InsertAccountScripts :copyfrom
INSERT INTO account_script_info (script,derivation_path,fk_account_name) VALUES ($1, $2, $3);

-- name: DeleteAccountScripts :exec
DELETE FROM account_script_info WHERE fk_account_name = $1;

-- name: DeleteAccount :exec
DELETE FROM account WHERE name = $1;

/* UTXO */
-- name: InsertUtxo :one
INSERT INTO utxo(tx_id,vout,value,asset,value_commitment,asset_commitment,value_blinder,asset_blinder,script,nonce,range_proof,surjection_proof,account_name,lock_timestamp)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING *;

-- name: InsertUtxoStatus :one
INSERT INTO utxo_status(block_height,block_time,block_hash,status,fk_utxo_id)
VALUES($1,$2,$3,$4,$5) RETURNING *;

-- name: GetUtxoForKey :many
SELECT * FROM utxo u left join utxo_status us on u.id = us.fk_utxo_id
WHERE u.tx_id = $1 AND u.vout = $2;

-- name: GetAllUtxos :many
SELECT * FROM utxo u left join utxo_status us on u.id = us.fk_utxo_id;

-- name: GetUtxosForAccount :many
SELECT * FROM utxo u left join utxo_status us on u.id = us.fk_utxo_id
WHERE u.account_name = $1;

-- name: UpdateUtxo :one
UPDATE utxo SET value=$1,asset=$2,value_commitment=$3,asset_commitment=$4,value_blinder=$5,asset_blinder=$6,script=$7,nonce=$8,range_proof=$9,surjection_proof=$10,account_name=$11,lock_timestamp=$12 WHERE tx_id=$13 and vout=$14 RETURNING *;

-- name: DeleteUtxoStatuses :exec
DELETE FROM utxo_status WHERE fk_utxo_id = $1;

-- name: DeleteUtxosForAccountName :exec
DELETE FROM utxo WHERE account_name=$1;

-- name: GetUtxosForAccountName :many
SELECT * FROM utxo WHERE account_name=$1;