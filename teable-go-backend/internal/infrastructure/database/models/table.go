package models

import (
	"time"

	"gorm.io/gorm"
)

// Base 数据库（逻辑库）模型
type Base struct {
	ID               string         `gorm:"primaryKey;type:varchar(30)" json:"id"`
	SpaceID          string         `gorm:"column:space_id;index;type:varchar(30);not null" json:"space_id"`
	Name             string         `gorm:"not null;type:varchar(255)" json:"name"`
	Description      *string        `gorm:"type:text" json:"description"`
	Icon             *string        `gorm:"type:varchar(100)" json:"icon"`
	CreatedBy        string         `gorm:"column:created_by;type:varchar(30);not null" json:"created_by"`
	CreatedTime      time.Time      `gorm:"autoCreateTime;column:created_time" json:"created_time"`
	DeletedTime      gorm.DeletedAt `gorm:"column:deleted_time" json:"deleted_time,omitempty"`
	LastModifiedTime *time.Time     `gorm:"autoUpdateTime;column:last_modified_time" json:"last_modified_time"`
}

func (Base) TableName() string { return "base" }

// TableMeta 表元数据模型（逻辑表定义）
type TableMeta struct {
	ID               string         `gorm:"primaryKey;type:varchar(30)" json:"id"`
	BaseID           string         `gorm:"column:base_id;index;type:varchar(30);not null" json:"base_id"`
	Name             string         `gorm:"not null;type:varchar(255)" json:"name"`
	Description      *string        `gorm:"type:text" json:"description"`
	Icon             *string        `gorm:"type:varchar(100)" json:"icon"`
	DbTableName      string         `gorm:"column:db_table_name;uniqueIndex;type:varchar(255);not null" json:"db_table_name"`
	Order            int            `gorm:"type:int;default:0" json:"order"`
	Version          int            `gorm:"type:int;default:1" json:"version"`
	CreatedBy        string         `gorm:"column:created_by;type:varchar(30);not null" json:"created_by"`
	CreatedTime      time.Time      `gorm:"autoCreateTime;column:created_time" json:"created_time"`
	DeletedTime      gorm.DeletedAt `gorm:"column:deleted_time" json:"deleted_time,omitempty"`
	LastModifiedTime *time.Time     `gorm:"autoUpdateTime;column:last_modified_time" json:"last_modified_time"`
}

func (TableMeta) TableName() string { return "table_meta" }

