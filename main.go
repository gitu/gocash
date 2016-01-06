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
)

var app = negroni.New()

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

	apiRoutes := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	apiNegroni := negroni.New(NewAuth(), negroni.Wrap(apiRoutes))

	mainRouter.PathPrefix("/api").Handler(apiNegroni)
	apiRoutes.HandleFunc("/user", handlers.UserHandler)

	return mainRouter
}

// User model
type User struct {
	UserId   string `form:"userid" json:"userid" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.UserId == "asdf" && user.Password == "asdf" {

		// Create JWT token
		token := jwt.New(jwt.SigningMethodHS512)
		token.Claims["userid"] = user.UserId
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
		log.Printf("Unsuccessful Login for user %v", user.UserId)
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
		context.Set(r, "user", token.Claims["userid"])
		next(rw, r)
	} else {
		log.Printf("Error while parsing token: %v", err.Error())
		http.Error(rw, err.Error(), http.StatusUnauthorized)
	}
}
