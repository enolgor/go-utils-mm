package lambda

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	aes "github.com/enolgor/go-utils/crypto"
	"github.com/enolgor/go-utils/server"
)

type Auth struct {
	secret   string
	formPath string
	expiry   time.Duration
	checker  func(user, pass string) bool
}

func NewAuth(secret, formPath string, expiry time.Duration, checker func(user, pass string) bool) *Auth {
	return &Auth{secret, formPath, expiry, checker}
}

func (a *Auth) AuthHandler() server.ChainHandler {
	return func(w http.ResponseWriter, req *http.Request) bool {
		if authstr := a.getAuthString(req); authstr != "" {
			if a.checkAuthentication(authstr) {
				return true
			}
		}
		a.authenticateResponse(w, req.URL.String())
		return false
	}
}

func (a *Auth) getAuthString(req *http.Request) string {
	if cook, err := req.Cookie("auth"); err != nil {
		return req.Header.Get("x-auth")
	} else {
		return cook.Value
	}
}

func (a *Auth) FormHandler() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var redirect, authstr, user, pass string
		var err error
		if redirect = req.URL.Query().Get("redirect"); redirect == "" {
			redirect = "/"
		}
		if err = req.ParseForm(); err != nil {
			a.authenticateResponse(w, redirect)
			return
		}
		if user = req.FormValue("user"); user == "" {
			a.authenticateResponse(w, redirect)
			return
		}
		if pass = req.FormValue("pass"); pass == "" {
			a.authenticateResponse(w, redirect)
			return
		}
		if authstr, err = a.generateAuthentication(); err != nil {
			a.authenticateResponse(w, redirect)
			return
		}

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(200)
	}
}

func (a *Auth) checkAuthentication(authstr string) bool {
	var decrypted string
	var err error
	if decrypted, err = aes.Decrypt(a.secret, authstr); err != nil {
		return false
	}
	idx := strings.Index(decrypted, ":")
	if idx == -1 {
		return false
	}
	return a.checker(decrypted[:idx], decrypted[idx+1:])
}

func (a *Auth) generateAuthentication(user, pass string) (string, error) {
	decrypted := fmt.Sprintf("%s:%s", user, pass)
	return aes.Encrypt(a.secret, decrypted)
}

func (a *Auth) authenticateResponse(w http.ResponseWriter, redirect string) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="UTF-8">
		</head>
		<body>
			<form action="%s?redirect=%s" method="post">
				<label for="user">User:</label>
				<input type="text" id="user" name="user"><br><br>
				<label for="pass">Password:</label>
				<input type="password" id="pass" name="pass"><br><br>
				<input type="submit" value="Authenticate">
			</form>
		</body>
	</html>
	`, a.formPath, url.QueryEscape(redirect))
}
