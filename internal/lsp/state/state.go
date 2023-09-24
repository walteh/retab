// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/hashicorp/go-memdb"
)

const (
	documentsTableName = "documents"
	tracerName         = "github.com/walteh/retab/internal/lsp/state"
)

var dbSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		documentsTableName: {
			Name: documentsTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&DirHandleFieldIndexer{Field: "Dir"},
							&memdb.StringFieldIndex{Field: "Filename"},
						},
					},
				},
				"dir": {
					Name:    "dir",
					Indexer: &DirHandleFieldIndexer{Field: "Dir"},
				},
			},
		},
	},
}

type StateStore struct {
	DocumentStore *DocumentStore
	db            *memdb.MemDB
}

func NewStateStore() (*StateStore, error) {
	db, err := memdb.NewMemDB(dbSchema)
	if err != nil {
		return nil, err
	}

	return &StateStore{
		db: db,
		DocumentStore: &DocumentStore{
			db:           db,
			tableName:    documentsTableName,
			logger:       defaultLogger,
			TimeProvider: time.Now,
		},
	}, nil
}

func (s *StateStore) SetLogger(logger *log.Logger) {
	s.DocumentStore.logger = logger
}

var defaultLogger = log.New(ioutil.Discard, "", 0)
