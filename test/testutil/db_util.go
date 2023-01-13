package testutil

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	DB     *sql.DB
	dbUser = "root"
	dbPass = "secret"
	dbHost = "127.0.0.1"
	dbPort = "5432"
	dbName = "oceand-db-test"
)

func SetupDB() error {
	db, err := createDBConnection()
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func ShutdownDB() error {
	if err := TruncateDB(); err != nil {
		return err
	}

	return DB.Close()
}

func TruncateDB() error {
	truncateQuery := `
          SELECT truncate_tables('%s')
 `
	formattedQuery := fmt.Sprintf(truncateQuery, dbUser)
	_, err := DB.ExecContext(context.Background(), formattedQuery)
	if err != nil {
		return err
	}

	return nil
}

func createDBConnection() (*sql.DB, error) {
	formattedURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("pgx", formattedURL)
	if err != nil {
		return nil, err
	}

	DB = db

	truncateFunctionQuery := `
	  CREATE OR REPLACE FUNCTION truncate_tables(username IN VARCHAR) RETURNS void AS $$
	  DECLARE
	      statements CURSOR FOR
		  SELECT tablename FROM pg_tables
		  WHERE tableowner = username AND schemaname = 'public' AND tablename NOT LIKE '%migrations';
	  BEGIN
	      FOR stmt IN statements LOOP
		  EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
	      END LOOP;
	  END;
	  $$ LANGUAGE plpgsql;
`

	_, err = db.ExecContext(context.Background(), truncateFunctionQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
}
