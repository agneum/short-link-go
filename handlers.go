package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
)

type DefaultResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LinkResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Link    *Link  `json:"link"`
}

func handleRequests(conn *sqlx.DB) {
	router := httprouter.New()

	router.GET("/", home)
	router.GET("/not-found", notFound)
	router.GET("/code/:code", redirect(conn))
	router.POST("/generate-link", generateLink(conn))

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		log.Fatal("Required parameter service port is not set")
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
			link := newLink(generateCode(6), url)
			if err = link.insertNewLink(conn); err != nil {
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
