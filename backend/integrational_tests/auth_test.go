package integrationstests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	h "github.com/Pyegorchik/bdd/backend/internal/handler"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-cmp/cmp"
)

func (s *TestSuiteUser) TestAuthByMsg() {
	//
	// Default auth by address
	//

	/// получение сообщения для зашифровки
	addr := s.accounts[1].auth.From.String()
	dataReq, err := json.Marshal(models.AuthMessageRequest{
		Address: &addr,
	})
	s.Require().NoError(err)

	httpReq := httptest.NewRequest(http.MethodPost, "/g1/auth/message", bytes.NewReader(dataReq))
	recorder := httptest.NewRecorder()
	s.handler.ServeHTTP(recorder, httpReq)
	resp := recorder.Result()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	defer resp.Body.Close()
	s.Require().Equal(200, resp.StatusCode)

	var respMsg models.AuthMessageResponse
	s.Require().NoError(json.NewDecoder(bytes.NewReader(data)).Decode(&respMsg))

	// формирование сигнатуры
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(*respMsg.Message), *respMsg.Message)))
	sig, err := crypto.Sign(hash, s.accounts[1].pk)
	s.Require().NoError(err)

	signature := hexutil.Encode(sig)

	// запрос авторизации/регистрации по сообщению
	dataReq, err = json.Marshal(models.AuthBySignatureRequest{
		Address:   &addr,
		Signature: &signature,
	})
	s.Require().NoError(err)

	httpReq = httptest.NewRequest(http.MethodPost, "/g1/auth/by_signature", bytes.NewReader(dataReq))

	recorder = httptest.NewRecorder()
	s.handler.ServeHTTP(recorder, httpReq)
	resp = recorder.Result()

	defer resp.Body.Close()

	s.Require().Equal(200, resp.StatusCode)

	cookies := resp.Cookies()
	for _, c := range cookies {
		if c.Name == h.NameCookie {
			s.Require().NotEmpty(c.Value)
		}
	}

	dataAuth, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	var resAuth models.AuthResponse
	s.Require().NoError(json.NewDecoder(bytes.NewReader(dataAuth)).Decode(&resAuth))

	diff := cmp.Diff(s.accounts[1].auth.From.String(), resAuth.User.Address)
	s.Require().Empty(diff)
}
