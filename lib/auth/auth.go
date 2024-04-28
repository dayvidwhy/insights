package auth

import "database/sql"

type AuthStore struct {
	db *sql.DB
}

type AuthHandler struct {
	store *AuthStore
}

func SetupAuth(db *sql.DB) *AuthStore {
	return &AuthStore{db: db}
}

func NewAuth(store *AuthStore) *AuthHandler {
	return &AuthHandler{store: store}
}
