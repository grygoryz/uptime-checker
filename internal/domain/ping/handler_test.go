package ping_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	s = server.NewTest()
	s.Init()
	m.Run()
}

func createCheck(t *testing.T, cookie string) string {
	req, _ := http.NewRequest("GET", "/v1/channels", nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var channels []channel.GetChannelsResponseItem
	err := json.Unmarshal(response.Body.Bytes(), &channels)
	if err != nil {
		t.Fatal(err)
	}

	dto := check.CreateCheckBody{
		Name:        "testcheck",
		Description: "some description",
		Interval:    60,
		Grace:       3600,
		Channels:    []int{channels[0].Id},
	}

	body, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}

	req, _ = http.NewRequest("POST", "/v1/checks", bytes.NewReader(body))
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusCreated, response.Code)

	var ch check.CreateCheckResponse
	err = json.Unmarshal(response.Body.Bytes(), &ch)
	if err != nil {
		t.Fatal(err)
	}

	return ch.Id
}

func getCheck(t *testing.T, cookie string, id string) check.Check {
	req, _ := http.NewRequest("GET", "/v1/checks/"+id, nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	var ch check.Check
	err := json.Unmarshal(response.Body.Bytes(), &ch)
	if err != nil {
		t.Fatal(err)
	}

	return ch
}

func getLastPing(t *testing.T, checkId string) entity.Ping {
	var ping entity.Ping
	err := s.DB().Get(&ping, `SELECT id, "type", "date", source, user_agent, duration, body
    FROM pings
	WHERE check_id = $1
	ORDER BY date DESC
	LIMIT 1`, checkId)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("want ping to exist in db")
		}
		t.Fatal(err)
	}

	return ping
}

func getLastFlip(t *testing.T, checkId string) entity.Flip {
	var flip entity.Flip
	err := s.DB().Get(&flip, `SELECT "to", "date"
    FROM flips
	WHERE check_id = $1
	ORDER BY date DESC
	LIMIT 1`, checkId)
	if err != nil {
		if err == sql.ErrNoRows {
			t.Error("want flip to exist in db")
		}
		t.Fatal(err)
	}

	return flip
}

