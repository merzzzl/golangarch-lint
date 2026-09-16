package migrations

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

var (
	ErrVersion           = errors.New("migration requires version 1")
	ErrInvalidMode       = errors.New("invalid v1 declaration mode")
	ErrMultipleDocuments = errors.New("config must contain exactly one YAML document")
)

// Migrate decodes legacy v1 strictly and emits v2 YAML, preserving $module.
func Migrate(data []byte) ([]byte, []string, error) {
	var old V1

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(&old); err != nil {
		return nil, nil, fmt.Errorf("decoding v1: %w", err)
	}

	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, nil, ErrMultipleDocuments
	}

	if old.Version != 1 {
		return nil, nil, fmt.Errorf("%w, got %d", ErrVersion, old.Version)
	}

	return old.Migrate()
}
