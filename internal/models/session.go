package models

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"slashbase.com/backend/internal/config"
	"slashbase.com/backend/internal/db"
)

type UserSession struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"not null"`
	IsActive  bool
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	User User `gorm:"foreignkey:user_id"`
}

func NewUserSession(userID string) (*UserSession, error) {
	var err error = nil
	if userID == "" {
		return nil, errors.New("user id cannot be empty")
	}
	return &UserSession{
		ID:       uuid.NewString(),
		UserID:   userID,
		IsActive: true,
	}, err
}

func (session UserSession) GetAuthToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sessionID": session.ID,
	})
	tokenString, err := token.SignedString([]byte(config.GetConfig().AuthTokenSecret))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func (session UserSession) SetInActive() error {
	session.IsActive = false
	return session.Save()
}

func (session UserSession) Save() error {
	return db.GetDB().Save(&session).Error
}
