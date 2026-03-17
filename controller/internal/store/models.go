package store

import "time"

type Project struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"uniqueIndex;not null"` // e.g. "my-app"
	ContainerID string    `gorm:"not null"`             // Docker container ID
	ContainerIP string    `gorm:"not null"`             // Internal Docker IP e.g. 172.17.0.2
	HostPort    string    `gorm:"not null"`             // Host port exposed on localhost e.g. "10123"
	ImageName   string    `gorm:"not null"`             // e.g. "my-app:latest"
	Status      string    `gorm:"default:'running'"`    // "running" | "stopped" | "error"
	CreatedAt   time.Time                               // auto-set by GORM
	UpdatedAt   time.Time                               // auto-updated on Save
}
