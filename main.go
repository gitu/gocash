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

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var app = negroni.New()

func getDb() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Printf("error [%v] while connecting with url [%v]", err, os.Getenv("DATABASE_URL"))
	}
	return db, err
}

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
	mainRouter.HandleFunc("/setup", SetupDatabase)

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

	db, _ := getDb()

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
		db, _ := getDb()
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

func SetupDatabase(w http.ResponseWriter, r *http.Request) {

	db, _ := getDb()

	conf, err := goose.NewDBConf(os.Getenv("DB_CONF"), os.Getenv("DB_CONF_ENV"), os.Getenv("DB_CONF_SCHEMA"))
	if err != nil {
		log.Printf("error while getting config: %v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbVersionPre, err := goose.GetDBVersion(conf)
	if err != nil {
		log.Printf("error while getting last version from db: %v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	target, err := goose.GetMostRecentDBVersion(conf.MigrationsDir)
	if err != nil {
		log.Printf("error while getting last version: %v", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		if err := goose.RunMigrations(conf, conf.MigrationsDir, target); err != nil {
			log.Printf("error while running migrations: %v", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, userErr := db.Query("select id from users")
		if userErr != nil {
			encryptedPassword, encErr := bcrypt.GenerateFromPassword([]byte(r.PostFormValue("password")), 15)
			if encErr != nil {
				log.Printf("Error generating password")
			}
			t := db.MustBegin()
			db.MustExec("INSERT INTO users (user_name, full_name, password_hash, is_enabled) select $1, $2, $3, $4 where NOT EXISTS (select id from users where user_name=$1)", r.PostFormValue("user_name"), r.PostFormValue("full_name"), encryptedPassword, true)
			t.Commit()
		}

		dbVersionPost, err := goose.GetDBVersion(conf)
		if err != nil {
			log.Printf("error while getting last version from db: %v", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Successfully updated from %v to %v", dbVersionPre, dbVersionPost)
	} else {

		addVal := ""
		_, userErr := db.Query("select id from users")
		if userErr != nil {
			addVal = "User name: <input type='text' name='user_name'><br>" +
				"Password: <input type='text' name='password'><br>" +
				"Full Name: <input type='text' name='full_name'><br>"
		}

		if target == dbVersionPre && userErr == nil {
			fmt.Fprintf(w, "All Up to Date!")
		} else {

			fmt.Fprintf(w, "<html><body>Upgrade from %v to %v <br/><form action='/setup' method='post'>"+
				"%v"+
				"<input type='submit' value='upgrade'></form></body></html>", dbVersionPre, target, addVal)
		}
	}
}
