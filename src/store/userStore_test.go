package store

import (
	"context"
	"testing"

	"github.com/bottlepay/portfolio-data/model"
)

func TestFakeUserStore(t *testing.T) {

	store := NewFakeUserStore()

	if len(store.usersMap) != 0 {
		t.Error("Store isn't empty after creation")
	}

	if _, err := store.GetUser(context.Background(), 1); err != UserNotFoundError {
		t.Error("GetUser should fail with UserNotFoundError")
	}

	store.Populate()

	if len(store.usersMap) != 1 {
		t.Error("Store Populate() didn't work")
	}

	anotherUser := model.NewUser(2)
	error := store.AddUser(context.Background(), anotherUser)
	if error != nil || len(store.usersMap) != 2 {
		t.Error("Store AddUser for second user failed")
	}

	anotherUser2 := model.NewUser(2)
	retryerror := store.AddUser(context.Background(), anotherUser2)
	if retryerror != UserAlreadyExistsError {
		t.Error("Store AddUser should prevent duplicates by ID")
	}
}
