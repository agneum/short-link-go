package main

import (
	"testing"

	"github.com/jmoiron/sqlx"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestShouldCreateShortLink(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expecting", err)
	}

	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	link := newLink("abcdef", "http://golang.org")

	mock.
		ExpectExec("INSERT INTO links").
		WithArgs(link.Code, link.Url, link.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err = link.insertNewLink(sqlxDB); err != nil {
		t.Errorf("Error '%s' was not expected", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expections: %s", err)
	}
}
