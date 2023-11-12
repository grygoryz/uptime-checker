package auth_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/auth"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/test"
	"net/http"
	"strings"
	"testing"
)

var s *server.Server

func TestMain(m *testing.M) {
	cfg := config.New(true)
	s = server.New(cfg)
	s.Init()
	m.Run()
}

func TestHandler_SignUp_ValidInput(t *testing.T) {
	dto := auth.SignUpBody{Email: "authtest1@test.com", Password: "123123123"}
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)

	test.CheckCode(t, http.StatusCreated, response.Code)
	var user entity.User
	err = s.DB().Get(&user, "SELECT id, email, password FROM users WHERE email = $1", dto.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("want user to exist in db")
		}
		t.Fatal(err)
	}
	if user.Password == dto.Password {
		t.Error("want password to be hashed")
	}

	var channel entity.Channel
	err = s.DB().Get(&channel, "SELECT kind, email FROM channels WHERE user_id = $1", user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("want default channel to exist in db")
		}
		t.Fatal(err)
	}
	if channel.Kind != entity.EmailChannel || channel.Email != user.Email {
		t.Errorf(
			"want default channel email of type %v and with email %v, got %v and %v",
			entity.EmailChannel,
			user.Email,
			channel.Kind,
			channel.Email,
		)
	}
}

func TestHandler_SignUp_InvalidInput(t *testing.T) {
	body, err := json.Marshal(auth.SignUpBody{Email: "invalid email", Password: "1"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)

	test.CheckCode(t, http.StatusBadRequest, response.Code)
}

func TestHandler_SignUp_ExistingEmail(t *testing.T) {
	body, err := json.Marshal(auth.SignUpBody{Email: "authtest2@test.com", Password: "123123123"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response = test.ExecuteRequest(s, req)

	test.CheckCode(t, http.StatusConflict, response.Code)
}

func TestHandler_SignIn_ValidCredentials(t *testing.T) {
	body, err := json.Marshal(auth.SignUpBody{Email: "authtest3@test.com", Password: "123123123"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/auth/signin", bytes.NewReader(body))
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	cookie := response.Header().Get("Set-Cookie")
	if !strings.Contains(cookie, "sessionId") {
		t.Error("want sessionId cookie to be set")
	}
}

func TestHandler_SignIn_NotExistedUser(t *testing.T) {
	body, err := json.Marshal(auth.SignInBody{Email: "authnotexisting@test.com", Password: "123123123"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signin", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusNotFound, response.Code)

	cookie := response.Header().Get("Set-Cookie")
	if strings.Contains(cookie, "sessionId") {
		t.Error("want sessionId cookie to be empty")
	}
}

func TestHandler_SignIn_InvalidPassword(t *testing.T) {
	body, err := json.Marshal(auth.SignUpBody{Email: "authtest4@test.com", Password: "123123123"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	body, err = json.Marshal(auth.SignInBody{Email: "authtest4@test.com", Password: "invalid pass"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ = http.NewRequest("PUT", "/v1/auth/signin", bytes.NewReader(body))
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusUnauthorized, response.Code)

	cookie := response.Header().Get("Set-Cookie")
	if strings.Contains(cookie, "sessionId") {
		t.Error("want sessionId cookie to be empty")
	}
}

func TestHandler_SignOut(t *testing.T) {
	body, err := json.Marshal(auth.SignUpBody{Email: "authtest5@test.com", Password: "123123123"})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/auth/signin", bytes.NewReader(body))
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/auth/signout", bytes.NewReader(body))
	req.Header.Set("Cookie", response.Header().Get("Set-Cookie"))
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	cookie := response.Header().Get("Set-Cookie")
	if cookie != "sessionId=; Max-Age=0; HttpOnly" {
		t.Error("want sessionId cookie to be removed")
	}
}

func TestHandler_Check_AuthorizedUser(t *testing.T) {
	dto := auth.SignUpBody{Email: "authtest6@test.com", Password: "123123123"}
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/auth/signup", bytes.NewReader(body))
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/auth/signin", bytes.NewReader(body))
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/v1/auth/check", bytes.NewReader(body))
	req.Header.Set("Cookie", response.Header().Get("Set-Cookie"))
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var user auth.CheckResponse
	err = json.Unmarshal(response.Body.Bytes(), &user)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != dto.Email {
		t.Errorf("want user email %v, got %v", dto.Email, user.Email)
	}
}
