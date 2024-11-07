package main

// TODO: move to another file.
// User struct represents a user with login credentials
type User struct {
	Phonenumber string
}

// NewUser creates a new User interface with the provided credentials
func NewUser(username string, password int64) *User {
	return &User{
		Phonenumber: username,
		// Password:    fmt.Sprintf("%011d", password),
	}
}
