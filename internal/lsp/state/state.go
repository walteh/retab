// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package state

import (
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/go-memdb"
)

const (
	documentsTableName      = "documents"
	jobsTableName           = "jobs"
	moduleTableName         = "module"
	moduleIdsTableName      = "module_ids"
	moduleChangesTableName  = "module_changes"
	providerSchemaTableName = "provider_schema"
	providerIdsTableName    = "provider_ids"
	walkerPathsTableName    = "walker_paths"
	registryModuleTableName = "registry_module"

	tracerName = "github.com/walteh/retab/internal/lsp/state"
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
		jobsTableName: {
			Name: jobsTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &StringerFieldIndexer{Field: "ID"},
				},
				"priority_dependecies_state": {
					Name: "priority_dependecies_state",
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&JobPriorityIndex{
								PriorityIntField:   "Priority",
								IsDirOpenBoolField: "IsDirOpen",
							},
							&SliceLengthIndex{Field: "DependsOn"},
							&memdb.UintFieldIndex{Field: "State"},
						},
					},
				},
				"dir_state": {
					Name: "dir_state",
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&DirHandleFieldIndexer{Field: "Dir"},
							&memdb.UintFieldIndex{Field: "State"},
						},
					},
				},
				"dir_state_type": {
					Name: "dir_state_type",
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&DirHandleFieldIndexer{Field: "Dir"},
							&memdb.UintFieldIndex{Field: "State"},
							&memdb.StringFieldIndex{Field: "Type"},
						},
					},
				},
				"state_type": {
					Name: "state_type",
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.UintFieldIndex{Field: "State"},
							&memdb.StringFieldIndex{Field: "Type"},
						},
					},
				},
				"state": {
					Name: "state",
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.UintFieldIndex{Field: "State"},
						},
					},
				},
				"depends_on": {
					Name: "depends_on",
					Indexer: &JobIdSliceIndex{
						Field: "DependsOn",
					},
					AllowMissing: true,
				},
			},
		},
		moduleTableName: {
			Name: moduleTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Path"},
				},
			},
		},
		providerSchemaTableName: {
			Name: providerSchemaTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&StringerFieldIndexer{Field: "Address"},
							&StringerFieldIndexer{Field: "Source"},
							&VersionFieldIndexer{Field: "Version"},
						},
						AllowMissing: true,
					},
				},
			},
		},
		registryModuleTableName: {
			Name: registryModuleTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&StringerFieldIndexer{Field: "Source"},
							&VersionFieldIndexer{Field: "Version"},
						},
						AllowMissing: true,
					},
				},
				"source_addr": {
					Name:    "source_addr",
					Indexer: &StringerFieldIndexer{Field: "Source"},
				},
			},
		},
		providerIdsTableName: {
			Name: providerIdsTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Address"},
				},
			},
		},
		moduleIdsTableName: {
			Name: moduleIdsTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Path"},
				},
			},
		},
		moduleChangesTableName: {
			Name: moduleChangesTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &DirHandleFieldIndexer{Field: "DirHandle"},
				},
				"time": {
					Name:    "time",
					Indexer: &TimeFieldIndex{Field: "FirstChangeTime"},
				},
			},
		},
		walkerPathsTableName: {
			Name: walkerPathsTableName,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &DirHandleFieldIndexer{Field: "Dir"},
				},
				"is_dir_open_state": {
					Name: "is_dir_open_state",
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.BoolFieldIndex{Field: "IsDirOpen"},
							&memdb.UintFieldIndex{Field: "State"},
						},
					},
				},
			},
		},
	},
}

type StateStore struct {
	DocumentStore *DocumentStore
	JobStore      *JobStore
	WalkerPaths   *WalkerPathStore
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
		JobStore: &JobStore{
			db:                db,
			tableName:         jobsTableName,
			logger:            defaultLogger,
			nextJobHighPrioMu: &sync.Mutex{},
			nextJobLowPrioMu:  &sync.Mutex{},
		},
		WalkerPaths: &WalkerPathStore{
			db:              db,
			tableName:       walkerPathsTableName,
			logger:          defaultLogger,
			nextOpenDirMu:   &sync.Mutex{},
			nextClosedDirMu: &sync.Mutex{},
		},
	}, nil
}

func (s *StateStore) SetLogger(logger *log.Logger) {
	s.DocumentStore.logger = logger
	s.JobStore.logger = logger
	s.WalkerPaths.logger = logger
}

var defaultLogger = log.New(ioutil.Discard, "", 0)
