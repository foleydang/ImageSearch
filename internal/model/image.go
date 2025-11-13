package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Image 图片模型
type Image struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	FileName  string    `gorm:"size:255;not null" json:"file_name"`
	FilePath  string    `gorm:"size:255;not null" json:"file_path"`
	Extension string    `gorm:"size:10;not null" json:"extension"`
	Width     int       `gorm:"not null" json:"width"`
	Height    int       `gorm:"not null" json:"height"`
	Size      int64     `gorm:"not null" json:"size"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// ImageEmbedding 图片嵌入向量模型
type ImageEmbedding struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	ImageID   uuid.UUID `gorm:"type:uuid;not null;index" json:"image_id"`
	Embedding []float32 `gorm:"type:blob;not null" json:"embedding"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	Image     Image     `gorm:"foreignKey:ImageID" json:"image,omitempty"`
}

// BeforeCreate 创建前的钩子函数，用于生成UUID
func (i *Image) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// BeforeCreate 创建前的钩子函数，用于生成UUID
func (ie *ImageEmbedding) BeforeCreate(tx *gorm.DB) error {
	if ie.ID == uuid.Nil {
		ie.ID = uuid.New()
	}
	return nil
}