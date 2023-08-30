package examples

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/enolgor/go-utils/crypto"
	"github.com/enolgor/go-utils/parse"
	"github.com/enolgor/go-utils/server"
	"github.com/golang-jwt/jwt/v5"
)

var hashedPasswords map[string]string = map[string]string{}

func init() {
	cost := crypto.OptimalCost(250 * time.Millisecond)
	hashedPasswords["admin"], _ = crypto.HashPassword("test", cost)
}

func Server(port int) {
	key := parse.Must(parse.HexBytes)("bc27bec0c4291b4e43a2ec657d8afc9b668e158c6acd4004ffb1faa16c5b88bf")
	jwt := server.NewJwtAuth(key, 5*time.Minute, func(user, pass string) (bool, error) {
		hpass, ok := hashedPasswords[user]
		return ok && crypto.ComparePassword(hpass, pass) == nil, nil
	})
	strictAuth := jwt.StrictAuthHandler("/login")
	softAuth := jwt.SoftAuthHandler()
	login := jwt.LoginHandler()
	form := jwt.SampleAuthForm("/login", "/")

	var router *server.Router
	var err error
	if router, err = server.NewRouterBuilder().
		Get("/login", server.Handle(server.Get, softAuth, form)).
		Post("/login", server.Handle(server.Post, login)).
		Get("/(.*)", server.Handle(server.Get, strictAuth, hello)).
		Build(); err != nil {
		log.Fatal(err)
	}
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		panic(err)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	var claims jwt.Claims
	if claims = server.JwtClaims(req); claims == nil {
		server.Response(w).Status(http.StatusInternalServerError).WithBody("claims not found").AsTextPlain()
		return
	}
	sub, err := claims.GetSubject()
	if err != nil {
		server.Response(w).Status(http.StatusInternalServerError).WithBody(err).AsTextPlain()
		return
	}
	pathParams := server.PathParams(req)
	if pathParams[0] == "panic" {
		panic("example panic!")
	}
	data := struct {
		User string `json:"user"`
	}{User: sub}
	server.Response(w).WithBody(data).AsJson()
}
