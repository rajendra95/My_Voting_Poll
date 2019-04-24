package main

import (
	"net/http"
	"database/sql"
	"log"
	_"github.com/go-sql-driver/mysql"
)

// in main file we will call different handlers and initialize the db.

const  hashCost  =8
var db*sql.DB

func main(){
	// signin and signup handlers we will need
	http.HandleFunc("/signup",Signup)
	http.HandleFunc("/signin", Signin)
	// initialize database
	initDB()
	//start the server on 8085 port (considering k8s will run something on 8080)
	log.Fatal(http.ListenAndServe(":8085",nil))
}

func  initDB()  {
	var err error
	//connect to db
	db,err=sql.Open("mysql","dbname =    sslmode=disable")
	if err!=nil{
		panic(err)
	}
}
