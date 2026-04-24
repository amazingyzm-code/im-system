package db

import (
	"im-system/pkg/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var err error
	DB, err = gorm.Open(mysql.Open(config.Global.MySQL.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return DB.AutoMigrate(&User{}, &Message{}, &GroupMember{})
}

// User 用户表
type User struct {
	ID           int64     `gorm:"primaryKey"`
	Username     string    `gorm:"uniqueIndex;size:64"`
	PasswordHash string    `gorm:"size:256"`
	Avatar       string    `gorm:"size:256"`
	CreatedAt    time.Time
}

// Message 消息持久化表
type Message struct {
	ID         int64  `gorm:"primaryKey"`           // 雪花 ID
	MsgType    int    `gorm:"index"`
	TargetType int
	FromUID    int64  `gorm:"index"`
	ToID       int64  `gorm:"index"`                // 单聊=接收方UID，群聊=群ID
	Content    string `gorm:"type:text"`
	Timestamp  int64  `gorm:"index"`
}

// GroupMember 群组成员表
type GroupMember struct {
	GroupID  int64 `gorm:"primaryKey;autoIncrement:false"`
	UID      int64 `gorm:"primaryKey;autoIncrement:false"`
	Role     int   // 0=普通成员 1=管理员 2=群主
	JoinedAt int64
}
