package main

import (
	"database/sql"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var linkTable = `
CREATE TABLE IF NOT EXISTS links (
  id int(10) unsigned NOT NULL AUTO_INCREMENT,
  code varchar(6) NOT NULL,
  link LONGTEXT NOT NULL,
  created_at DATETIME,
  PRIMARY KEY (id), UNIQUE KEY UNIQ_Code (code)
);`

var log = logrus.New()

type Link struct {
	ID        int     `db:"id"`
	Code      string  `db:"code"`
	Link      string  `db:"link"`
	CreatedAt *string `db:"created_at"`
}

func main() {
	conn, err := sqlx.Connect("mysql", "root@tcp(localhost:3306)/short_links")
	if err != nil {
		log.Fatal(err)
	}
	conn.MustExec(linkTable)

	handleRequests(conn)
}

func handleRequests(conn *sqlx.DB) {
	router := httprouter.New()

	router.GET("/", home)
	router.GET("/not-found", notFound)
	router.GET("/code/:code", redirect(conn))
	router.POST("/generate-link", generateLink(conn))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	http.ListenAndServe(":"+port, router)
}

func home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hello"))
}

func notFound(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Link not-found"))
}

func redirect(conn *sqlx.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var redirectRecord Link

		code := p.ByName("code")
		err := conn.Get(&redirectRecord, "SELECT * FROM links WHERE code = ?", code)

		if err != nil {
			log.Printf("Code '%v' not found\n", code)
			http.Redirect(w, r, "/not-found", 302)
			return
		}

		http.Redirect(w, r, redirectRecord.Link, 302)
	}
}

func generateLink(conn *sqlx.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var link Link

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		url := r.PostFormValue("url")

		if url == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Url is required"))
			return
		}

		err = conn.Get(&link, "SELECT * FROM links WHERE link = ?", url)

		if err == sql.ErrNoRows {
			code, err := insertNewLink(conn, url)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(r.Host + "/code/" + code))
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(r.Host + "/code/" + link.Code))
	}
}

func insertNewLink(conn *sqlx.DB, url string) (string, error) {
	datetime := time.Now().Format(time.RFC3339)
	code := generateCode(6)

	_, err := conn.Exec("INSERT INTO links (code, link, created_at) VALUES (?, ?, ?)", code, url, datetime)
	if err != nil {
		return "", err
	}

	return code, nil
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
