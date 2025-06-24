package models

type Student struct {
    ID         uint      `gorm:"primaryKey"`
    FirstName  string    `gorm:"not null"`
    LastName   string    `gorm:"not null"`
    Email      string    `gorm:"unique;not null"`
    Password   string    `gorm:"not null"` // hashed
    Passkey    string    `gorm:"not null"`
}