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
	DialogID   int64
	UserAdress string
}

type Message struct {
	ID            int64
	DialogID      int64
	SenderAddress string
	SenderID      int64
	Content       string
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

func RecepientsToRecepinetsResponce(rc []*DialogParticipant) []*models.DialogsResponseItems0 {
	var res []*models.DialogsResponseItems0
	for _, v := range rc {
		res = append(res, &models.DialogsResponseItems0{
			RecepeintAddress: v.UserAdress,
			DialogID:         v.DialogID,
		})
	}

	return res
}

func MessageToMessageResponse(msgs []*Message) []*models.MessagesResponseItems0 {
	var res []*models.MessagesResponseItems0
	for _, v := range msgs {
		res = append(res, &models.MessagesResponseItems0{
			SenderAddress: v.SenderAddress,
			Content:       v.Content,
			MessageID:     v.ID,
		})
	}

	return res
}
