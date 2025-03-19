package compress

import (
	"bytes"
	"context"
	"io"
	"testing"

	store "github.com/kkrt-labs/go-utils/store"
	"github.com/kkrt-labs/go-utils/store/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	memStore := memory.New()

	tests := []struct {
		desc      string
		encoding  store.ContentEncoding
		key       string
		data      []byte
		headers   *store.Headers
		expectErr bool
	}{
		{
			desc:      "gzip",
			encoding:  store.ContentEncodingGzip,
			key:       "gzip",
			data:      []byte("gzip data"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			s := New(memStore, tt.encoding)
			ctx := context.TODO()

			// Store the data
			err := s.Store(ctx, tt.key, bytes.NewReader(tt.data), tt.headers)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Load the data
			reader, err := s.Load(ctx, tt.key, tt.headers)
			require.NoError(t, err)

			loadedData, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, tt.data, loadedData)
		})
	}
}
