package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/bottlepay/portfolio-data/model"
)

var (
	UserNotFoundError      = fmt.Errorf("User not found")
	UserAlreadyExistsError = fmt.Errorf("User exists already")
)

// UserStore is used to store and get users
// context and error will be needed for a real world implementation
type UserStore interface {
	// GetUser returns a model.User corresponding to the int32 ID
	GetUser(context.Context, int32) (*model.User, error)
	// AddUser adds a model.User to the store
	AddUser(context.Context, *model.User) error
}

// FakeUserStore stores users in memory
type FakeUserStore struct {
	usersMap map[int32]*model.User

	l sync.RWMutex
}

func NewFakeUserStore() *FakeUserStore {
	s := &FakeUserStore{
		make(map[int32]*model.User),
		sync.RWMutex{},
	}
	return s
}

// Adds the fake User to the store
func (s *FakeUserStore) Populate() {
	fakeuser := model.NewUser(1)
	fakeuser.Custodians = []int32{1, 2, 3, 4}
	s.AddUser(context.Background(), fakeuser)
}

func (s *FakeUserStore) GetUser(context context.Context, id int32) (*model.User, error) {
	s.l.RLock()
	defer s.l.RUnlock()

	if u, found := s.usersMap[id]; found {
		return u, nil
	}
	return nil, UserNotFoundError
}

func (s *FakeUserStore) AddUser(context context.Context, user *model.User) error {
	s.l.Lock()
	defer s.l.Unlock()

	if _, exists := s.usersMap[user.ID]; exists {
		return UserAlreadyExistsError
	}
	s.usersMap[user.ID] = user
	return nil
}

func init() {
	// Check interface implementation
	var _ UserStore = (*FakeUserStore)(nil)
}
