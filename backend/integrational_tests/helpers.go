package integrationstests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	h "github.com/Pyegorchik/bdd/backend/internal/handler"
	"github.com/Pyegorchik/bdd/backend/models"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func makeAuthRequest(handler http.Handler, account *Signer) (string, error) {
	addr := account.auth.From.String()
	dataReq, err := json.Marshal(models.AuthMessageRequest{
		Address: &addr,
	})
	if err != nil {
		return "", err
	}

	httpReq := httptest.NewRequest(http.MethodPost, "/g1/auth/message", bytes.NewReader(dataReq))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpReq)
	resp := recorder.Result()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non 200 response: %d %s", resp.StatusCode, string(data))
	}

	var respMsg models.AuthMessageResponse
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&respMsg)
	if err != nil {
		return "", err
	}

	// формирование сигнатуры
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(*respMsg.Message), *respMsg.Message)))
	sig, err := crypto.Sign(hash, account.pk)
	if err != nil {
		return "", err
	}

	signature := hexutil.Encode(sig)

	// запрос авторизации/регистрации по сообщению
	dataReq, err = json.Marshal(models.AuthBySignatureRequest{
		Address:   &addr,
		Signature: &signature,
	})
	if err != nil {
		return "", err
	}

	httpReq = httptest.NewRequest(http.MethodPost, "/g1/auth/by_signature", bytes.NewReader(dataReq))

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpReq)
	resp = recorder.Result()

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non 200 response: %d %s", resp.StatusCode, string(data))
	}

	cookies := resp.Cookies()
	for _, c := range cookies {
		if c.Name == h.NameCookie {
			return c.Value, nil
		}
	}

	return "", fmt.Errorf("not found cookie in response")
}

func makeJsonRequest(handler http.Handler, cookie string, method string, url string, body any, res any) error {
	var b io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		b = bytes.NewReader(data)
	}
	httpReq := httptest.NewRequest(method, url, b)
	httpReq.AddCookie(&http.Cookie{
		Name:     h.NameCookie,
		Value:    cookie,
		Expires:  time.Now().Add(time.Minute),
		Secure:   true,
		HttpOnly: true,
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpReq)
	resp := recorder.Result()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d %s", resp.StatusCode, string(data))
	}

	if res != nil {
		if err = json.Unmarshal(data, res); err != nil {
			return fmt.Errorf("json unmarshal failed: %w %s", err, string(data))
		}
	}
	return nil
}

func makeJsonRequestWithError(handler http.Handler, cookie string, method string, url string, body any, res any) error {
	var b io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		b = bytes.NewReader(data)
	}
	httpReq := httptest.NewRequest(method, url, b)
	httpReq.AddCookie(&http.Cookie{
		Name:     "access-token",
		Value:    cookie,
		Expires:  time.Now().Add(time.Minute),
		Secure:   true,
		HttpOnly: true,
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpReq)
	resp := recorder.Result()
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return fmt.Errorf("200 response: %d %s", resp.StatusCode, string(data))
	}

	if res != nil {
		if err = json.Unmarshal(data, res); err != nil {
			return fmt.Errorf("json unmarshal failed: %w %s", err, string(data))
		}
	}

	return nil
}
