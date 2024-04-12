package access

type Access int

const (
	User Access = iota
	Admin
	NotAccess
)

func GetAccess(token string) Access {
	if token == "user_token" {
		return User
	}
	if token == "admin_token" {
		return Admin
	}
	return NotAccess
}
