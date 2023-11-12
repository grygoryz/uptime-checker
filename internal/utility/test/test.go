package test

import (
	"bytes"
	"encoding/json"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ExecuteRequest(s *server.Server, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router().ServeHTTP(rr, req)

	return rr
}

func CheckCode(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Errorf("want code %d, got %d", want, got)
	}
}

func Authorize(t *testing.T, s *server.Server) (string, auth.CheckResponse) {
	body, err := json.Marshal(auth.SignUpBody{Email: t.Name() + "@test.com", Password: "123123123"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := ExecuteRequest(s, req)
	CheckCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/auth/signin", bytes.NewReader(body))
	response = ExecuteRequest(s, req)
	CheckCode(t, http.StatusOK, response.Code)

	cookie := response.Header().Get("Set-Cookie")

	req, _ = http.NewRequest("GET", "/v1/auth/check", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response = ExecuteRequest(s, req)
	CheckCode(t, http.StatusOK, response.Code)

	var user auth.CheckResponse
	err = json.Unmarshal(response.Body.Bytes(), &user)
	if err != nil {
		t.Fatal(err)
	}

	return cookie, user
}
