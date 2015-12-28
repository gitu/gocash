package main

import (
	"log"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"

	"errors"
	"github.com/gitu/gocash/backend/handlers"
	"github.com/gorilla/context"
	"net/http"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

var app = negroni.New()

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if os.Getenv("SESSION_SECRET") == "" {
		log.Fatal("No SESSION_SECRET set!")
	}
	gothic.Store = store

	goth.UseProviders(
		gplus.New(os.Getenv("GPLUS_KEY"), os.Getenv("GPLUS_SECRET"), os.Getenv("SERVER_URL")+"/auth/gplus/callback"),
	)
	gothic.GetProviderName = getProviderName

	//These middleware is common to all routes
	app.Use(negroni.NewRecovery())
	app.Use(negroni.NewLogger())
	app.Use(negroni.NewStatic(http.Dir(os.Getenv("CLIENT_DIR"))))
	app.Use(NewAuth())
	app.UseHandler(NewRoute())

}

func main() {
	app.Run(":" + os.Getenv("PORT"))
}

func NewRoute() *mux.Router {

	mainRouter := mux.NewRouter().StrictSlash(true)

	apiRoutes := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)
	apiNegroni := negroni.New(NewUserRequired(), negroni.Wrap(apiRoutes))

	mainRouter.PathPrefix("/api").Handler(apiNegroni)
	mainRouter.HandleFunc("/auth/{provider}/callback", CallbackHandler)
	mainRouter.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler)
	mainRouter.HandleFunc("/logout", LogoutHandler)

	apiRoutes.HandleFunc("/user", handlers.UserHandler)

	return mainRouter
}

func getProviderName(req *http.Request) (string, error) {
	provider := mux.Vars(req)["provider"]
	if provider == "" {
		return provider, errors.New("you must select a provider")
	}
	return provider, nil
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		log.Fatalln(w, err)
		return
	}

	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	session.Values["user"] = user

	session.Save(r, w)

	http.Redirect(w, r, "/", 302)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	session.Options.MaxAge = -1

	session.Save(r, w)

	http.Redirect(w, r, "/", 302)
}

type Auth struct {
}

func NewAuth() *Auth {
	return &Auth{}
}

func (l *Auth) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}

	user := session.Values["user"]
	if user != nil {
		context.Set(r, "user", user)
	}
	next(rw, r)
}

type UserRequired struct {
}

func NewUserRequired() *UserRequired {
	return &UserRequired{}
}

func (l *UserRequired) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	user := context.Get(r, "user")
	if user == nil {
		http.Error(rw, "Unauthorized", 401)
		return
	}
	next(rw, r)
}
