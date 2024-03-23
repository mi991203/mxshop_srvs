package model

import (
	"time"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
	IsDeleted bool
}

type User struct {
	BaseModel
	Mobile string `gorm:"uniqueIndex:idx_mobile;type:varchar(11);not null"`
	Password string `gorm:"type:varchar(100);not null"`
	NickName string `gorm:"type:varchar(20)"`
	BirthDay *time.Time `gorm:"type:datetime"`
	Gender string `gorm:"column:gender;default:male;type:varchar(6) comment 'female表示女，male表示男'"`
	Role int `gorm:"column:role;defalut:1;type:int comment '1表示普通用户，2表示管理员'"`
}