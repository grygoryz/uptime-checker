package channel_test

import (
	"bytes"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/channel"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/check"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/test"
	"net/http"
	"strconv"
	"testing"
)

var s *server.Server

func TestMain(m *testing.M) {
	s = server.NewTest()
	s.Init()
	m.Run()
}

func createChannel(t *testing.T, cookie string, dto channel.CreateChannelBody) channel.CreateChannelResponse {
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("POST", "/v1/channels", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	var ch channel.CreateChannelResponse
	err = json.Unmarshal(response.Body.Bytes(), &ch)
	if err != nil {
		t.Fatal(err)
	}

	return ch
}

func TestHandler_CreateChannel_ValidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	dtoEmail := channel.CreateChannelBody{Kind: entity.EmailChannel, Email: "test1@test.com"}
	chEmail := createChannel(t, cookie, dtoEmail)

	dtoWebhook := channel.CreateChannelBody{Kind: entity.WebhookChannel, WebhookURL: "https://test.com"}
	chWebhook := createChannel(t, cookie, dtoWebhook)

	req, _ := http.NewRequest("GET", "/v1/channels", nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var channels []channel.GetChannelsResponseItem
	err := json.Unmarshal(response.Body.Bytes(), &channels)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range channels {
		if c.Id == chEmail.Id && (c.Kind != dtoEmail.Kind || *c.Email != dtoEmail.Email) {
			t.Errorf("want channel of type %v and with email %v to exist", dtoEmail.Kind, dtoEmail.Email)
		}
		if c.Id == chWebhook.Id && (c.Kind != dtoWebhook.Kind || *c.WebhookURL != dtoWebhook.WebhookURL) {
			t.Errorf("want channel of type %v and with webhook url %v to exist", dtoWebhook.Kind, dtoWebhook.WebhookURL)
		}
	}
}

func TestHandler_CreateChannel_InalidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	cases := []struct {
		name string
		dto  channel.CreateChannelBody
	}{
		{
			name: "invalid kind filed",
			dto:  channel.CreateChannelBody{Kind: "invalidkind", Email: "test@test.com"},
		},
		{
			name: "email kind without Email field",
			dto:  channel.CreateChannelBody{Kind: entity.EmailChannel, WebhookURL: "https://example.com"},
		},
		{
			name: "webhook kind without WebhookURL field",
			dto:  channel.CreateChannelBody{Kind: entity.WebhookChannel, Email: "test@test.com"},
		},
		{
			name: "invalid email field",
			dto:  channel.CreateChannelBody{Kind: entity.EmailChannel, Email: "invalidemail"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, err := json.Marshal(c.dto)
			if err != nil {
				t.Fatal(err)
			}

			req, _ := http.NewRequest("POST", "/v1/channels", bytes.NewReader(body))
			req.Header.Set("Cookie", cookie)
			response := test.ExecuteRequest(s, req)
			test.CheckCode(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestHandler_UpdateChannel_ValidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	ch := createChannel(t, cookie, channel.CreateChannelBody{Kind: entity.EmailChannel, Email: "test1@test.com"})

	// update channel
	newDTO := channel.UpdateChannelBody{Kind: entity.WebhookChannel, WebhookURL: "https://test.com"}
	body, err := json.Marshal(newDTO)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/channels/"+strconv.Itoa(ch.Id), bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/v1/channels", nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var channels []channel.GetChannelsResponseItem
	err = json.Unmarshal(response.Body.Bytes(), &channels)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range channels {
		if c.Id == ch.Id && (c.Kind != newDTO.Kind || *c.WebhookURL != newDTO.WebhookURL) {
			t.Errorf("want channel of type %v and with webhook url %v to exist", newDTO.Kind, newDTO.WebhookURL)
		}
	}
}

func TestHandler_UpdateChannel_InvalidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	ch := createChannel(t, cookie, channel.CreateChannelBody{Kind: entity.EmailChannel, Email: "test1@test.com"})

	cases := []struct {
		name string
		dto  channel.UpdateChannelBody
	}{
		{
			name: "invalid kind filed",
			dto:  channel.UpdateChannelBody{Kind: "invalidkind", Email: "test@test.com"},
		},
		{
			name: "email kind without Email field",
			dto:  channel.UpdateChannelBody{Kind: entity.EmailChannel, WebhookURL: "https://example.com"},
		},
		{
			name: "webhook kind without WebhookURL field",
			dto:  channel.UpdateChannelBody{Kind: entity.WebhookChannel, Email: "test@test.com"},
		},
		{
			name: "invalid email field",
			dto:  channel.UpdateChannelBody{Kind: entity.EmailChannel, Email: "invalidemail"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body, err := json.Marshal(c.dto)
			if err != nil {
				t.Fatal(err)
			}

			req, _ := http.NewRequest("PUT", "/v1/channels/"+strconv.Itoa(ch.Id), bytes.NewReader(body))
			req.Header.Set("Cookie", cookie)
			response := test.ExecuteRequest(s, req)
			test.CheckCode(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestHandler_DeleteChannel_WithoutChecks(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	ch := createChannel(t, cookie, channel.CreateChannelBody{Kind: entity.EmailChannel, Email: "test1@test.com"})

	req, _ := http.NewRequest("DELETE", "/v1/channels/"+strconv.Itoa(ch.Id), nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/v1/channels", nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var channels []channel.GetChannelsResponseItem
	err := json.Unmarshal(response.Body.Bytes(), &channels)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range channels {
		if c.Id == ch.Id {
			t.Errorf("want channel to be deleted")
		}
	}
}

func TestHandler_DeleteChannel_WithDependentCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	ch := createChannel(t, cookie, channel.CreateChannelBody{Kind: entity.EmailChannel, Email: "test1@test.com"})

	body, err := json.Marshal(check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{ch.Id},
	})
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("POST", "/v1/checks", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	req, _ = http.NewRequest("DELETE", "/v1/channels/"+strconv.Itoa(ch.Id), nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusBadRequest, response.Code)
}
