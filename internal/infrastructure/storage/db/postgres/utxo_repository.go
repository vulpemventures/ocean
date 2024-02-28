package postgresdb

import (
	"bytes"
	"context"
	"database/sql"
	"sync"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vulpemventures/ocean/internal/core/domain"
	"github.com/vulpemventures/ocean/internal/infrastructure/storage/db/postgres/sqlc/queries"
)

const (
	utxoSpent = iota
	utxoConfirmed
)

type utxoRepositoryPg struct {
	pgxPool          *pgxpool.Pool
	querier          *queries.Queries
	chLock           *sync.Mutex
	chEvents         chan domain.UtxoEvent
	externalChEvents chan domain.UtxoEvent
}

func NewUtxoRepositoryPgImpl(pgxPool *pgxpool.Pool) domain.UtxoRepository {
	return newUtxoRepositoryPgImpl(pgxPool)
}

func newUtxoRepositoryPgImpl(pgxPool *pgxpool.Pool) *utxoRepositoryPg {
	return &utxoRepositoryPg{
		pgxPool:          pgxPool,
		querier:          queries.New(pgxPool),
		chLock:           &sync.Mutex{},
		chEvents:         make(chan domain.UtxoEvent),
		externalChEvents: make(chan domain.UtxoEvent),
	}
}

func (u *utxoRepositoryPg) AddUtxos(
	ctx context.Context, utxos []*domain.Utxo,
) (int, error) {
	count := 0

	utxosInfo := make([]domain.UtxoInfo, 0, len(utxos))
	conn, err := u.pgxPool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()
	for _, v := range utxos {
		tx, err := conn.Begin(ctx)
		if err != nil {
			return 0, err
		}

		querierWithTx := u.querier.WithTx(tx)

		req := queries.InsertUtxoParams{
			TxID:                v.TxID,
			Vout:                int32(v.VOut),
			Value:               int64(v.Value),
			Asset:               v.Asset,
			ValueCommitment:     v.ValueCommitment,
			AssetCommitment:     v.AssetCommitment,
			ValueBlinder:        v.ValueBlinder,
			AssetBlinder:        v.AssetBlinder,
			Script:              v.Script,
			Nonce:               v.Nonce,
			RangeProof:          v.RangeProof,
			SurjectionProof:     v.SurjectionProof,
			AccountName:         v.AccountName,
			LockTimestamp:       v.LockTimestamp,
			LockExpiryTimestamp: v.LockExpiryTimestamp,
		}
		utxo, err := querierWithTx.InsertUtxo(ctx, req)
		if err != nil {
			tx.Rollback(ctx)
			if pqErr, ok := err.(*pgconn.PgError); pqErr != nil && ok && pqErr.Code == uniqueViolation {
				continue
			} else {
				return 0, err
			}
		}

		if v.IsSpent() {
			if _, err := querierWithTx.InsertUtxoStatus(ctx, queries.InsertUtxoStatusParams{
				BlockHeight: int32(v.SpentStatus.BlockHeight),
				BlockTime:   v.SpentStatus.BlockTime,
				BlockHash:   v.SpentStatus.BlockHash,
				Status:      utxoSpent,
				FkUtxoID:    utxo.ID,
			}); err != nil {
				tx.Rollback(ctx)
				return 0, err
			}
		}

		if v.IsConfirmed() {
			if _, err := querierWithTx.InsertUtxoStatus(ctx, queries.InsertUtxoStatusParams{
				BlockHeight: int32(v.ConfirmedStatus.BlockHeight),
				BlockTime:   v.ConfirmedStatus.BlockTime,
				BlockHash:   v.ConfirmedStatus.BlockHash,
				Status:      utxoConfirmed,
				FkUtxoID:    utxo.ID,
			}); err != nil {
				tx.Rollback(ctx)
				return 0, err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}

		utxosInfo = append(utxosInfo, v.Info())
		count++
	}

	if len(utxosInfo) > 0 {
		go u.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoAdded,
			Utxos:     utxosInfo,
		})

	}

	return count, nil
}

