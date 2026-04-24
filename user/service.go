package user

import (
	"errors"
	"im-system/pkg/db"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	jwtSecret string
}

func NewService(jwtSecret string) *Service {
	return &Service{jwtSecret: jwtSecret}
}

func (s *Service) Register(username, password string) (*db.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &db.User{Username: username, PasswordHash: string(hash)}
	if err := db.DB.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(username, password string) (string, error) {
	var user db.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user not found")
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("wrong password")
	}
	return generateToken(user.ID, s.jwtSecret)
}

func (s *Service) GetByID(uid int64) (*db.User, error) {
	var user db.User
	if err := db.DB.First(&user, uid).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
