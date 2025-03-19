package memory

import (
	"bytes"
	"context"
	"fmt"
	"io"

	store "github.com/kkrt-labs/go-utils/store"
)

type Store struct {
	data map[string][]byte
}

func New() store.Store {
	return &Store{
		data: make(map[string][]byte),
	}
}

func (s *Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	s.data[key] = data
	return nil
}

func (s *Store) Load(ctx context.Context, key string, headers *store.Headers) (io.ReadCloser, error) {
	data, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}
