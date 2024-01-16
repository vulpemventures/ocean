CREATE TABLE external_script (
    account varchar(32) NOT NULL PRIMARY KEY,
    script varchar(255) NOT NULL,
    blinding_key bytea
);