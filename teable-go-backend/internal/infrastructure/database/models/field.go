package models

import (
	"time"

)

// Field 字段定义（简化基础字段类型）
type Field struct {
	ID                  string     `gorm:"primaryKey;type:varchar(30)" json:"id"`
	TableID             string     `gorm:"column:table_id;index;type:varchar(30);not null" json:"table_id"`
	Name                string     `gorm:"not null;type:varchar(255)" json:"name"`
	Description         *string    `gorm:"type:text" json:"description"`
	Type                string     `gorm:"type:varchar(50);not null" json:"type"`
	CellValueType       string     `gorm:"column:cell_value_type;type:varchar(50);not null" json:"cell_value_type"`
	IsMultipleCellValue *bool      `gorm:"column:is_multiple_cell_value" json:"is_multiple_cell_value"`
	DbFieldType         string     `gorm:"column:db_field_type;type:varchar(50);not null" json:"db_field_type"`
	DbFieldName         string     `gorm:"column:db_field_name;type:varchar(255);not null" json:"db_field_name"`
	NotNull             *bool      `gorm:"column:not_null" json:"not_null"`
	Unique              *bool      `gorm:"column:unique" json:"unique"`
	IsPrimary           *bool      `gorm:"column:is_primary" json:"is_primary"`
	IsComputed          *bool      `gorm:"column:is_computed" json:"is_computed"`
	IsLookup            *bool      `gorm:"column:is_lookup" json:"is_lookup"`
	Order               float64    `gorm:"type:decimal(10,2)" json:"order"`
	Version             int        `gorm:"type:int;default:1" json:"version"`
	CreatedBy           string     `gorm:"column:created_by;type:varchar(30);not null" json:"created_by"`
	CreatedTime         time.Time  `gorm:"autoCreateTime;column:created_time" json:"created_time"`
	LastModifiedTime    *time.Time `gorm:"autoUpdateTime;column:last_modified_time" json:"last_modified_time"`
}

func (Field) TableName() string { return "field" }

