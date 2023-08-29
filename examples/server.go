package examples

import (
	"fmt"
	"net/http"
	"time"

	"github.com/enolgor/go-utils/crypto"
	"github.com/enolgor/go-utils/parse"
	"github.com/enolgor/go-utils/server"
	"github.com/golang-jwt/jwt/v5"
)

var hashedPasswords map[string]string = map[string]string{
	"eneko": "$2a$10$nIhmpNzH1uDfXJ7i.NSByekfJ3KbKOO3W1Kf9qfeZ3MYg.YvMUV9i",
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
	form := jwt.SampleAuthForm("/doLogin", "/")

	router := server.NewRouter()
	router.Register("GET", "/", server.Handle(server.Get, strictAuth, hello))
	router.Register("GET", "/login", server.Handle(server.Get, softAuth, form))
	router.Register("POST", "/doLogin", server.Handle(server.Post, login))
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
	fmt.Println(pathParams.Get("idiot"))
	fmt.Println(pathParams["idiot"])
	fmt.Println(pathParams.Get("idiota"))
	server.Response(w).WithBody(fmt.Sprintf("hello %s", sub)).AsTextPlain()
}
