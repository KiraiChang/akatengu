package request

import "errors"

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (l *Login) Validate() error {
	if l.Username == "" {
		return errors.New("username is empty")
	}
	if l.Password == "" {
		return errors.New("password is empty")
	}
	return nil
}

type CreateUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (l *CreateUser) Validate() error {
	if l.Username == "" {
		return errors.New("username is empty")
	}
	if l.Password == "" {
		return errors.New("password is empty")
	}
	return nil
}
