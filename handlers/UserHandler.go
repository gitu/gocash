package handlers

import (
	"net/http"

	"github.com/gorilla/context"
	"gopkg.in/unrolled/render.v1"
	"log"
)

var view = render.New(render.Options{
	Directory:     "templates",
	Extensions:    []string{".html"},
	IsDevelopment: true,
	IndentJSON:    true,
})

func UserHandler(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user")
	log.Println(user)
	view.JSON(w, 200, user)
}
