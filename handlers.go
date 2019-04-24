package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var err error

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" *** Signup handler got called *****")
	if r.Method != "POST" {
		http.ServeFile(w, r, "signup.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println(username)
	fmt.Println(password)

	var user string

	rows, err := db.Query("SELECT username FROM users WHERE username=?", username)
	if err != nil {
		panic(err.Error())
	}
	count := 0
	for rows.Next() {
		count++
	}
	if count > 0 {
		w.Write([]byte("USER already existed"))
		return
	}
	err = db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)
	fmt.Println("Dabatase error is ", err)
	switch {
	case err == sql.ErrNoRows:

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Unable to crate accout right now (ONE)", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("INSERT INTO users(username,password) VALUES(?,?)", username, hashedPassword)
		if err != nil {
			http.Error(w, "Unable to create your account (TWO)", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("User Accout Created"))

	case err != nil:
		http.Error(w, "Unable to create your account (THREE)", http.StatusInternalServerError)
		return
	default:
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	fmt.Println("Redirecting.......")
	http.Redirect(w, r, "/", http.StatusOK)
}
func Login(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("***** LOGIN HANDLER *******")

	//.Println("r.Method before IF", r.Method)
	if r.Method != "POST" {
		http.ServeFile(w, r, "login.html")
		return
	}
	//fmt.Println("r.Method after IF", r.Method)
	username := r.FormValue("username")
	password := r.FormValue("password")

	var dbusername string
	var dbpassword string

	//fmt.Println(username)
	//fmt.Println(password)

	err = db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&dbusername, &dbpassword)
	if err != nil {
		fmt.Print("FIRST")
		w.Write([]byte("User Does Not Exist"))
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(password))
	if err != nil {
		fmt.Println("SECOND")
		w.Write([]byte("Credentials Did not Match!"))
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	//w.Write([]byte("Hello!" + dbusername))
	http.Redirect(w, r, "/register", 301)
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
func Register(w http.ResponseWriter, r *http.Request) {
	/*

		fmt.Println("Register handler Called!")
		fmt.Println("r.Method before IF LOOP", r.Method)
		if r.Method != "POST" {
			fmt.Println("r.Method in IF LOOP", r.Method)
			http.ServeFile(w, r, "register.html")
			return
		}
		//http.ServeFile(w, r, "register.html")

		VoterID := r.FormValue("VoterId")
		LastName := r.FormValue("lastname")
		FirstName := r.FormValue("FirstName")
		State := r.FormValue("State")
		City := r.FormValue("City")
		Age := r.FormValue("Age")
		Sex := r.FormValue("Sex")

		//fmt.Println("formvalue for LastName", r.FormValue(LastName))
		_, err = db.Exec("INSERT INTO Voters(VoterID,LastName,FirstName,Age,Sex,State,City)VALUES(?,?,?,?,?,?,?)", VoterID, LastName, FirstName, Age, Sex, State, City)
		fmt.Println("db.EXEC querry executed")
		if err != nil {
			fmt.Println("Error in registering", err)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		w.Write([]byte(" Registered Successully!"))
	*/

	//fmt.Println("***", r.Method)
	if r.Method != "POST" {
		http.ServeFile(w, r, "register.html")
	}
	//fmt.Println("*** came out of IF LOOP")
	http.Redirect(w, r, "/stored", 301)

}

func Storedb(w http.ResponseWriter, r *http.Request) {
	fmt.Println("r.method", r.Method)
	fmt.Println("r.URl", r.URL)
	var VoterID string
	var LastName string
	var FirstName string
	var Age string
	var Sex string
	var State string
	var City string

	switch r.Method {

	case "GET":
		fmt.Println("GET METHOD")
		http.ServeFile(w, r, "db.html")
	case "POST":
		fmt.Println("POST METHOD")
		// call ParseForm to parse the raw querry
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from Website! r.PostForm = %v \n", r.PostForm)

		VoterID = r.FormValue("VoterID")
		LastName = r.FormValue("LastName")
		FirstName = r.FormValue("FirstName")
		State = r.FormValue("State")
		City = r.FormValue("City")
		Age = r.FormValue("Age")
		Sex = r.FormValue("Sex")

	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}

	_, err = db.Exec("INSERT INTO Voters(VoterID,LastName,FirstName,Age,Sex,State,City)VALUES(?,?,?,?,?,?,?)", VoterID, LastName, FirstName, Age, Sex, State, City)
	//	_, err = statement.Exec(VoterID, LastName, FirstName, Age, Sex, State, City)
	if err != nil {
		panic(err.Error())
	}

	w.Write([]byte(" Registered Successully!"))
}