func (u *utxoRepositoryPg) GetUtxosByKey(
	ctx context.Context,
	utxoKeys []domain.UtxoKey,
) ([]*domain.Utxo, error) {
	utxos := make([]*domain.Utxo, 0, len(utxoKeys))
	for _, key := range utxoKeys {
		utxo, err := u.querier.GetUtxoForKey(ctx, queries.GetUtxoForKeyParams{
			TxID: key.TxID,
			Vout: int32(key.VOut),
		})
		if err != nil {
			return nil, err
		}

		if len(utxo) == 0 {
			continue
		}

		ut := &domain.Utxo{
			UtxoKey: domain.UtxoKey{
				TxID: utxo[0].TxID,
				VOut: uint32(utxo[0].Vout),
			},
			Value:               uint64(utxo[0].Value),
			Asset:               utxo[0].Asset,
			ValueCommitment:     utxo[0].ValueCommitment,
			AssetCommitment:     utxo[0].AssetCommitment,
			ValueBlinder:        utxo[0].ValueBlinder,
			AssetBlinder:        utxo[0].AssetBlinder,
			Script:              utxo[0].Script,
			Nonce:               utxo[0].Nonce,
			RangeProof:          utxo[0].RangeProof,
			SurjectionProof:     utxo[0].SurjectionProof,
			AccountName:         utxo[0].AccountName,
			LockTimestamp:       utxo[0].LockTimestamp,
			LockExpiryTimestamp: utxo[0].LockExpiryTimestamp,
		}

		for _, v := range utxo {
			if v.Status.Valid {
				switch v.Status.Int32 {
				case utxoSpent:
					ut.SpentStatus = domain.UtxoStatus{
						Txid:        v.TxID,
						BlockHeight: uint64(v.BlockHeight.Int32),
						BlockTime:   v.BlockTime.Int64,
						BlockHash:   v.BlockHash.String,
					}
				case utxoConfirmed:
					ut.ConfirmedStatus = domain.UtxoStatus{
						BlockHeight: uint64(v.BlockHeight.Int32),
						BlockTime:   v.BlockTime.Int64,
						BlockHash:   v.BlockHash.String,
					}
				}
			}
		}

		utxos = append(utxos, ut)
	}

	return utxos, nil
}

func (u *utxoRepositoryPg) GetAllUtxos(
	ctx context.Context,
) ([]*domain.Utxo, error) {
	resp := make([]*domain.Utxo, 0)
	utxos, err := u.querier.GetAllUtxos(ctx)
	if err != nil {
		return nil, nil
	}

	utxosByKey, err := u.convertToUtxos(utxos)
	if err != nil {
		return nil, err
	}

	for _, v := range utxosByKey {
		resp = append(resp, v)
	}

	return resp, nil
}

func (u *utxoRepositoryPg) GetSpendableUtxos(
	ctx context.Context,
) ([]*domain.Utxo, error) {
	resp := make([]*domain.Utxo, 0)
	utxos, err := u.querier.GetAllUtxos(ctx)
	if err != nil {
		return nil, nil
	}

	utxosByKey, err := u.convertToUtxos(utxos)
	if err != nil {
		return nil, nil
	}

	for _, v := range utxosByKey {
		if !v.IsLocked() && v.IsConfirmed() && !v.IsSpent() {
			resp = append(resp, v)
		}
	}

	return resp, nil
}

func (u *utxoRepositoryPg) GetAllUtxosForAccount(
	ctx context.Context, account string, scripts [][]byte,
) ([]*domain.Utxo, error) {
	resp := make([]*domain.Utxo, 0)
	utxos, err := u.querier.GetUtxosForAccount(ctx, account)
	if err != nil {
		return nil, nil
	}

	req := make([]queries.GetAllUtxosRow, 0, len(utxos))
	for _, v := range utxos {
		req = append(
			req,
			toGetAllUtxosRow(v),
		)

	}

	utxosByKey, err := u.convertToUtxos(req)
	if err != nil {
		return nil, nil
	}

	for _, v := range utxosByKey {
		found := len(scripts) <= 0
		for _, script := range scripts {
			if bytes.Equal(v.Script, script) {
				found = true
				break
			}
		}
		if found {
			resp = append(resp, v)
		}
	}

	return resp, nil
}

func (u *utxoRepositoryPg) GetSpendableUtxosForAccount(
	ctx context.Context, account string,
) ([]*domain.Utxo, error) {
	resp := make([]*domain.Utxo, 0)
	utxos, err := u.querier.GetUtxosForAccount(ctx, account)
	if err != nil {
		return nil, nil
	}

	req := make([]queries.GetAllUtxosRow, 0, len(utxos))
	for _, v := range utxos {
		req = append(
			req,
			toGetAllUtxosRow(v),
		)

	}

	utxosByKey, err := u.convertToUtxos(req)
	if err != nil {
		return nil, nil
	}

	for _, v := range utxosByKey {
		if !v.IsLocked() && v.IsConfirmed() && !v.IsSpent() {
			resp = append(resp, v)
		}
	}

	return resp, nil
}

