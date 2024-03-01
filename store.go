package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	password := flag.String("p", "112233", "password")
	flag.Parse()
	dsn := "oakward:" + *password + "@tcp(localhost)/store"

	store, err := openDB(dsn)
	if err != nil {
		log.Fatalln(err)
	}
	store.Close()

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
