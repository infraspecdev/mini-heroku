package store

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Println("open db error:", err)
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.AutoMigrate(&Project{}); err != nil {
		fmt.Println("automigrate error:", err)
		return nil, fmt.Errorf("automigrate: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Upsert(p *Project) error {
	return s.db.Save(p).Error
}

func (s *Store) GetByName(name string) (*Project, error) {
	var p Project
	result := s.db.Where("name = ?", name).First(&p)
	if result.Error != nil {
		return nil, result.Error
	}
	return &p, nil
}

func (s *Store) ListAll() ([]Project, error) {
	var projects []Project
	result := s.db.Find(&projects)
	return projects, result.Error
}

func (s *Store) UpdateStatus(name, status string) error {
	return s.db.Model(&Project{}).
		Where("name = ?", name).
		Update("status", status).Error
}
