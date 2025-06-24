package modules

import (
	"time"
)

type Base struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:create time"`
	UpdatedAt time.Time `json:"updated_at" gorm:"comment:update time"`
}

type BaseWithDeletedAt struct {
	Base
	DeletedAt *time.Time `json:"deleted_at" gorm:"comment:delete time"`
}

func (v *Base) GetId() uint {
	return v.ID
}
