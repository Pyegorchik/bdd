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


type User struct {
	ID      int64
	Role    Role
	Address common.Address
}


func UserToModel(u *User) *models.UserInfo {
	return &models.UserInfo{
		Address: u.Address.String(),
	}
}

