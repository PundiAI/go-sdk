package model

import (
	"time"
)

type Base struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:create time"`
	UpdatedAt time.Time `json:"updated_at" gorm:"comment:update time"`
}

func (b *Base) GetId() uint {
	return b.ID
}

func (b *Base) GetCreatedTimestamp() int64 {
	return b.CreatedAt.UnixMilli()
}

func (b *Base) GetUpdatedTimestamp() int64 {
	return b.UpdatedAt.UnixMilli()
}
