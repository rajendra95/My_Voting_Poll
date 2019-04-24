package main

import (
	"crypto/bcrypt"
	"database/sql"
	"encoding/json"
	"net/http"
)

// struct for getting the user credentials - username password

type Credentials struct {
	Password string `json:"password",db:"password"`
	Username string `json:"username",db:"username"`
}

// we will write handlers :

func Signup(w http.ResponseWriter, r *http.Request) {
	// parse and decode the request body into the new credentials instance
	creds := Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If something went wrong write 400 bad request in the response.
		w.WriteHeader(http.StatusBadRequest)
	}
	return
	// if OK -> encrypt the password using bcrypt algorithm
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	// now we have to insert this encrypted value into a database
	if _, err = db.Query("insert the values ($1,$2)", creds.Username, string(hashedPassword)); err != nil {
		// if we faced an issue in inserting the data into the database then we need to send 500 error
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
	// if everything went alright then we data will get stored inside the database and default status code 200 will be return.
}

func Signin(w http.ResponseWriter, r *http.Request) {
	// parse and decode the request body into creds instance
	creds := Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If something went wrong write 400 bad request in the response.
		w.WriteHeader(http.StatusBadRequest)
	}
	// Get the existing entry present in the database for the given username
	result := db.QueryRow("select password from users based on username ", creds.Username)

	// We create another instance of `Credentials` to store the credentials we get from the database
	storedCreds := Credentials{}
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		// There are two possibilites now - (Either The username will not be there in the database OR Internal server Error)

		// If an entry with the username does not exist, send an "Unauthorized"(401) status
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the error is of any other type, send a 500 status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// compare the stored and received credentials
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		// If two passwords does not match then return 401 unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
	}
	// If we reach this point, that means the users password was correct, and that they are authorized
	// The default 200 status is sent

}
