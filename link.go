package main

import (
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Link struct {
	ID        int    `db:"id" json:"-"`
	Code      string `db:"code" json:"code"`
	Url       string `db:"url" json:"url"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

func newLink(code, url string) *Link {
	l := &Link{
		Code:      code,
		Url:       url,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	return l
}

func (l *Link) insertNewLink(conn *sqlx.DB) error {
	_, err := conn.Exec("INSERT INTO links (code, url, created_at) VALUES (?, ?, ?)", l.Code, l.Url, l.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func generateCode(n int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, n)
	lettersLen := len(letters)

	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}

	return string(b)
}
