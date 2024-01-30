package main

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	ID        string
	Name      string
	Password  string
	FirstName string
	LastName  string
	Role      string
	Age       int
}

type UserService interface {
	CreateUser(createUserBody *CreateUserBody) (*User, error)
	GetUser(id string) (*User, error)
	ListUsers() ([]User, error)
	DeleteUser(id string) error
	UpdateUser(id string, updateUserBody *CreateUserBody) (*User, error)
	ValidateCredentials(loginUser *LoginUserBody) (*User, error)
}

type userService struct {
	users []User
}

// NewUserService returns a new user service.
func NewUserService() UserService {
	return &userService{
		users: []User{
			{ID: "embedded1", Name: "John", Password: "VMware1!"},
			{ID: "embedded2", Name: "Jane", Password: "VMware1!"},
		},
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(createUserBody *CreateUserBody) (*User, error) {
	newUser := User{
		ID:        uuid.NewString(),
		Name:      createUserBody.Name,
		Password:  createUserBody.Password,
		FirstName: createUserBody.FirstName,
		LastName:  createUserBody.LastName,
		Role:      createUserBody.Role,
		Age:       createUserBody.Age,
	}
	s.users = append(s.users, newUser)
	return &newUser, nil
}

// GetUser returns a user by ID
func (s *userService) GetUser(id string) (*User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("User not found")
}

// ListUsers returns a list of all users
func (s *userService) ListUsers() ([]User, error) {
	return s.users, nil
}

// DeleteUser deletes a user by ID
func (s *userService) DeleteUser(id string) error {
	for i, u := range s.users {
		if u.ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return nil
		}
	}
	return errors.New("User not found")
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(id string, updateUserBody *CreateUserBody) (*User, error) {
	for i, u := range s.users {
		if u.ID == id {
			// update the user properties
			s.users[i].Name = updateUserBody.Name
			s.users[i].FirstName = updateUserBody.FirstName
			s.users[i].LastName = updateUserBody.LastName
			s.users[i].Password = updateUserBody.Password
			s.users[i].Role = updateUserBody.Role
			s.users[i].Age = updateUserBody.Age
			return &s.users[i], nil
		}
	}
	return nil, errors.New("user not found")
}

// ValidateCredentials validates the given user credentials
func (s *userService) ValidateCredentials(loginUser *LoginUserBody) (*User, error) {
	for _, user := range s.users {
		if user.Name == loginUser.Name && user.Password == loginUser.Password {
			return &user, nil
		}
	}
	return nil, errors.New("invalid credentials")
}
