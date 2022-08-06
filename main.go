package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
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

func reqDebug(req *http.Request) debugFormat {
	debugLog := debugFormat{}
	debugLog.Method = req.Method
	debugLog.Url = req.URL
	debugLog.Header = req.Header
	debugLog.Host = req.Host
	debugLog.Form = req.Form
	debugLog.PostForm = req.PostForm
	debugLog.Trailer = req.Trailer
	debugLog.RemoteAddr = req.RemoteAddr
	debugLog.RequestUri = req.RequestURI
	debugLog.Tls = req.TLS
	return debugLog
}

func debug(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		marshal, err := json.Marshal(reqDebug(req))
		if err != nil {
			log.Fatalf("json format error: %s", err)
		}
		fmt.Println(string(marshal))
		endpoint(w, req)
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

func delay(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query().Get("duration")
	duration, err := time.ParseDuration(query)

	if err != nil {
		duration = 5 * time.Second
		if query == "" {
			_, _ = fmt.Fprintf(w, "duration is not found.\n")
		} else {
			_, _ = fmt.Fprintf(w, "invalid format: %s\n", query)
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(duration)
	}()
	wg.Wait()

	_, _ = fmt.Fprintf(w, "waited for %s\n", duration.String())
}

func main() {
	http.HandleFunc("/", health)
	http.HandleFunc("/hello", debug(hello))
	http.HandleFunc("/headers", debug(headers))
	http.HandleFunc("/delay", debug(delay))

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "80"
	}
	fmt.Printf("Listen :%s\n", httpPort)

	_ = http.ListenAndServe(":"+httpPort, nil)
}