func (u *utxoRepositoryPg) GetLockedUtxosForAccount(
	ctx context.Context, account string,
) ([]*domain.Utxo, error) {
	resp := make([]*domain.Utxo, 0)
	utxos, err := u.querier.GetUtxosForAccount(ctx, account)
	if err != nil {
		return nil, nil
	}

	req := make([]queries.GetAllUtxosRow, 0, len(utxos))
	for _, v := range utxos {
		req = append(
			req,
			toGetAllUtxosRow(v),
		)

	}

	utxosByKey, err := u.convertToUtxos(req)
	if err != nil {
		return nil, nil
	}

	for _, v := range utxosByKey {
		if v.IsLocked() {
			resp = append(resp, v)
		}
	}

	return resp, nil
}

func (u *utxoRepositoryPg) GetBalanceForAccount(
	ctx context.Context, account string,
) (map[string]*domain.Balance, error) {
	resp := make(map[string]*domain.Balance)
	utxos, err := u.querier.GetUtxosForAccount(ctx, account)
	if err != nil {
		return nil, nil
	}

	req := make([]queries.GetAllUtxosRow, 0, len(utxos))
	for _, v := range utxos {
		req = append(
			req,
			toGetAllUtxosRow(v),
		)

	}

	utxosByKey, err := u.convertToUtxos(req)
	if err != nil {
		return nil, nil
	}

	for _, v := range utxosByKey {
		if v.IsSpent() {
			continue
		}

		if _, ok := resp[v.Asset]; !ok {
			resp[v.Asset] = &domain.Balance{}
		}

		b := resp[v.Asset]
		if v.IsLocked() {
			b.Locked += v.Value
		} else {
			if v.IsConfirmed() {
				b.Confirmed += v.Value
			} else {
				b.Unconfirmed += v.Value
			}
		}
	}

	return resp, nil
}

func (u *utxoRepositoryPg) SpendUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, txid string,
) (int, error) {
	return u.spendUtxos(ctx, utxoKeys, txid)
}

func (u *utxoRepositoryPg) ConfirmSpendUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	return u.confirmSpendUtxos(ctx, utxoKeys, status)
}

func (u *utxoRepositoryPg) ConfirmUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	return u.confirmUtxos(ctx, utxoKeys, status)
}

func (u *utxoRepositoryPg) LockUtxos(
	ctx context.Context,
	utxoKeys []domain.UtxoKey, timestamp, expiryTimestamp int64,
) (int, error) {
	return u.lockUtxos(ctx, utxoKeys, timestamp, expiryTimestamp)
}

func (u *utxoRepositoryPg) UnlockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	return u.unlockUtxos(ctx, utxoKeys)
}

