package lambda

const authenticateResponseBody string = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
  </head>
  <body>
    <form action="/authenticate" method="post">
       <label for="pass">Pass:</label>
      <input type="password" id="pass" name="pass"><br><br>
      <input type="submit" value="Authenticate">
    </form>
  </body>
</html>
`

func authenticateResponse() Response {
	return Response{
		Body:       []byte(authenticateResponseBody),
		StatusCode: 401,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
	}
}
