package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const domain = "http://localhost:9595/"

const indexPage = `
<!DOCTYPE html>
<html>
<head>
	<title>URL Shortener</title>
  <meta charset="UTF-8">
  <style type="text/css">
  	body {
  		color: #555;
  	}
  	.container {
  		width: 960px;
  		margin: 20px auto;
  		text-align: center;
  	}
    .sub-text {
      font-size: 0.85em;
      margin-top: -15px;
    }
  	.box {
  		width: 500px;
  		background-color: #f9f3ec;
  		padding: 20px;
  		border-radius: 2px;
  		margin: 15px auto;
      position: relative;
      z-index: 100;
  	}
  	.box input[type="text"] {
  		width: 400px;
  		height: 35px;
  		margin: 0 auto;
  		border: 1px solid #bbb;
  		font-size: 1.1em;
  		padding: 0 10px 0 10px;
  		border-radius: 2px;
      -webkit-appearance: none;
      -moz-appearance:    none;
      appearance:         none;
  	}
  	.box input[type="submit"] {
  		width: 300px;
  		height: 35px;
  		border: none;
  		background-color: #2da7ee;
  		color: #fff;
  		font-size: 1em;
  		margin: 10px auto;
  		border-radius: 2px;
  		cursor: pointer;
      -webkit-appearance: none;
      -moz-appearance:    none;
      appearance:         none;
  	}
  	.box input[type="submit"]:hover {
  		background-color: #1398e6;
  	}
  	.footer {
  		font-size: 0.85em;
  	}
  	.footer a {
			color: #555;
			text-decoration: none;
			font-weight: bold;
  	}
  	.footer a:hover {
  		color: #5cbaf2;
  	}
    .extension {
      width: 500px;
      margin: 0 auto;
      font-size: 0.75em;
      padding: 25px 20px 20px 20px;
      background-color: #f3ebdf;
      border-radius: 2px;
      position: relative;
      top: -20px;
    }
  </style>
</head>
<body>
	<div class="container">
  	<h1>URL Shortener</h1>
    <div class="sub-text">A simple URL shortener written in Go</div>
  	<div class="box">
  	<p>Enter or paste your long URL in the box below.</p>
    <form action="/" method="post">
    <input type="text" name="url" placeholder="http://kiwi.io">
    <input type="submit" value="Shorten URL">
    </form>
    </div>
    <div class="footer">
      Created by Earl Balai<br />
      Powered by the <a href="http://golang.org">Go</a> language
    </div>
  </div>
</body>
</html>
`

func RegisterRoutes() {
	router := mux.NewRouter()

	router.HandleFunc("/", mainView)
	router.HandleFunc("/ext", extView).Methods("POST")
	router.HandleFunc("/l/{code}", navigate)

	http.Handle("/", router)
}

func WriteContent(w http.ResponseWriter, s string) {
	fmt.Fprintf(w, s)
}

func mainView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	switch r.Method {
	case "GET":
		WriteContent(w, indexPage)

	case "POST":
		createShortlink(w, r.FormValue("url"), false)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func createShortlink(w http.ResponseWriter, url string, extension bool) {
	w.Header().Set("Content-Type", "text/html")

	if !strings.Contains(url, ".") {
		http.Error(w, "Invalid url...", http.StatusForbidden)
		return
	}

	db := OpenDB()

	var count int64
	row := db.QueryRow("SELECT COUNT(*) FROM storage")

	err := row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	count++

	short := strconv.FormatInt(count, 36)

	var db_code string

	q := db.QueryRow("SELECT code FROM storage WHERE url = $1", url)

	e := q.Scan(&db_code)
	if e != nil {
		db.Exec("INSERT INTO storage(id, url, code) VALUES($1, $2, $3)", count, url, short)
		fmt.Printf("New URL stored to DB. URL: %s with CODE: %s\n", url, short)
	} else {
		short = db_code
		fmt.Printf("URL already exists. Code: %s\n", db_code)
	}

	defer db.Close()

	short_url := domain + short

	result := `
<!DOCTYPE html>
<html>
<head>
	<title>URL Shortener</title>
  <meta charset="UTF-8">
  <style type="text/css">
  	body {
  		color: #555;
  	}
  	.container {
  		width: 960px;
  		margin: 20px auto;
  		text-align: center;
  	}
  	.box {
  		width: 500px;
  		background-color: #f9f3ec;
  		padding: 20px;
  		border-radius: 2px;
  		margin: 0 auto;
  	}
  	.box a {
  		color: #5cbaf2;
  		font-size: 1.3em;
  		font-weight: bold;
  		text-decoration: none;
  	}
  	.footer {
  		margin-top: 15px;
  		font-size: 0.85em;
  	}
  	.footer a{
			color: #555;
			text-decoration: none;
			font-weight: bold;
  	}
  	.footer a:hover {
  		color: #5cbaf2;
  	}
  </style>
</head>
<body>
	<div class="container">
  	<h1>URL Shortener</h1>
  	<p>Your short URL has been generated!</p>
  	<div class="box">
  		<a href="` + short_url + `" id="url">` + short_url + `</a>
  	</div>
    <div class="footer">
      Created by Earl Balai<br />
      Powered by the <a href="http://golang.org">Go</a> language
    </div>
  </div>
</body>
</html>
	`
	if extension {
		WriteContent(w, short_url)
	} else {
		WriteContent(w, result)
	}
}

func navigate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["code"]

	//Check db for ID
	db := OpenDB()

	var url string

	row := db.QueryRow("SELECT url FROM storage WHERE code = $1", id)

	err := row.Scan(&url)
	if err != nil {
		http.Error(w, "404 url not found", http.StatusNotFound)
	}

	defer db.Close()

	if !strings.Contains(url, "http://") {
		url = "http://" + url
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

//Extension
func extView(w http.ResponseWriter, r *http.Request) {
	createShortlink(w, r.FormValue("url"), true)
}
