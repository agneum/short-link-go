package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/pressly/chi"
)

var linkTable = `
CREATE TABLE IF NOT EXISTS links (
  id int(10) unsigned NOT NULL AUTO_INCREMENT,
  code varchar(6) NOT NULL,
  link LONGTEXT NOT NULL,
  created_at DATETIME,
  PRIMARY KEY (id), UNIQUE KEY UNIQ_Code (code)
);`

type Link struct {
	ID        int    `db:"id"`
	Code      string `db:"code"`
	Link      string `db:"link"`
	CreatedAt string `db:"created_at"`
}

func main() {
	conn, err := sqlx.Connect("mysql", "root@tcp(localhost:3306)/short_links")
	if err != nil {
		log.Fatal(err)
	}
	conn.MustExec(linkTable)

	r := chi.NewRouter()
	r.Get("/", home)
	r.Get("/", notFound)
	r.Get("/code/:code", redirect)
	r.Post("/generate-code", generateCode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	http.ListenAndServe(":"+port, r)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Link not-found"))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	w.Write([]byte(fmt.Sprintf("Code: %v", code)))
}

func generateCode(w http.ResponseWriter, r *http.Request) {

}
