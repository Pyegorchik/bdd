package integrationstests

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Pyegorchik/bdd/backend/models"
)

func (s *TestSuiteUser) TestSendMessage() {
	sender := s.accounts[1]
	senderAddress := sender.auth.From.String()
	cookie, err := makeAuthRequest(s.handler, sender)
	s.Require().NoError(err)

	recepeintAddress := s.accounts[2].auth.From.String()
	noContent := ""

	emptyMsg := &models.SendMessageRequest{
		Content:     &noContent,
		RecipientID: &recepeintAddress,
	}

	// Recepeint not registered
	var resErr *models.ErrorResponse
	err = makeJsonRequestWithError(s.handler, cookie, http.MethodPost, "/g1/dialogs/message", emptyMsg, &resErr)
	s.Require().NoError(err)

	resTargetError := &models.ErrorResponse{
		Code:    400,
		Detail:  fmt.Sprintf("user with address %v is not registered", recepeintAddress),
		Message: "Bad Request",
	}
	s.Require().Equal(resTargetError, resErr)

	_, err = makeAuthRequest(s.handler, s.accounts[2])
	s.Require().NoError(err)

	// Succeed
	content := "messageTearDownTest3"
	msg := &models.SendMessageRequest{
		Content:     &content,
		RecipientID: &recepeintAddress,
	}

	var resSuccessfull *models.SuccessResponse
	err = makeJsonRequest(s.handler, cookie, http.MethodPost, "/g1/dialogs/message", msg, &resSuccessfull)
	s.Require().NoError(err)

	resTrue := true
	resTargetSuccessfull := &models.SuccessResponse{
		Success: &resTrue,
	}
	s.Require().Equal(resTargetSuccessfull, resSuccessfull)

	var resDialogs *models.DialogsResponse
	err = makeJsonRequest(s.handler, cookie, http.MethodGet, "/g1/dialogs", nil, &resDialogs)
	s.Require().NoError(err)

	targetDialogs := models.DialogsResponse([]*models.DialogsResponseItems0{{DialogID: 1, RecepeintAddress: strings.ToLower(recepeintAddress)}})
	s.Require().Equal(&targetDialogs, resDialogs)

	dialogId := 1
	messageId := int64(1)

	var resDialogMessages *models.MessagesResponse
	err = makeJsonRequest(s.handler, cookie, http.MethodGet, fmt.Sprintf("/g1/dialogs/%d/messages", dialogId), nil, &resDialogMessages)
	s.Require().NoError(err)

	targetDialogMessages := models.MessagesResponse(
		[]*models.MessagesResponseItems0{{MessageID: messageId, SenderAddress: strings.ToLower(senderAddress), Content: content}})
	s.Require().Equal(&targetDialogMessages, resDialogMessages)
}
