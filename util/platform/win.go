//go:build windows
// +build windows

package platform

import (
	"os/user"
)

func HasSetUser() bool {
	return false
}

func HasSetGroup() bool {
	return false
}

func SetUser(username string) error {
	return ErrNotSupported
}

func SetGroup(name string) error {
	return ErrNotSupported
}

func GetCurrentUser() (*user.User, error) {
	return nil, ErrNotSupported
}

func GetCurrentGroup() (*user.Group, error) {
	return nil, ErrNotSupported
}
