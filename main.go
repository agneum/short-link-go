package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
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
	router.GET("/code/:code", redir(conn))
	router.POST("/generate-code", generateCode(conn))

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

func redir(conn *sqlx.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var redirectRecord Link

		code := p.ByName("code")
		err := conn.Get(&redirectRecord, "SELECT * FROM links WHERE code = ?", code)

		if err == sql.ErrNoRows {
			log.Printf("Code '%v' not found\n", code)
			http.Redirect(w, r, "/not-found", 302)
		} else if err != nil {
			log.Printf("%+v\n", err)
			http.Redirect(w, r, "/not-found", 302)
		}
		http.Redirect(w, r, redirectRecord.Link, 302)

	}
}

func generateCode(conn *sqlx.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	}
}
