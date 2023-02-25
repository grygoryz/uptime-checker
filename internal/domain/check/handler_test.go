package check_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gitlab.com/grygoryz/uptime-checker/config"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/channel"
	"gitlab.com/grygoryz/uptime-checker/internal/domain/check"
	"gitlab.com/grygoryz/uptime-checker/internal/entity"
	"gitlab.com/grygoryz/uptime-checker/internal/server"
	"gitlab.com/grygoryz/uptime-checker/internal/utility/test"
	"net/http"
	"testing"
	"time"
)

var s *server.Server

func TestMain(m *testing.M) {
	cfg := config.New(true)
	s = server.New(cfg)
	s.Init()
	m.Run()
}

func getChannels(t *testing.T, cookie string) []channel.GetChannelsResponseItem {
	req, _ := http.NewRequest("GET", "/v1/channels", nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var channels []channel.GetChannelsResponseItem
	err := json.Unmarshal(response.Body.Bytes(), &channels)
	if err != nil {
		t.Fatal(err)
	}

	return channels
}

func createCheck(t *testing.T, cookie string, dto check.CreateCheckBody) check.CreateCheckResponse {
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("POST", "/v1/checks", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	var ch check.CreateCheckResponse
	err = json.Unmarshal(response.Body.Bytes(), &ch)
	if err != nil {
		t.Fatal(err)
	}

	return ch
}

func TestHandler_CreateCheck_ValidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	channels := getChannels(t, cookie)

	dto := check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	}
	ch := createCheck(t, cookie, dto)

	req, _ := http.NewRequest("GET", "/v1/checks/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var found check.Check
	err := json.Unmarshal(response.Body.Bytes(), &found)
	if err != nil {
		t.Fatal(err)
	}

	// compare
	type cmpS struct {
		Name        string
		Description string
		Interval    int
		Grace       int
		LastPing    *time.Time
		NextPing    *time.Time
		LastStarted *time.Time
		Status      entity.CheckStatus
		Channels    []int
	}
	sample := cmpS{
		Name:        dto.Name,
		Description: dto.Description,
		Interval:    dto.Interval,
		Grace:       dto.Grace,
		LastPing:    nil,
		NextPing:    nil,
		LastStarted: nil,
		Status:      entity.CheckNew,
		Channels:    dto.Channels,
	}
	actual := cmpS{
		Name:        found.Name,
		Description: found.Description,
		Interval:    found.Interval,
		Grace:       found.Grace,
		LastPing:    found.LastPing,
		NextPing:    found.NextPing,
		LastStarted: found.LastStarted,
		Status:      found.Status,
	}
	for _, c := range found.Channels {
		actual.Channels = append(actual.Channels, c.Id)
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want check values to be the same, diff (-want, +got)\n: %s", diff)
	}
}

func TestHandler_CreateCheck_InvalidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	dto := check.CreateCheckBody{
		Name:        "",
		Description: "",
		Interval:    0,
		Grace:       31536001,
		Channels:    []int{},
	}
	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("POST", "/v1/checks", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)

	test.CheckCode(t, http.StatusBadRequest, response.Code)
}

func TestHandler_UpdateCheck_ValidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	channels := getChannels(t, cookie)

	ch := createCheck(t, cookie, check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	})

	newDto := check.UpdateCheckBody{
		CreateCheckBody: check.CreateCheckBody{
			Name:        "testcheck2",
			Description: "some description2",
			Interval:    61,
			Grace:       3601,
			Channels:    []int{channels[0].Id},
		},
	}
	body, err := json.Marshal(newDto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", "/v1/checks/"+ch.Id, bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/v1/checks/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var found check.Check
	err = json.Unmarshal(response.Body.Bytes(), &found)
	if err != nil {
		t.Fatal(err)
	}

	// compare
	type cmpS struct {
		Name        string
		Description string
		Interval    int
		Grace       int
		LastPing    *time.Time
		NextPing    *time.Time
		LastStarted *time.Time
		Status      entity.CheckStatus
		Channels    []int
	}
	sample := cmpS{
		Name:        newDto.Name,
		Description: newDto.Description,
		Interval:    newDto.Interval,
		Grace:       newDto.Grace,
		LastPing:    nil,
		NextPing:    nil,
		LastStarted: nil,
		Status:      entity.CheckNew,
		Channels:    newDto.Channels,
	}
	actual := cmpS{
		Name:        found.Name,
		Description: found.Description,
		Interval:    found.Interval,
		Grace:       found.Grace,
		LastPing:    found.LastPing,
		NextPing:    found.NextPing,
		LastStarted: found.LastStarted,
		Status:      found.Status,
	}
	for _, c := range found.Channels {
		actual.Channels = append(actual.Channels, c.Id)
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want check values to be the same, diff (-want, +got)\n: %s", diff)
	}
}

