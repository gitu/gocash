package main

import (
	"log"
	"os"

	"github.com/codegangsta/negroni"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"encoding/json"
	"fmt"
	"github.com/gitu/gocash/handlers"
	"github.com/gorilla/context"
	"net/http"
	"time"

	"github.com/gitu/gocash/Godeps/_workspace/src/golang.org/x/crypto/bcrypt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var app = negroni.New()

var db *sqlx.DB

type User struct {
	Id           int64  `json:"id" db:"id"`
	UserName     string `json:"userName" db:"user_name"`
	Password     string `json:"password,omitempty" `
	PasswordHash []byte `json:"-" db:"password_hash"`
	FullName     string `json:"fullName" db:"full_name"`
	IsEnabled    bool   `json:"isEnabled" db:"is_enabled"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("error [%v] while connecting with url [%v]", err, os.Getenv("DATABASE_URL"))
		log.Fatalln(err)
	}

	genpw, err := bcrypt.GenerateFromPassword([]byte("asdf"), 15)
	if err != nil {
		log.Printf("Error generating password")
	}

	db.MustExec("INSERT INTO users (user_name, full_name, password_hash, is_enabled) select $1, $2, $3, $4 where NOT EXISTS (select id from users where user_name=$1)", "asdf", "asdf", genpw, true)

	//These middleware is common to all routes
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(negroni.NewStatic(http.Dir(os.Getenv("CLIENT_DIR"))))
	app.UseHandler(NewRoute())

}

func main() {
	app.Run(":" + os.Getenv("PORT"))
}

func NewRoute() *mux.Router {

	mainRouter := mux.NewRouter().StrictSlash(true)
	mainRouter.HandleFunc("/auth", Authenticate)

	apiRoutes := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	apiNegroni := negroni.New(NewAuth(), negroni.Wrap(apiRoutes))

	mainRouter.PathPrefix("/api").Handler(apiNegroni)
	apiRoutes.HandleFunc("/user", handlers.UserHandler)

	return mainRouter
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dbUser := User{}
	err = db.Get(&dbUser, "SELECT * FROM users WHERE user_name=$1", user.UserName)
	if err != nil {
		log.Printf("Unsuccessful Login for user %v does not exists: %v", user.UserName, err)
		http.Error(w, "MEEP MEEP", http.StatusUnauthorized)
		return
	}

	if bcrypt.CompareHashAndPassword(dbUser.PasswordHash, []byte(user.Password)) == nil {
		// Create JWT token
		token := jwt.New(jwt.SigningMethodHS512)
		token.Claims["user_id"] = dbUser.Id
		token.Claims["user_name"] = dbUser.UserName
		token.Claims["iat"] = time.Now().Unix()
		token.Claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
		token.Claims["nbf"] = time.Now().Unix()
		tokenString, err := token.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
		if err != nil {
			log.Printf("Error while generating token: %v", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		data := map[string]string{
			"token": tokenString,
		}

		js, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error while unmarshalling: %v", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		log.Printf("Unsuccessful Login for user %v", user.UserName)
		http.Error(w, "MEEP MEEP", http.StatusUnauthorized)
	}
}

type Auth struct {
}

func NewAuth() *Auth {
	return &Auth{}
}

func (l *Auth) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token, err := jwt.ParseFromRequest(r, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("TOKEN_SECRET")), nil
	})
	if err == nil && token.Valid {
		dbUser := User{}
		err = db.Get(&dbUser, "SELECT * FROM users WHERE id=$1", token.Claims["user_id"])
		if err == nil && dbUser.IsEnabled {
			context.Set(r, "user", dbUser)
			next(rw, r)
		} else {
			log.Printf("Claims: %v", token.Claims)
			http.Error(rw, err.Error(), http.StatusUnauthorized)
		}
	} else {
		log.Printf("Error while parsing token: %v", err.Error())
		http.Error(rw, err.Error(), http.StatusUnauthorized)
	}
}
