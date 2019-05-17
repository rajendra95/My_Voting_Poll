package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"
	//"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var err error
//Session name to store session under
const SessionName  ="Voting-Session"

var (
	sessionStore sessions.Store
	log = logrus.WithField("cmd",SessionName)
)

// to handle the error if anything goes wrong
func handleSessionError(w http.ResponseWriter, err error){
	log.WithField("err",err).Info("Error handling session")
	http.Error(w,"Application Error", http.StatusInternalServerError)
}


func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "signup.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

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
	http.Redirect(w, r, "/", http.StatusOK)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "login.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	var dbusername string
	var dbpassword string

	err = db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&dbusername, &dbpassword)
	if err != nil {
		w.Write([]byte("User Does Not Exist"))
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(password))
	if err != nil {
		w.Write([]byte("Credentials Did not Match!"))
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Session-handling
	session,err:=sessionStore.Get(r,username)
	if err!=nil{
		handleSessionError(w,err)
		return
	}
	session.Values["username"]=username
	if err:=session.Save(r,w);err!=nil{
		handleSessionError(w,err)
		return
	}
	log.WithField("username",username).Info("Completed Login & Session is saved")

	//redirect  to the next page.
	http.Redirect(w, r, "/register", 301)
}


func Logout(w http.ResponseWriter, r *http.Request){

	session,err:=sessionStore.Get(r,SessionName)
	if err!=nil{
		handleSessionError(w,err)
		return
	}
	session.Values["username"]=""
	if err:=session.Save(r,w);err!=nil{
		handleSessionError(w,err)
		return
	}
	log.Info(" Successfully LOG OUT")
	http.Redirect(w,r,"/",301)
}


func Homepage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "register.html")
	}
	http.Redirect(w, r, "/stored", 301)
}

func TermsandConditions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "terms.html")
	case "POST":
		http.Redirect(w, r, "/", 301)
	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}
}
func Storedb(w http.ResponseWriter, r *http.Request) {
	var VoterID, LastName, FirstName, Age, Sex, State, City string
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "db.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
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
	if err != nil {
		panic(err.Error())
	}
	http.Redirect(w, r, "/vote", 301)
}

func Vote(w http.ResponseWriter, r *http.Request) {
	var Party_Name string
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "vote.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		Party_Name = r.FormValue("Party")
	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}
	_, err := db.Exec("INSERT INTO parties(Party_Name)VALUES(?)", Party_Name)
	if err != nil {
		panic(err.Error())
	}
	http.Redirect(w, r, "/result", 301)
}

func outputHTML(w http.ResponseWriter, filename string, data interface{}) {
	t, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func Result(w http.ResponseWriter, r *http.Request) {

	var count1, count2, count3 int
	type sample struct {
		Myvar string
		Votes int
	}
	switch r.Method {
	case "GET":
		fmt.Println("RESULT GET METHOD")
		http.ServeFile(w, r, "result.html")

	case "POST":
		username := r.FormValue("username")
		password := r.FormValue("password")
		var dbusername string
		var dbpassword string

		err = db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&dbusername, &dbpassword)
		if err != nil {
			w.Write([]byte("User Does Not Exist"))
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(password))
		if err != nil {

			w.Write([]byte("Credentials Did not Match!"))
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		rows, err := db.Query("SELECT Party_Name FROM parties WHERE Party_Name=?", "BJP")
		if err != nil {
			panic(err.Error())
		}
		for rows.Next() {
			count1++
		}
		rows_MNS, err1 := db.Query("SELECT Party_Name FROM parties WHERE Party_Name=?", "MNS")
		if err1 != nil {
			panic(err.Error())
		}
		count2 = 0
		for rows_MNS.Next() {
			count2++
		}
		rows_INC, err2 := db.Query("SELECT Party_Name FROM parties WHERE Party_Name=?", "INC")
		if err2 != nil {
			panic(err.Error())
		}
		count3 = 0
		for rows_INC.Next() {
			count3++
		}
		s := sample{Myvar: "BJP received :-", Votes: count1}
		outputHTML(w, "final.html", s)

	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}
	http.Redirect(w, r, "/final", 301)
}

func Final(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "final.html")
	}
}



//links to Follow
//https://devcenter.heroku.com/articles/go-sessions
//https://github.com/heroku-examples/go-sessions-demo/blob/master/main.go
//https://curtisvermeeren.github.io/2018/05/13/Golang-Gorilla-Sessions
//https://github.com/CurtisVermeeren/Gorilla-Sessions-Tutorial/blob/master/CookiestoreSession/main.go
//https://github.com/rivo/sessions
//https://gowebexamples.com/sessions/

