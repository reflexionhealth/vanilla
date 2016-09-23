package httpx

import (
	"fmt"
	"log"
	"net/http"
)

func ExampleMux_main() {
	index := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome!\n")
	}
	hello := func(w http.ResponseWriter, r *http.Request) {
		ps:= GetParams(r.Context())
		fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
	}

	mux := NewMux()
	mux.GET("/", index)
	mux.GET("/hello/:name", hello)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
