package repo

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User entity
type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// Users repository
type Users interface {
	AddUser(name string) (*User, error)
	GetUser(id string) (*User, error)
	ListUsers() map[string]User
}

// UsersMap is in-memory "map" implementation of Users repository
// it stores users as ID => User
// ids is a map between name and User.ID
type UsersMap struct {
	storage map[string]User
	ids     map[string]string
}

// NewUsersMap constructor
func NewUsersMap(storage map[string]User, ids map[string]string) *UsersMap {
	return &UsersMap{
		storage: storage,
		ids:     ids,
	}
}

// AddUser to map
func (r *UsersMap) AddUser(name string) (*User, error) {
	if _, ok := r.ids[name]; ok {
		return nil, fmt.Errorf("Username %v is taken", name)
	}
	// generate Unique ID
	id := uuid.New().String()
	u := User{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}
	r.storage[id] = u
	r.ids[name] = id
	return &u, nil
}

// GetUser by ID
func (r *UsersMap) GetUser(id string) (*User, error) {
	if u, ok := r.storage[id]; ok {
		return &u, nil
	}
	return nil, fmt.Errorf("Can not find user by provided ID")
}

// ListUser returns storage
func (r *UsersMap) ListUsers() map[string]User {
	return r.storage
}
