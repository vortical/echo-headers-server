package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8001, "port to run on (defaults to 8001)")
	flag.Parse()

	url := fmt.Sprintf("0.0.0.0:%d", *port)
	fmt.Printf("starting %s ....\n", url)

	http.HandleFunc("/headers", handler)
	log.Fatal(http.ListenAndServe(url, nil))
}

func handler(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(writer, "Method[%s] Url[%s] Protocol[%s]\n", req.Method, req.URL.String(), req.Proto)
	for key, value := range req.Header {
		fmt.Fprintf(writer, "Header[%q] = %q\n", key, value)
	}
	fmt.Fprintf(writer, "Host[%s]\n", req.Host)
	fmt.Fprintf(writer, "RemoteAddr[%s]\n", req.RemoteAddr)

	if err := req.ParseForm(); err != nil {
		log.Print(err)
	}

	for key, value := range req.Form {
		fmt.Fprintf(writer, "Form[%q] = %q\n", key, value)
	}

	fmt.Fprintf(writer, "URL path: %q\n", req.URL.Path)
}
