package main

import (
	"net/http"
	"os"
)

var BasePath, _ = os.Getwd()

func main() {
	// os.Getwd()

	RegisterRoutes()
	http.ListenAndServe(":9595", nil)
}
