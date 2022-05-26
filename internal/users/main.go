package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	handler := http.HandlerFunc(SayHalloHandler)

	fmt.Printf("Starting users service at port: %v", os.Getenv("PORT"))
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), handler); err != nil {
		panic(err)
	}
}

func SayHalloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from users service")
}