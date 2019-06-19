package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username      interface{}
	Authenticated bool
}

var err error

//Session name to store session under
const SessionName = "Voting-Session"

///var sessionStore *sessions.CookieStore
var encryptionKey = securecookie.GenerateRandomKey(32)
var loggeduserSession = sessions.NewCookieStore([]byte(encryptionKey))

//var userMap = make(map[string]interface{})

func Init() {
	fmt.Println("Init function called")
	loggeduserSession.Options = &sessions.Options{Path: "/", MaxAge: 86400 * 1, HttpOnly: true}
	///	gob.Register(User{})
	fmt.Println("Init function has been ended")
}

// to handle the error if anything goes wrong
func handleSessionError(w http.ResponseWriter, err error) {
	fmt.Println("session handler has invoked the error")
	http.Error(w, "Unable to retrieve the session data", http.StatusInternalServerError)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SignUP Called")
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
		//	fmt.Println("User Account Has been created")
		w.Write([]byte("User Accout Created"))

	case err != nil:
		http.Error(w, "Unable to create your account (THREE)", http.StatusInternalServerError)
		return
	default:
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	http.Redirect(w, r, "/", 301)
	fmt.Println("SignUP has been ended")
}

func Login(w http.ResponseWriter, r *http.Request) {

	//u := User{}
	fmt.Println("Login Handler called")

	session, err := loggeduserSession.Get(r, SessionName)

	//fmt.Println("value of Sesion is ", session)
	//fmt.Println("value of error is ", err)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	fmt.Println("Login Method ", r.Method)
	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	var dbusername string
	var dbpassword string
	if r.Method == "POST" {
		err = db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&dbusername, &dbpassword)
		if err != nil {
			w.Write([]byte("User Does Not Exist"))
			session.AddFlash("USER DOES NOT EXIST")
			err = session.Save(r, w)
			if err != nil {
				handleSessionError(w, err)
			}
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(password))
		if err != nil {
			w.Write([]byte("Credentials Did not Match!"))
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// a new session will be created when user sucessfully logged IN
		session, err = loggeduserSession.New(r, SessionName)
		//fmt.Println("session values", session)
		//fmt.Println("error value is ", err)
		if err != nil {
			handleSessionError(w, err)
		}
		session.Values["username"] = username //from r.Formvalue

		if err := session.Save(r, w); err != nil {
			handleSessionError(w, err)
			return
		}
		fmt.Println("New session has been created and username has been added")

		//	u = User{Username: session.Values["username"], Authenticated: true}
		//fmt.Println("LOGIN OutputHTML called", u)
		//
		fmt.Println("redirect  to the next page.")
		http.Redirect(w, r, "/stored", 301)
		//	outputHTML(w, "db.html", u)
	}

}
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// there is no session as user is not logged in
	fmt.Println("forgot password has been invoked")
	if r.Method != "POST" {
		http.ServeFile(w, r, "forgotpass.html")
		return
	}
	username := r.FormValue("username")

	//	var dbusername string
	//check whether this user exist already or not!

	rows, err := db.Query("SELECT username FROM users WHERE username=?", username)
	if err != nil {
		panic(err.Error())
	}
	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		w.Write([]byte("USER Does exist"))
		http.Redirect(w, r, "/forbidden", 403)
		return
	}
	http.Redirect(w, r, "/resetpass", 301)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("password reset handler has been called")
	if r.Method != "POST" {
		http.ServeFile(w, r, "resetpass.html")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("pwd1")
	fmt.Println("username", username)
	fmt.Println("new password", password)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Unable to crate accout right now (ONE)", http.StatusInternalServerError)
		return
	}
	_, err = db.Exec("UPDATE users SET password=? WHERE username =?", hashedPassword, username)
	//	_, err = db.Exec("UPDATE users SET password=? WHERE username =?", , hashedPassword)
	if err != nil {
		http.Error(w, "Unable to create your account (TWO)", http.StatusInternalServerError)
		return
	}
	//	fmt.Println("User Account Has been created")
	w.Write([]byte("Password has been successfully Updated"))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "logout.html")
	fmt.Println("LOGOUT HANDLER IS CALLED")
	session, err := loggeduserSession.Get(r, SessionName)
	fmt.Println("sesion value in LOGOUT HANDLER", session)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	session.Values["username"] = ""
	//loggeduserSession.MaxAge(-1)
	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	//to check whether any session is ther or not after logout
	session, err = loggeduserSession.Get(r, SessionName)
	if err != nil {
		fmt.Println("error in logout function", err)
	}
	fmt.Println("LOGOUT HANDLER ENDED")
	http.Redirect(w, r, "/logout", 301)
	return
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func Register(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Register called")
	//u := User{}
	session, err := loggeduserSession.Get(r, SessionName)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	if session.Values["username"] == "" {
		//NO USER found
		http.Redirect(w, r, "/forbidden", 301)
		return
	}
	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	u := User{Username: session.Values["username"], Authenticated: true}
	//outputHTML(w, "register.html", u)
	if r.Method != "POST" {
		//http.ServeFile(w, r, "register.html")
		outputHTML(w, "register.html", u)
	}

	http.Redirect(w, r, "/vote", 301)
}

