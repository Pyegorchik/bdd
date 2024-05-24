package domain

import (
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/ethereum/go-ethereum/common"
)

type (
	Role int
)

type UserWithTokenNumber struct {
	ID         int64
	Role       Role
	Number     int
	Authorized bool
}

type AuthMessage struct {
	Address   string
	Message   string
	CreatedAt int64
}

type UserChain struct {
	ID      int64
	Role    Role
	Address common.Address
}

type Dialog struct {
	ID int64
}

type DialogParticipant struct {
	DialogID int64
	UserID   int64
}

type Message struct {
	ID       int64
	DialogID int64
	SenderID int64
	Content  string
}

type User struct {
	ID           int64
	Username     string
	PasswordHash string
}

func UserToModel(u *UserChain) *models.UserInfo {
	return &models.UserInfo{
		Address: u.Address.String(),
	}
}
