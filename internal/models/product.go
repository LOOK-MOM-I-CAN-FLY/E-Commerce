
package models

import "time"

type Product struct {
    ID          uint      `gorm:"primaryKey"`
    Title       string    `gorm:"not null"`
    Description string
    FilePath    string    `gorm:"not null"`
    UserID      uint
    CreatedAt   time.Time
}
