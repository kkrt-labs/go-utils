package multistore

import (
	"context"
	"fmt"
	"io"

	store "github.com/kkrt-labs/go-utils/store"
)

type Store struct {
	stores []store.Store
}

func New(stores ...store.Store) store.Store {
	return &Store{stores: stores}
}

func (m *Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	for _, s := range m.stores {
		if err := s.Store(ctx, key, reader, headers); err != nil {
			return err
		}
	}
	return nil
}

func (m *Store) Load(ctx context.Context, key string, headers *store.Headers) (io.ReadCloser, error) {
	// Try stores in order until we find the data or encounter an error
	for _, s := range m.stores {
		reader, err := s.Load(ctx, key, headers)
		if err != nil {
			return nil, fmt.Errorf("failed to load from store: %w", err)
		}
		if reader != nil {
			return reader, nil
		}
	}
	return nil, fmt.Errorf("key %s not found in any store", key)
}
