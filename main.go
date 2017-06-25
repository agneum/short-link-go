package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
)

type DatabaseSpecification struct {
	Host     string `default:"localhost"`
	Port     int    `default:"3306"`
	User     string `default:"root"`
	Password string `default:""`
	Name     string `default:"short_links"`
}

var linkTable = `
CREATE TABLE IF NOT EXISTS links (
  id int(10) unsigned NOT NULL AUTO_INCREMENT,
  code varchar(6) NOT NULL,
  url LONGTEXT NOT NULL,
  created_at DATETIME,
  PRIMARY KEY (id), UNIQUE KEY UNIQ_Code (code)
);`

var log = logrus.New()

func main() {
	var s DatabaseSpecification
	err := envconfig.Process("database", &s)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", s.User, s.Password, s.Host, s.Port, s.Name))
	if err != nil {
		log.Fatal(err)
	}
	conn.MustExec(linkTable)

	handleRequests(conn)
}
