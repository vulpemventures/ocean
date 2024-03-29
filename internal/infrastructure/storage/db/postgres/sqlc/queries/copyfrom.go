// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: copyfrom.go

package queries

import (
	"context"
)

// iteratorForInsertAccountScripts implements pgx.CopyFromSource.
type iteratorForInsertAccountScripts struct {
	rows                 []InsertAccountScriptsParams
	skippedFirstNextCall bool
}

func (r *iteratorForInsertAccountScripts) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForInsertAccountScripts) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].Script,
		r.rows[0].DerivationPath,
		r.rows[0].FkAccountName,
	}, nil
}

func (r iteratorForInsertAccountScripts) Err() error {
	return nil
}

func (q *Queries) InsertAccountScripts(ctx context.Context, arg []InsertAccountScriptsParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"account_script_info"}, []string{"script", "derivation_path", "fk_account_name"}, &iteratorForInsertAccountScripts{rows: arg})
}
