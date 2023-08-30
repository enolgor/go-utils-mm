package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtAuth struct {
	parser     *jwt.Parser
	key        []byte
	expiration time.Duration
	verifyUser func(user, pass string) (bool, error)
}

const jwtCookieName string = "_token"

type jwtContextKey int

const contextJwtClaims jwtContextKey = iota

func NewJwtAuth(key []byte, expiration time.Duration, verifyUser func(user, pass string) (bool, error)) *JwtAuth {
	return &JwtAuth{
		parser:     jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})),
		key:        key,
		expiration: expiration,
		verifyUser: verifyUser,
	}
}

func (ja *JwtAuth) SoftAuthHandler() ChainHandler {
	return func(w http.ResponseWriter, req *http.Request) bool {
		var jwtString string
		if jwtString = getJwtString(req); jwtString == "" {
			return true
		}
		tkn, err := ja.parser.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
			return ja.key, nil
		})
		if err != nil {
			return true
		}
		if !tkn.Valid {
			return true
		}
		AddContextValue(req, contextJwtClaims, tkn.Claims)
		return true
	}
}

func (ja *JwtAuth) StrictAuthHandler(redirect string) ChainHandler {
	return func(w http.ResponseWriter, req *http.Request) bool {
		if strings.Contains(redirect, "?") {
			redirect = fmt.Sprintf("%s&redirect=%s", redirect, url.QueryEscape(req.URL.Path))
		} else {
			redirect = fmt.Sprintf("%s?redirect=%s", redirect, url.QueryEscape(req.URL.Path))
		}
		var jwtString string
		if jwtString = getJwtString(req); jwtString == "" {
			Response(w).Status(http.StatusUnauthorized).Redirect(redirect)
			return false
		}
		tkn, err := ja.parser.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
			return ja.key, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				Response(w).Status(http.StatusUnauthorized).Redirect(redirect)
				return false
			}
			Response(w).Status(http.StatusBadRequest).Redirect(redirect)
			return false
		}
		if !tkn.Valid {
			Response(w).Status(http.StatusUnauthorized).Redirect(redirect)
			return false
		}
		AddContextValue(req, contextJwtClaims, tkn.Claims)
		return true
	}
}

func (ja *JwtAuth) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var redirect, user, pass string
		var err error
		if redirect = req.URL.Query().Get("redirect"); redirect == "" {
			redirect = "/"
		}
		if err = req.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if user = req.FormValue("user"); user == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if pass = req.FormValue("pass"); pass == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if ok, err := ja.verifyUser(user, pass); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expiration := time.Now().Add(ja.expiration)
		claims := jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			Subject:   user,
		}

		tkn, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(ja.key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		Response(w).
			WithCookie(&http.Cookie{
				Name:     jwtCookieName,
				Value:    tkn,
				HttpOnly: true,
				Expires:  expiration,
				SameSite: http.SameSiteStrictMode,
			}).
			Redirect(redirect)
	}
}

func (ja *JwtAuth) SampleAuthForm(target, defaultRedirect string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Has("redirect") {
			defaultRedirect = req.URL.Query().Get("redirect")
		}
		if HasContextValue[jwt.Claims](req, contextJwtClaims) {
			Response(w).Redirect(defaultRedirect)
			return
		}

		Response(w).WithBody(fmt.Sprintf(`
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
		`, target, defaultRedirect)).AsHtml()
	}
}

func JwtClaims(req *http.Request) jwt.Claims {
	var claims jwt.Claims
	if ok := GetContextValue(req, contextJwtClaims, &claims); ok {
		return claims
	}
	return nil
}

func getJwtString(req *http.Request) string {
	if cook, err := req.Cookie(jwtCookieName); err == nil {
		return cook.Value
	}
	header := req.Header.Get("Authentication")
	if header == "" {
		return ""
	}
	idx := strings.Index(header, "Bearer ")
	if idx != 0 {
		return ""
	}
	return header[7:]
}
