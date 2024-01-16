package inmemory

import (
	"context"
	"sync"

	"github.com/vulpemventures/ocean/internal/core/domain"
)

type scriptInmemoryStore struct {
	scripts map[string]domain.AddressInfo
	lock    *sync.RWMutex
}

type scriptRepository struct {
	store    *scriptInmemoryStore
	chEvents chan domain.ExternalScriptEvent
	chLock   *sync.Mutex
}

func NewExternalScriptRepository() domain.ExternalScriptRepository {
	return newExternalScriptRepository()
}

func newExternalScriptRepository() *scriptRepository {
	return &scriptRepository{
		store: &scriptInmemoryStore{
			scripts: make(map[string]domain.AddressInfo),
			lock:    &sync.RWMutex{},
		},
	}
}

func (r *scriptRepository) AddScript(
	ctx context.Context, info domain.AddressInfo,
) (bool, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	done, err := r.addScript(ctx, info)
	if err != nil {
		return false, err
	}

	if done {
		go r.publishEvent(domain.ExternalScriptEvent{
			EventType: domain.ExternalScriptAdded,
			Info:      info,
		})
	}

	return done, nil
}

func (r *scriptRepository) GetAllScripts(
	ctx context.Context,
) ([]domain.AddressInfo, error) {
	r.store.lock.RLock()
	defer r.store.lock.RUnlock()

	return r.getScripts(ctx)
}

func (r *scriptRepository) DeleteScript(
	ctx context.Context, scriptHash string,
) (bool, error) {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	done, err := r.deleteScript(ctx, scriptHash)
	if err != nil {
		return false, err
	}

	if done {
		go r.publishEvent(domain.ExternalScriptEvent{
			EventType: domain.ExternalScriptAdded,
			Info: domain.AddressInfo{
				Account: scriptHash,
			},
		})
	}

	return done, nil
}

func (r *scriptRepository) addScript(
	_ context.Context, info domain.AddressInfo,
) (bool, error) {
	if _, ok := r.store.scripts[info.Account]; ok {
		return false, nil
	}

	r.store.scripts[info.Account] = info

	return true, nil
}

func (r *scriptRepository) getScripts(
	_ context.Context,
) ([]domain.AddressInfo, error) {
	scripts := make([]domain.AddressInfo, 0, len(r.store.scripts))
	for _, info := range r.store.scripts {
		scripts = append(scripts, info)
	}
	return scripts, nil
}

func (r *scriptRepository) deleteScript(
	_ context.Context, scriptHash string,
) (bool, error) {
	if _, ok := r.store.scripts[scriptHash]; !ok {
		return false, nil
	}

	delete(r.store.scripts, scriptHash)

	return true, nil
}

func (r *scriptRepository) publishEvent(event domain.ExternalScriptEvent) {
	r.chLock.Lock()
	defer r.chLock.Unlock()

	r.chEvents <- event
}

func (r *scriptRepository) reset() {
	r.store.lock.Lock()
	defer r.store.lock.Unlock()

	r.store.scripts = make(map[string]domain.AddressInfo)
}

func (r *scriptRepository) close() {}