func (u *utxoRepositoryPg) DeleteUtxosForAccount(
	ctx context.Context, accountName string,
) error {
	conn, err := u.pgxPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	// Rollback rolls back the transaction. Rollback will return ErrTxClosed if the
	// Tx is already closed, but is otherwise safe to call multiple times. Hence, a
	// defer tx.Rollback() is safe even if tx.Commit() will be called first in a
	// non-error condition.
	defer tx.Rollback(ctx)

	querierWithTx := u.querier.WithTx(tx)

	utxos, err := querierWithTx.GetUtxosForAccount(ctx, accountName)
	if err != nil {
		return err
	}

	for _, v := range utxos {
		err = querierWithTx.DeleteUtxoStatuses(ctx, v.ID)
		if err != nil {
			return err
		}
	}

	if err := querierWithTx.DeleteUtxosForAccountName(ctx, accountName); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (u *utxoRepositoryPg) GetEventChannel() chan domain.UtxoEvent {
	return u.externalChEvents
}

func (u *utxoRepositoryPg) publishEvent(event domain.UtxoEvent) {
	u.chLock.Lock()
	defer u.chLock.Unlock()

	u.chEvents <- event
	// send over channel without blocking in case nobody is listening.
	select {
	case u.externalChEvents <- event:
	default:
	}
}

func (u *utxoRepositoryPg) close() {
	close(u.chEvents)
	close(u.externalChEvents)
}

func (u *utxoRepositoryPg) spendUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, txid string,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := u.spendUtxo(ctx, key, txid)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}
	if count > 0 {
		go u.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoSpent,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (u *utxoRepositoryPg) spendUtxo(
	ctx context.Context, key domain.UtxoKey, txid string,
) (bool, *domain.UtxoInfo, error) {
	utxos, err := u.GetUtxosByKey(ctx, []domain.UtxoKey{key})
	if err != nil {
		return false, nil, err
	}

	if len(utxos) <= 0 {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsSpent() {
		return false, nil, nil
	}

	if err := utxo.Spend(txid); err != nil {
		return false, nil, err
	}

	if err := u.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (u *utxoRepositoryPg) confirmSpendUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := u.confirmSpendUtxo(ctx, key, status)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}
	if count > 0 {
		go u.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoConfirmedSpend,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (u *utxoRepositoryPg) confirmSpendUtxo(
	ctx context.Context, key domain.UtxoKey, status domain.UtxoStatus,
) (bool, *domain.UtxoInfo, error) {
	utxos, err := u.GetUtxosByKey(ctx, []domain.UtxoKey{key})
	if err != nil {
		return false, nil, err
	}

	if len(utxos) <= 0 {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsConfirmedSpent() {
		return false, nil, nil
	}

	if err := utxo.ConfirmSpend(status); err != nil {
		return false, nil, err
	}

	if err := u.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (u *utxoRepositoryPg) updateUtxo(
	ctx context.Context, utxo *domain.Utxo,
) error {
	ut, err := u.querier.UpdateUtxo(ctx, queries.UpdateUtxoParams{
		Value:               int64(utxo.Value),
		Asset:               utxo.Asset,
		ValueCommitment:     utxo.ValueCommitment,
		AssetCommitment:     utxo.AssetCommitment,
		ValueBlinder:        utxo.ValueBlinder,
		AssetBlinder:        utxo.AssetBlinder,
		Script:              utxo.Script,
		Nonce:               utxo.Nonce,
		RangeProof:          utxo.RangeProof,
		SurjectionProof:     utxo.SurjectionProof,
		AccountName:         utxo.AccountName,
		LockTimestamp:       utxo.LockTimestamp,
		LockExpiryTimestamp: utxo.LockExpiryTimestamp,
		TxID:                utxo.TxID,
		Vout:                int32(utxo.VOut),
	})
	if err != nil {
		return err
	}

	if err := u.querier.DeleteUtxoStatuses(ctx, ut.ID); err != nil {
		return err
	}

	if utxo.IsSpent() {
		if _, err := u.querier.InsertUtxoStatus(ctx, queries.InsertUtxoStatusParams{
			BlockHeight: int32(utxo.SpentStatus.BlockHeight),
			BlockTime:   utxo.SpentStatus.BlockTime,
			BlockHash:   utxo.SpentStatus.BlockHash,
			Status:      utxoSpent,
			FkUtxoID:    ut.ID,
			TxID: sql.NullString{
				String: utxo.SpentStatus.Txid,
				Valid:  true,
			},
		}); err != nil {
			return err
		}
	}
	if utxo.IsConfirmed() {
		if _, err := u.querier.InsertUtxoStatus(ctx, queries.InsertUtxoStatusParams{
			BlockHeight: int32(utxo.ConfirmedStatus.BlockHeight),
			BlockTime:   utxo.ConfirmedStatus.BlockTime,
			BlockHash:   utxo.ConfirmedStatus.BlockHash,
			Status:      utxoConfirmed,
			FkUtxoID:    ut.ID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (u *utxoRepositoryPg) confirmUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, status domain.UtxoStatus,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := u.confirmUtxo(ctx, key, status)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}

	if count > 0 {
		go u.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoConfirmed,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (u *utxoRepositoryPg) confirmUtxo(
	ctx context.Context, key domain.UtxoKey, status domain.UtxoStatus,
) (bool, *domain.UtxoInfo, error) {
	utxos, err := u.GetUtxosByKey(ctx, []domain.UtxoKey{key})
	if err != nil {
		return false, nil, err
	}

	if len(utxos) <= 0 {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsConfirmed() {
		return false, nil, nil
	}

	if err := utxo.Confirm(status); err != nil {
		return false, nil, err
	}
	if err := u.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (u *utxoRepositoryPg) lockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey, timestamp, expiryTimestamp int64,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := u.lockUtxo(ctx, key, timestamp, expiryTimestamp)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}

	if count > 0 {
		go u.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoLocked,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (u *utxoRepositoryPg) lockUtxo(
	ctx context.Context, key domain.UtxoKey, timestamp, expiryTimestamp int64,
) (bool, *domain.UtxoInfo, error) {
	utxos, err := u.GetUtxosByKey(ctx, []domain.UtxoKey{key})
	if err != nil {
		return false, nil, err
	}

	if len(utxos) <= 0 {
		return false, nil, nil
	}

	utxo := utxos[0]
	if utxo.IsLocked() {
		return false, nil, nil
	}

	utxo.Lock(timestamp, expiryTimestamp)
	if err := u.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (u *utxoRepositoryPg) unlockUtxos(
	ctx context.Context, utxoKeys []domain.UtxoKey,
) (int, error) {
	count := 0
	utxosInfo := make([]domain.UtxoInfo, 0)
	for _, key := range utxoKeys {
		done, info, err := u.unlockUtxo(ctx, key)
		if err != nil {
			return -1, err
		}
		if done {
			count++
			utxosInfo = append(utxosInfo, *info)
		}
	}

	if count > 0 {
		go u.publishEvent(domain.UtxoEvent{
			EventType: domain.UtxoUnlocked,
			Utxos:     utxosInfo,
		})
	}

	return count, nil
}

func (u *utxoRepositoryPg) unlockUtxo(
	ctx context.Context, key domain.UtxoKey,
) (bool, *domain.UtxoInfo, error) {
	utxos, err := u.GetUtxosByKey(ctx, []domain.UtxoKey{key})
	if err != nil {
		return false, nil, err
	}

	if len(utxos) <= 0 {
		return false, nil, nil
	}

	utxo := utxos[0]
	if !utxo.IsLocked() {
		return false, nil, nil
	}

	utxo.Unlock()
	if err := u.updateUtxo(ctx, utxo); err != nil {
		return false, nil, err
	}

	utxoInfo := utxo.Info()
	return true, &utxoInfo, nil
}

func (u *utxoRepositoryPg) convertToUtxos(
	utxos []queries.GetAllUtxosRow,
) (map[domain.UtxoKey]*domain.Utxo, error) {
	utxosByKey := make(map[domain.UtxoKey]*domain.Utxo)
	for _, v := range utxos {
		key := domain.UtxoKey{
			TxID: v.TxID,
			VOut: uint32(v.Vout),
		}

		utxo, ok := utxosByKey[key]
		if !ok {
			utxo = &domain.Utxo{
				UtxoKey:             key,
				Value:               uint64(v.Value),
				Asset:               v.Asset,
				ValueCommitment:     v.ValueCommitment,
				AssetCommitment:     v.AssetCommitment,
				ValueBlinder:        v.ValueBlinder,
				AssetBlinder:        v.AssetBlinder,
				Script:              v.Script,
				Nonce:               v.Nonce,
				RangeProof:          v.RangeProof,
				SurjectionProof:     v.SurjectionProof,
				AccountName:         v.AccountName,
				LockTimestamp:       v.LockTimestamp,
				LockExpiryTimestamp: v.LockExpiryTimestamp,
			}
			utxosByKey[key] = utxo
		}
		if v.Status.Valid {
			switch v.Status.Int32 {
			case utxoSpent:
				utxo.SpentStatus = domain.UtxoStatus{
					Txid:        v.TxID,
					BlockHeight: uint64(v.BlockHeight.Int32),
					BlockTime:   v.BlockTime.Int64,
					BlockHash:   v.BlockHash.String,
				}
			case utxoConfirmed:
				utxo.ConfirmedStatus = domain.UtxoStatus{
					BlockHeight: uint64(v.BlockHeight.Int32),
					BlockTime:   v.BlockTime.Int64,
					BlockHash:   v.BlockHash.String,
				}
			}
		}
	}

	return utxosByKey, nil
}

func (u *utxoRepositoryPg) reset(
	querier *queries.Queries, ctx context.Context,
) {
	querier.ResetUtxos(ctx)
}

func toGetAllUtxosRow(v queries.GetUtxosForAccountRow) queries.GetAllUtxosRow {
	return queries.GetAllUtxosRow{
		TxID:                v.TxID,
		Vout:                v.Vout,
		Value:               v.Value,
		Asset:               v.Asset,
		ValueCommitment:     v.ValueCommitment,
		AssetCommitment:     v.AssetCommitment,
		ValueBlinder:        v.ValueBlinder,
		AssetBlinder:        v.AssetBlinder,
		Script:              v.Script,
		Nonce:               v.Nonce,
		RangeProof:          v.RangeProof,
		SurjectionProof:     v.SurjectionProof,
		AccountName:         v.AccountName,
		LockTimestamp:       v.LockTimestamp,
		LockExpiryTimestamp: v.LockExpiryTimestamp,
		ID_2:                v.ID_2,
		BlockHeight:         v.BlockHeight,
		BlockTime:           v.BlockTime,
		BlockHash:           v.BlockHash,
		Status:              v.Status,
		FkUtxoID:            v.FkUtxoID,
	}
}
