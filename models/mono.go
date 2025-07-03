package models

import "gorm.io/gorm"

type MonoSession struct {
	gorm.Model
	StudentID   uint   // foreign key to students table
	Reference   string `gorm:"unique"`
	MonoURL     string
	CustomerID  string
	AccountID  string

}
