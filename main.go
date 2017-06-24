package main

import (
	"database/sql"
	"encoding/json"
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
  url LONGTEXT NOT NULL,
  created_at DATETIME,
  PRIMARY KEY (id), UNIQUE KEY UNIQ_Code (code)
);`

var log = logrus.New()

type Link struct {
	ID        int    `db:"id" json:"-"`
	Code      string `db:"code" json:"code"`
	Url       string `db:"url" json:"url"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

type DefaultResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Link    *Link  `json:"link"`
}

func newLink(code, url string) *Link {
	l := &Link{
		Code:      code,
		Url:       url,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	return l
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

		http.Redirect(w, r, redirectRecord.Url, 302)
	}
}

func generateLink(conn *sqlx.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		var link Link
		w.Header().Set("Content-Type", "application/json")

		err := r.ParseForm()
		if err != nil {
			log.Error(err)
			err = json.NewEncoder(w).Encode(
				&DefaultResponse{false, "Parsing error"},
			)
			if err != nil {
				log.Error(err)
			}
			return
		}

		url := r.PostFormValue("url")
		if url == "" {
			err = json.NewEncoder(w).Encode(
				&DefaultResponse{false, "Url is required"},
			)
			if err != nil {
				log.Error(err)
			}
			return
		}

		err = conn.Get(&link, "SELECT * FROM links WHERE url = ?", url)

		if err == sql.ErrNoRows {
			link, err := insertNewLink(conn, url)
			if err != nil {
				err = json.NewEncoder(w).Encode(
					&DefaultResponse{false, "Insert error"},
				)
				if err != nil {
					log.Error(err)
				}
			}

			err = json.NewEncoder(w).Encode(
				&LinkResponse{true, "Short link has been created", link},
			)
			if err != nil {
				log.Error(err)
			}
			return
		} else if err != nil {
			err = json.NewEncoder(w).Encode(
				&DefaultResponse{false, "Database error"},
			)
			if err != nil {
				log.Error(err)
			}
			return
		}

		err = json.NewEncoder(w).Encode(
			&LinkResponse{true, "Short link already exists", &link},
		)
		if err != nil {
			log.Error(err)
		}
	}
}

func insertNewLink(conn *sqlx.DB, url string) (*Link, error) {
	code := generateCode(6)
	l := newLink(code, url)

	_, err := conn.Exec("INSERT INTO links (code, url, created_at) VALUES (?, ?, ?)", l.Code, l.Url, l.CreatedAt)
	if err != nil {
		return nil, err
	}

	return l, nil
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
