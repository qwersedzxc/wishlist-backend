package request

// UserRegisterRequest - запрос на регистрацию пользователя
type UserRegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserLoginRequest - запрос на вход пользователя
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
