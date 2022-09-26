package sqlite

import (
	"github.com/myfintech/ark/src/go/lib/ark"
	"github.com/myfintech/ark/src/go/lib/ark/derivation"
	"github.com/myfintech/ark/src/go/lib/ark/storage/graph"
	"github.com/myfintech/ark/src/go/lib/dag"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DefaultStore is the default sqlite storage interface for Ark

// Store implements the Store interface
type Store struct {
	DB *gorm.DB `gorm:"-"`
}

// GetTargetByKey not implemented
// FIXME: implement GetTargetByKey
func (s *Store) GetTargetByKey(_ string) (ark.RawTarget, error) {
	panic("implement me")
}

// Open opens the db file that was passed
func (s *Store) Open(connection string) error {
	db, err := gorm.Open(sqlite.Open(connection), &gorm.Config{})
	if err != nil {
		return err
	}
	s.DB = db
	return nil
}

// Migrate ...
func (s *Store) Migrate() error {
	return s.DB.AutoMigrate(
		&ark.RawTarget{},
		&ark.GraphEdge{},
	)
}

// GetTargets returns a list of targets
func (s *Store) GetTargets() ([]ark.RawTarget, error) {
	var targets []ark.RawTarget

	result := s.DB.Find(&targets)

	return targets, result.Error
}

// AddTarget adds a new target to the database
func (s *Store) AddTarget(target ark.RawTarget) (artifact ark.RawArtifact, err error) {
	artifact, err = derivation.RawArtifactFromRawTarget(target)
	if err != nil {
		return
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.FirstOrCreate(&target).Error; err != nil {
			return err
		}

		return tx.Updates(&target).Error
	})

	if err != nil {
		return
	}
	return
}

// ConnectTargets creates edges between source and destination targets
// The data in this table is used to build a graph
func (s *Store) ConnectTargets(edge ark.GraphEdge) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		return tx.FirstOrCreate(&edge).Error
	})
}

// GetGraphEdges find all the edges to build a graph
func (s Store) GetGraphEdges() ([]ark.GraphEdge, error) {
	var edges []ark.GraphEdge

	result := s.DB.Find(&edges)

	return edges, result.Error
}

// GetGraph finds all targets and edges and builds a directed acyclic graph
func (s *Store) GetGraph() (*dag.AcyclicGraph, error) {
	return graph.FromStore(s)
}
