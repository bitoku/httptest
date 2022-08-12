package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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

type settings struct {
	addr            string
	idleTimeout     time.Duration
	enableKeepAlive bool
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

func enableKeepAlive() bool {
	enableKeepAliveEnv := os.Getenv("ENABLE_KEEP_ALIVE")
	if enableKeepAliveEnv != "" {
		enableKeepAlive, err := strconv.ParseBool(enableKeepAliveEnv)
		if err != nil {
			_ = fmt.Errorf("invalid ENABLE_KEEP_ALIVE: %s\n", enableKeepAliveEnv)
			enableKeepAlive = true
		}
		return enableKeepAlive
	}
	return true
}

func idleTimeout() time.Duration {
	idleTimeoutEnv := os.Getenv("IDLE_TIMEOUT")
	if idleTimeoutEnv != "" {
		duration, err := time.ParseDuration(idleTimeoutEnv)
		if err != nil {
			_ = fmt.Errorf("invalid IDLE_TIMEOUT: %s\n", idleTimeoutEnv)
			duration = 0
		}
		return duration
	}
	return 0
}

func addr() string {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "80"
	}
	return ":" + httpPort
}

func debug(endpoint func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		marshal, err := json.Marshal(reqDebug(req))
		if err != nil {
			_ = fmt.Errorf("json format error: %s", err)
		}
		fmt.Println(string(marshal))
		endpoint(w, req)
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	_, _ = fmt.Fprintf(w, "ok\n")
}

func hello(w http.ResponseWriter, _ *http.Request) {
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

func httpError(w http.ResponseWriter, req *http.Request) {
	codeQuery := req.URL.Query().Get("code")
	code, err := strconv.Atoi(codeQuery)
	if err != nil {
		_, _ = fmt.Fprintf(w, "invalid code: %s\n", codeQuery)
		code = 400
	}
	message := req.URL.Query().Get("message")
	if message == "" {
		message = "error"
	}
	http.Error(w, message, code)
}

func info(conf *settings) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "addr: %v\n", conf.addr)
		_, _ = fmt.Fprintf(w, "enable keep alive: %v\n", conf.enableKeepAlive)
		_, _ = fmt.Fprintf(w, "idle timeout: %v\n", conf.idleTimeout)
	}
}

func closeCon(server *http.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := server.Close()
		if err != nil {
			// TODO
			return
		}
	}
}

func addHeaders(w http.ResponseWriter, req *http.Request) {
	queries := req.URL.Query()

	for key, values := range queries {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	for name, headers := range w.Header() {
		for _, h := range headers {
			_, _ = fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	conf := &settings{
		addr:            addr(),
		idleTimeout:     idleTimeout(),
		enableKeepAlive: enableKeepAlive(),
	}

	server := &http.Server{
		Addr:        conf.addr,
		IdleTimeout: conf.idleTimeout,
	}

	server.SetKeepAlivesEnabled(conf.enableKeepAlive)

	http.HandleFunc("/", health)
	http.HandleFunc("/info", info(conf))
	http.HandleFunc("/hello", debug(hello))
	http.HandleFunc("/headers", debug(headers))
	http.HandleFunc("/delay", debug(delay))
	http.HandleFunc("/error", debug(httpError))
	http.HandleFunc("/close", debug(closeCon(server)))
	http.HandleFunc("/addheaders", debug(addHeaders))

	fmt.Println("Start Server")

	log.Fatal(server.ListenAndServe())
}
