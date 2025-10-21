package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	FirstName string `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName  string `gorm:"not null" json:"last_name"`
	Nickname  string `gorm:"not null;uniqueIndex" json:"nickname"`
	Email     string `gorm:"not null;uniqueIndex" json:"email"`
	Password  string `gorm:"not null" json:"-"`
}