func TermsandConditions(w http.ResponseWriter, r *http.Request) {
	//u := User{}
	session, err := loggeduserSession.Get(r, SessionName)
	if err != nil {
		handleSessionError(w, err)
		return
	}

	if session.Values["username"] == "" {
		//NO USER found
		http.Redirect(w, r, "/forbidden", 301)
		return
	}

	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	u := User{Username: session.Values["username"], Authenticated: true}
	//	outputHTML(w, "terms.html", u)

	switch r.Method {
	case "GET":
		//http.ServeFile(w, r, "terms.html")
		outputHTML(w, "terms.html", u)
	case "POST":
		http.Redirect(w, r, "/", 301)
	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}
}
func Storedb(w http.ResponseWriter, r *http.Request) {

	fmt.Println("DB  handler has been called")
	//	u := User{}
	session, err := loggeduserSession.Get(r, SessionName)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	if session.Values["username"] == "" {
		//NO USER found
		http.Redirect(w, r, "/forbidden", 301)
		return
	}

	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	u := User{Username: session.Values["username"], Authenticated: true}
	//	outputHTML(w, "db.html", u)

	var VoterID, LastName, FirstName, Age, Sex, State, City string
	switch r.Method {
	case "GET":
		//http.ServeFile(w, r, "db.html")
		outputHTML(w, "db.html", u)
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
	//	u = User{Username: session.Values["username"], Authenticated: true}
	//	outputHTML(w, "register.html", u)
	//	fmt.Println("randome", u)
	http.Redirect(w, r, "/register", 301)
}

func Vote(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Vote  has been called")
	//	u := User{}
	session, err := loggeduserSession.Get(r, SessionName)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	fmt.Println("Before POST VOTE check username is present or not")
	fmt.Println("Value of Signed-IN user", session.Values["username"])
	if session.Values["username"] == "" {
		//NO USER
		fmt.Println("Redirecting to forbidden page as no user found")
		http.Redirect(w, r, "/forbidden", 301)
		return
	}

	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	u := User{Username: session.Values["username"], Authenticated: true}
	//outputHTML(w, "vote.html", u)

	var Party_Name string
	switch r.Method {
	case "GET":
		//http.ServeFile(w, r, "vote.html")
		outputHTML(w, "vote.html", u)
	case "POST":
		/*
			if session.Values["username"] == "" {
				fmt.Println("Value of Logged in User is", session.Values["username"])
				http.Redirect(w, r, "/forbidden", 301)
			}*/
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		Party_Name = r.FormValue("Party")
	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}
	/*
		fmt.Println("before taking the vote check the user session again")
		fmt.Println("Value of Logged in User is", session.Values["username"])
		if session.Values["username"] == "" {
			http.Redirect(w, r, "/forbidden", 301)
		} */
	_, err = db.Exec("INSERT INTO parties(Party_Name)VALUES(?)", Party_Name)
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

var count1, count2, count3 int

type sample struct {
	Username interface{}
	A        int
	B        int
	C        int
}

func Result(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Result has been called")
	//	u := User{}
	session, err := loggeduserSession.Get(r, SessionName)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	if session.Values["username"] == "" {
		//NO USER found
		http.Redirect(w, r, "/forbidden", 301)
	}
	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	u := User{Username: session.Values["username"], Authenticated: true}
	//outputHTML(w, "result.html", u)

	switch r.Method {
	case "GET":
		fmt.Println("RESULT GET METHOD")
		//http.ServeFile(w, r, "result.html")
		outputHTML(w, "result.html", u)

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

	default:
		fmt.Fprintf(w, "Service only Supports GET and POST")
	}

	http.Redirect(w, r, "/final", 301)
	//s := sample{A: count1, B: count2, C: count3}
	//outputHTML(w, "final.html", s)
}

func Final(w http.ResponseWriter, r *http.Request) {
	//u := User{}
	fmt.Println("Final has been called")
	fmt.Println("count values are", count1, count2, count3)
	session, err := loggeduserSession.Get(r, SessionName)
	if err != nil {
		handleSessionError(w, err)
		return
	}
	if session.Values["username"] == "" {
		//NO USER found
		http.Redirect(w, r, "/forbidden", 301)
	}
	if err := session.Save(r, w); err != nil {
		handleSessionError(w, err)
		return
	}
	s := sample{Username: session.Values["username"], A: count1, B: count2, C: count3}
	outputHTML(w, "final.html", s)

	//u = User{Username: session.Values["username"], Authenticated: true}
	//	outputHTML(w, "final.html", u)
	/*
		if r.Method == "GET" {
			http.ServeFile(w, r, "final.html")
		}*/
}
func forbidden(w http.ResponseWriter, r *http.Request) {
	fmt.Println("forbidden handler is called")
	fmt.Println("r.method is", r.Method)
	http.ServeFile(w, r, "forbidden.html")
	return

}
