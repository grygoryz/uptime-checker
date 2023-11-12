package ping

type CheckIdParam struct {
	Id string `json:"id" validate:"uuid4"`
}
