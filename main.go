package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type debugFormat struct {
	Method     string
	Url        any
	Header     any
	Host       string
	Form       any
	PostForm   any
	Trailer    any
	RemoteAddr string
	RequestUri string
	Tls        any
}

func reqDebug(r *http.Request) debugFormat {
	debugLog := debugFormat{}
	debugLog.Method = r.Method
	debugLog.Url = r.URL
	debugLog.Header = r.Header
	debugLog.Host = r.Host
	debugLog.Form = r.Form
	debugLog.PostForm = r.PostForm
	debugLog.Trailer = r.Trailer
	debugLog.RemoteAddr = r.RemoteAddr
	debugLog.RequestUri = r.RequestURI
	debugLog.Tls = r.TLS
	return debugLog
}

func debug(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		marshal, err := json.Marshal(reqDebug(r))
		if err != nil {
			log.Fatalf("json format error: %s", err)
		}
		fmt.Println(string(marshal))
		endpoint(w, r)
	}
}

func health(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "ok\n")
}

func hello(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			_, _ = fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	http.HandleFunc("/", health)
	http.HandleFunc("/hello", debug(hello))
	http.HandleFunc("/headers", debug(headers))

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "80"
	}
	fmt.Printf("Listen :%s\n", httpPort)

	_ = http.ListenAndServe(":"+httpPort, nil)
}
