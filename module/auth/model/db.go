package model

import (
	"context"
	"fmt"
	"sync"

	"github.com/tkeel-io/tkeel/keel"
	"github.com/tkeel-io/tkeel/logger"

	daprc "github.com/dapr/go-sdk/client"
)

var (
	_db   *database
	_once sync.Once

	_log = logger.NewLogger("Keel.PluginAuth")
)

type DB interface {
	Insert(ctx context.Context, key string, data []byte) error
	Select(ctx context.Context, key string) ([]byte, error)
}

type database struct {
	dClient daprc.Client
}

func getDB() DB {
	c := keel.GetClient()
	_once.Do(
		func() {
			_db = &database{c}
		})
	return _db
}

func (db *database) Insert(ctx context.Context, key string, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := db.dClient.SaveState(ctx, keel.PrivateStore, key, data); err != nil {
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func (db *database) Select(ctx context.Context, key string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	item, err := db.dClient.GetState(ctx, keel.PrivateStore, key)
	if err != nil {
		return nil, fmt.Errorf("error get state: %w", err)
	}
	return item.Value, nil
}
