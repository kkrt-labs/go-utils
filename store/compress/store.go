package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"fmt"
	"io"

	store "github.com/kkrt-labs/go-utils/store"
)

type Store struct {
	store           store.Store
	contentEncoding store.ContentEncoding
}

func New(store store.Store, encoding store.ContentEncoding) store.Store {
	return &Store{
		store:           store,
		contentEncoding: encoding,
	}
}

func (c *Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	var compressedReader io.Reader

	switch c.contentEncoding {
	case store.ContentEncodingGzip:
		fmt.Println("compressing with gzip")
		buf := bytes.NewBuffer(nil)
		gw := gzip.NewWriter(buf)
		defer gw.Close()
		if _, err := io.Copy(gw, reader); err != nil {
			return fmt.Errorf("failed to compress with gzip: %w", err)
		}
		compressedReader = buf

	case store.ContentEncodingZlib:
		fmt.Println("compressing with zlib")
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		defer zw.Close()
		if _, err := io.Copy(zw, reader); err != nil {
			return fmt.Errorf("failed to compress with zlib: %w", err)
		}
		compressedReader = &buf

	case store.ContentEncodingFlate:
		fmt.Println("compressing with flate")
		var buf bytes.Buffer
		fw, err := flate.NewWriter(&buf, flate.BestCompression)
		if err != nil {
			return fmt.Errorf("failed to create flate writer: %w", err)
		}
		defer fw.Close()
		if _, err := io.Copy(fw, reader); err != nil {
			return fmt.Errorf("failed to compress with flate: %w", err)
		}
		compressedReader = &buf

	case store.ContentEncodingPlain:
		fmt.Println("compressing with plain")
		compressedReader = reader
	default:
		return fmt.Errorf("unsupported content encoding: %s", c.contentEncoding)
	}

	if headers == nil {
		headers = &store.Headers{}
	}
	headers.ContentEncoding = c.contentEncoding

	return c.store.Store(ctx, key, compressedReader, headers)
}

func (c *Store) Load(ctx context.Context, key string, headers *store.Headers) (io.ReadCloser, error) {
	if headers == nil {
		headers = &store.Headers{}
	}
	headers.ContentEncoding = c.contentEncoding

	reader, err := c.store.Load(ctx, key, headers)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		switch headers.ContentEncoding {
		case store.ContentEncodingGzip:
			fmt.Println("decompressing with gzip")
			return gzip.NewReader(reader)
		case store.ContentEncodingZlib:
			fmt.Println("decompressing with zlib")
			return zlib.NewReader(reader)
		case store.ContentEncodingFlate:
			fmt.Println("decompressing with flate")
			return flate.NewReader(reader), nil
		case store.ContentEncodingPlain:
			fmt.Println("decompressing with plain")
			return reader, nil
		default:
			return nil, fmt.Errorf("unsupported content encoding: %s", headers.ContentEncoding)
		}
	}

	return reader, nil
}
