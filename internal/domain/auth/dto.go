package auth

type SignUpBody struct {
	Email    string `json:"email" validate:"required,max=320,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type SignInBody struct {
	Email    string `json:"email" validate:"required,max=320,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type CheckResponse struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}
