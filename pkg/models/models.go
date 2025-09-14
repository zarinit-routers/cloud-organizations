package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name       string    `json:"name"`
	Passphrase *string   `json:"passphrase"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	Members    []Member  `json:"members"`
}
type Member struct {
	UserID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"userId"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
	OrganizationID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()" json:"organizationId"`
}
