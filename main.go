package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// in main file we will call different handlers and initialize the db.

const hashCost = 8

var db *sql.DB

func main() {

	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/", Homepage)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/stored", Storedb)
	http.HandleFunc("/vote", Vote)
	http.HandleFunc("/terms&conditions", TermsandConditions)
	// initialize database
	initDB()
	log.Fatal(http.ListenAndServe(":8085", nil))
}

func initDB() {
	var err error
	fmt.Println("Initializing the database.......")
	db, err = sql.Open("mysql", "LoginUser:LoginPassword@tcp(127.0.0.1:3306)/User_Login_Database")
	if err != nil {
		panic(err)
	}
	fmt.Println("Database has been initialize successfully")
}
