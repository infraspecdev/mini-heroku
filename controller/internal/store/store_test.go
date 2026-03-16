package store

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	dbPath := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return s
}

func makeProject(name string) *Project {
	return &Project{
		Name:        name,
		ContainerID: "container-" + name,
		ContainerIP: "172.17.0.2",
		HostPort:    "10123",
		ImageName:   name + ":latest",
		Status:      "running",
	}
}

func TestNewStoreAndUpsert(t *testing.T) {
	s := newTestStore(t)
	if err := s.Upsert(makeProject("app-one")); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
}

func TestGetByName(t *testing.T) {
	s := newTestStore(t)
	want := makeProject("my-app")
	if err := s.Upsert(want); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := s.GetByName("my-app")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if got.Name != want.Name {
		t.Fatalf("GetByName().Name = %q, want %q", got.Name, want.Name)
	}
	if got.Status != want.Status {
		t.Fatalf("GetByName().Status = %q, want %q", got.Status, want.Status)
	}
}

func TestGetByNameNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.GetByName("missing")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("GetByName() error = %v, want gorm.ErrRecordNotFound", err)
	}
}

func TestListAll(t *testing.T) {
	s := newTestStore(t)
	if err := s.Upsert(makeProject("app-a")); err != nil {
		t.Fatalf("Upsert(app-a) error = %v", err)
	}
	if err := s.Upsert(makeProject("app-b")); err != nil {
		t.Fatalf("Upsert(app-b) error = %v", err)
	}

	projects, err := s.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("ListAll() len = %d, want 2", len(projects))
	}
}

func TestUpdateStatus(t *testing.T) {
	s := newTestStore(t)
	if err := s.Upsert(makeProject("status-app")); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	if err := s.UpdateStatus("status-app", "stopped"); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	got, err := s.GetByName("status-app")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if got.Status != "stopped" {
		t.Fatalf("GetByName().Status = %q, want %q", got.Status, "stopped")
	}
}
