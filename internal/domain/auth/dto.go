package auth

type CreateUserBody struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=20"`
}

type CreateUserParams struct {
	FirstName string `json:"firstName" validate:"min=20"`
}