func TestHandler_UpdateCheck_InvalidInput(t *testing.T) {
	cookie, _ := test.Authorize(t, s)

	newDto := check.UpdateCheckBody{
		CreateCheckBody: check.CreateCheckBody{
			Name:        "",
			Description: "",
			Interval:    0,
			Grace:       0,
			Channels:    []int{},
		},
	}
	body, err := json.Marshal(newDto)
	if err != nil {
		t.Fatal(err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("PUT", "/v1/checks/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusBadRequest, response.Code)
}

func TestHandler_DeleteCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	channels := getChannels(t, cookie)

	ch := createCheck(t, cookie, check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	})

	req, _ := http.NewRequest("DELETE", "/v1/checks/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/v1/checks/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusNotFound, response.Code)
}

func TestHandler_PauseCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	channels := getChannels(t, cookie)

	dto := check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	}
	ch := createCheck(t, cookie, dto)

	req, _ := http.NewRequest("PUT", "/v1/checks/"+ch.Id+"/pause", nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/v1/checks/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var found check.Check
	err := json.Unmarshal(response.Body.Bytes(), &found)
	if err != nil {
		t.Fatal(err)
	}

	// compare
	type cmpS struct {
		Name        string
		Description string
		Interval    int
		Grace       int
		LastPing    *time.Time
		NextPing    *time.Time
		LastStarted *time.Time
		Status      entity.CheckStatus
		Channels    []int
	}
	sample := cmpS{
		Name:        dto.Name,
		Description: dto.Description,
		Interval:    dto.Interval,
		Grace:       dto.Grace,
		LastPing:    nil,
		NextPing:    nil,
		LastStarted: nil,
		Status:      entity.CheckPaused,
		Channels:    dto.Channels,
	}
	actual := cmpS{
		Name:        found.Name,
		Description: found.Description,
		Interval:    found.Interval,
		Grace:       found.Grace,
		LastPing:    found.LastPing,
		NextPing:    found.NextPing,
		LastStarted: found.LastStarted,
		Status:      found.Status,
	}
	for _, c := range found.Channels {
		actual.Channels = append(actual.Channels, c.Id)
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want check values to be the same, diff (-want, +got)\n: %s", diff)
	}

	var flip int
	err = s.DB().Get(&flip, `SELECT count(*) FROM flips WHERE check_id = $1 AND "to" = $2`, ch.Id, entity.FlipPaused)
	if flip != 1 {
		t.Errorf("want flip to %v to be created", entity.FlipPaused)
	}
}

func TestHandler_GetPings(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	channels := getChannels(t, cookie)

	dto := check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	}
	ch := createCheck(t, cookie, dto)

	now := time.Now().UnixMilli()

	req, _ := http.NewRequest("PUT", "/v1/pings/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	url := fmt.Sprintf("/v1/checks/%v/pings?limit=10&offset=0&from=%v&to=%v", ch.Id, now, time.Now().UnixMilli())
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var pings check.GetPingsResponse
	err := json.Unmarshal(response.Body.Bytes(), &pings)
	if err != nil {
		t.Fatal(err)
	}

	if len(pings.Items) != 2 {
		t.Errorf("want 2 pings, got %v", len(pings.Items))
	}
}

func TestHandler_GetFlips(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	channels := getChannels(t, cookie)

	dto := check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	}
	ch := createCheck(t, cookie, dto)

	now := time.Now().UnixMilli()

	req, _ := http.NewRequest("PUT", "/v1/pings/"+ch.Id, nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/pings/"+ch.Id+"/fail", nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	url := fmt.Sprintf("/v1/checks/%v/flips?limit=10&offset=0&from=%v&to=%v", ch.Id, now, time.Now().UnixMilli())
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var flips check.GetFlipsResponse
	err := json.Unmarshal(response.Body.Bytes(), &flips)
	if err != nil {
		t.Fatal(err)
	}

	if len(flips.Items) != 2 {
		t.Errorf("want 2 flips, got %v", len(flips.Items))
	}
}