func TestHandler_CreateSuccessPing_NewCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingSuccess,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	p := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, p); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	// verify check
	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckUp {
		t.Errorf("want check Status to be %v, got %v", entity.CheckUp, ch.Status)
	}
	if !ch.NextPing.Equal(ch.LastPing.Add(time.Second * time.Duration(ch.Interval))) {
		t.Errorf(
			"want LastPing + Interval equals to NextPing, got %v. Check: %+v",
			ch.LastPing.Add(time.Duration(ch.Interval)),
			ch,
		)
	}
	if ch.LastStarted != nil {
		t.Errorf("want check LastStarted to be nil, got %v", ch.LastStarted)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipUp {
		t.Errorf("want flip status to be %v, got %v", entity.FlipUp, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_CreateSuccessPing_DownCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id+"/fail", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingSuccess,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	// verify check
	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckUp {
		t.Errorf("want check Status to be %v, got %v", entity.CheckUp, ch.Status)
	}
	if !ch.NextPing.Equal(ch.LastPing.Add(time.Second * time.Duration(ch.Interval))) {
		t.Errorf(
			"want LastPing + Interval equals to NextPing, got %v. Check: %+v",
			ch.LastPing.Add(time.Duration(ch.Interval)),
			ch,
		)
	}
	if ch.LastStarted != nil {
		t.Errorf("want check LastStarted to be nil, got %v", ch.LastStarted)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipUp {
		t.Errorf("want flip status to be %v, got %v", entity.FlipUp, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_CreateSuccessPing_PausedCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	req, _ := http.NewRequest("PUT", "/v1/checks/"+id+"/pause", nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	ua := "ua"
	source := "1.1.1.1"
	req, _ = http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingSuccess,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	// verify check
	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckUp {
		t.Errorf("want check Status to be %v, got %v", entity.CheckUp, ch.Status)
	}
	if !ch.NextPing.Equal(ch.LastPing.Add(time.Second * time.Duration(ch.Interval))) {
		t.Errorf(
			"want LastPing + Interval equals to NextPing, got %v. Check: %+v",
			ch.LastPing.Add(time.Duration(ch.Interval)),
			ch,
		)
	}
	if ch.LastStarted != nil {
		t.Errorf("want check LastStarted to be nil, got %v", ch.LastStarted)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipUp {
		t.Errorf("want flip status to be %v, got %v", entity.FlipUp, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_CreateSuccessPing_StartedCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id+"/start", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
	}
	sample := cmpPing{
		Type:      entity.PingSuccess,
		Source:    source,
		UserAgent: ua,
		Body:      "",
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}
	if ping.Duration == nil {
		t.Error("want check Duration not to be nil")
	}

	// verify check
	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckUp {
		t.Errorf("want check Status to be %v, got %v", entity.CheckUp, ch.Status)
	}
	if !ch.NextPing.Equal(ch.LastPing.Add(time.Second * time.Duration(ch.Interval))) {
		t.Errorf(
			"want LastPing + Interval equals to NextPing, got %v. Check: %+v",
			ch.LastPing.Add(time.Duration(ch.Interval)),
			ch,
		)
	}
	if ch.LastStarted != nil {
		t.Errorf("want check LastStarted to be nil, got %v", ch.LastStarted)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipUp {
		t.Errorf("want flip status to be %v, got %v", entity.FlipUp, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_CreateSuccessPing_UpCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingSuccess,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	// verify check
	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckUp {
		t.Errorf("want check Status to be %v, got %v", entity.CheckUp, ch.Status)
	}
	if !ch.NextPing.Equal(ch.LastPing.Add(time.Second * time.Duration(ch.Interval))) {
		t.Errorf(
			"want LastPing + Interval equals to NextPing, got %v. Check: %+v",
			ch.LastPing.Add(time.Duration(ch.Interval)),
			ch,
		)
	}
	if ch.LastStarted != nil {
		t.Errorf("want check LastStarted to be nil, got %v", ch.LastStarted)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	var count int
	err := s.DB().Get(&count, `SELECT count(*) FROM flips WHERE check_id = $1`, id)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("want flips count to be 1, got %v", count)
	}
}

func TestHandler_FailPing_NewCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id+"/fail", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingFail,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckDown {
		t.Errorf("want check Status to be %v, got %v", entity.CheckDown, ch.Status)
	}
	if ch.LastStarted != nil || ch.NextPing != nil {
		t.Errorf("want check LastStarted and NextPing to be nil, got %v and %v", ch.LastStarted, ch.NextPing)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipDown {
		t.Errorf("want flip status to be %v, got %v", entity.FlipDown, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_FailPing_UpCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id, nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/pings/"+id+"/fail", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingFail,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckDown {
		t.Errorf("want check Status to be %v, got %v", entity.CheckDown, ch.Status)
	}
	if ch.LastStarted != nil || ch.NextPing != nil {
		t.Errorf("want check LastStarted and NextPing to be nil, got %v and %v", ch.LastStarted, ch.NextPing)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipDown {
		t.Errorf("want flip status to be %v, got %v", entity.FlipDown, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_FailPing_PausedCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	req, _ := http.NewRequest("PUT", "/v1/checks/"+id+"/pause", nil)
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	ua := "ua"
	source := "1.1.1.1"
	req, _ = http.NewRequest("PUT", "/v1/pings/"+id+"/fail", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingFail,
		Source:    source,
		UserAgent: ua,
		Body:      "",
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckDown {
		t.Errorf("want check Status to be %v, got %v", entity.CheckDown, ch.Status)
	}
	if ch.LastStarted != nil || ch.NextPing != nil {
		t.Errorf("want check LastStarted and NextPing to be nil, got %v and %v", ch.LastStarted, ch.NextPing)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipDown {
		t.Errorf("want flip status to be %v, got %v", entity.FlipDown, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_FailPing_StartedCheck(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id+"/start", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("PUT", "/v1/pings/"+id+"/fail", nil)
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response = test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
	}
	sample := cmpPing{
		Type:      entity.PingFail,
		Source:    source,
		UserAgent: ua,
		Body:      "",
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}
	if ping.Duration == nil {
		t.Error("want check Duration not to be nil")
	}

	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckDown {
		t.Errorf("want check Status to be %v, got %v", entity.CheckDown, ch.Status)
	}
	if ch.LastStarted != nil || ch.NextPing != nil {
		t.Errorf("want check LastStarted and NextPing to be nil, got %v and %v", ch.LastStarted, ch.NextPing)
	}
	if !ch.LastPing.Equal(ping.Date) {
		t.Errorf("want check LastDate to be %v, got %v", ping.Date, ch.LastPing)
	}

	// verify flip
	f := getLastFlip(t, id)
	if f.To != entity.FlipDown {
		t.Errorf("want flip status to be %v, got %v", entity.FlipDown, f.To)
	}
	if !ch.LastPing.Equal(f.Date) {
		t.Errorf("want flip Date to be %v, got %v", ch.LastPing, f.Date)
	}
}

func TestHandler_StartPing(t *testing.T) {
	cookie, _ := test.Authorize(t, s)
	id := createCheck(t, cookie)

	ua := "ua"
	source := "1.1.1.1"
	body := "some body"
	req, _ := http.NewRequest("PUT", "/v1/pings/"+id+"/start", bytes.NewReader([]byte(body)))
	req.Header.Set("User-Agent", ua)
	req.RemoteAddr = source
	req.Header.Set("Cookie", cookie)
	response := test.ExecuteRequest(s, req)
	test.CheckCode(t, http.StatusOK, response.Code)

	// verify ping
	ping := getLastPing(t, id)
	type cmpPing struct {
		Type      entity.PingKind
		Source    string
		UserAgent string
		Body      string
		Duration  *int
	}
	sample := cmpPing{
		Type:      entity.PingStart,
		Source:    source,
		UserAgent: ua,
		Body:      body,
		Duration:  nil,
	}
	actual := cmpPing{
		Type:      ping.Type,
		Source:    ping.Source,
		UserAgent: ping.UserAgent,
		Body:      ping.Body,
		Duration:  ping.Duration,
	}
	if diff := cmp.Diff(sample, actual); diff != "" {
		t.Errorf("want ping values to be the same, diff (-want, +got)\n: %s", diff)
	}

	ch := getCheck(t, cookie, id)
	if ch.Status != entity.CheckStarted {
		t.Errorf("want check Status to be %v, got %v", entity.CheckStarted, ch.Status)
	}
	if ch.LastStarted == nil {
		t.Error("want check LastStarted not to be nil")
	}
	if !ch.LastStarted.Equal(ping.Date) {
		t.Errorf("want check LastStarted to be %v, got %v", ping.Date, ch.LastStarted)
	}
}
