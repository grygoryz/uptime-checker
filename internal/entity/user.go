package entity

// TODO: maybe create UserSecure
type User struct {
	Id       int    `db:"id"`
	Email    string `db:"email"`
	Password string `db:"password"`
}
