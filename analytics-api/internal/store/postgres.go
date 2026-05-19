package store

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"` // Never serialize
	Name     string `json:"name"`
	Role     string `json:"role"`
}

func NewPostgres(host, port, user, password, dbname string) (*Postgres, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	return &Postgres{db: db}, nil
}

func (pg *Postgres) FindUserByEmail(email string) (*User, error) {
	var u User
	err := pg.db.QueryRow(
		"SELECT id, email, password, name, role FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Name, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (pg *Postgres) Close() error {
	return pg.db.Close()
}
