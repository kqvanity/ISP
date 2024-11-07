package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName = "vodafone"
)

// // Login checks if the user credentials are valid
// func (u *User) Login() error {
// 	if u.Username == "" || u.Password == "" {
// 		return fmt.Errorf("username and password cannot be empty")
// 	}
// 	return nil
// }
//  (If credentials are valid & stored cookies expired)

// Common errors that can occur in the application
var (
	ErrEmptyCredentials = fmt.Errorf("username or password cannot be empty")
	ErrInvalidPassword  = fmt.Errorf("invalid password")
	ErrKeyringSave      = fmt.Errorf("error saving the password from keyring")
	ErrKeyringRetrieve  = fmt.Errorf("error retrieving the password from keyring")
)

// GetCredentials retrieves the stored credentials from the system keyring
func (v *Vault) GetCredentials(key string) (string, error) {
	value, err := keyring.Get("vodafone", key)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", ErrEmptyCredentials
	}
	return value, nil
}

// SaveCredentials stores the user credentials securely in the system keyring
func (v *Vault) SaveCredentials(key, value string) error {
	// if err := u.Login(); err != nil { return err }
	err := keyring.Set("vodafone", key, value)
	if err != nil {
		return ErrKeyringSave
	}
	return nil
}

func ValidateCredentials(username string, password string) error {
	if username == "" || password == "" {
		return ErrEmptyCredentials
	}
	if _, err := strconv.ParseInt(username, 10, 64); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

type Vault struct {
	Password string
	Token    string
}

// IsCredentialsError checks if the error is related to credentials
func (v *Vault) IsCredentialsError(err error) bool {
	return errors.Is(err, ErrEmptyCredentials) ||
		errors.Is(err, ErrInvalidPassword)
}

func (v *Vault) isKeyringError(err error) bool {
	return errors.Is(err, ErrKeyringSave) ||
		errors.Is(err, ErrKeyringRetrieve)
}

func (v *Vault) IsKeyringError(err error) bool {
	return err.Error() == "The name is not activatable"
}
