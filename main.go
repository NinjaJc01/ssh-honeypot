package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type loginData struct {
	username      string
	password      string
	remoteIP      string
	remoteVersion string
}

func main() {
	var (
		lport       uint
		lhost       string
		keyPath     string
		fingerprint string
		fakeShell   bool
	)
	flag.UintVar(&lport, "p", 2222, "Set the local port to listen for SSH on")
	flag.StringVar(&lhost, "i", "0.0.0.0", "The IP address to listen on")
	flag.StringVar(&keyPath, "k", "id_rsa", "Set a path to the private key to use for the SSH server")
	flag.BoolVar(&fakeShell, "shell", false, "Set true to lock the user in a fake shell")
	flag.StringVar(&fingerprint, "f", "OpenSSH_8.2p1 Debian-4", "Set the fingerprint of the SSH server, exclude the 'SSH-2.0-' prefix")
	flag.Parse()
	log.SetPrefix("SSH - ")
	privKeyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Panicln("Error reading privkey:\t", err.Error())
	}
	privateKey, err := gossh.ParsePrivateKey(privKeyBytes)
	if err != nil {
		log.Panicln("Error parsing privkey:\t", err.Error())
	}
	server := &ssh.Server{
		Addr: fmt.Sprintf("%s:%v", lhost, lport),
		Handler: func() ssh.Handler {
			if !fakeShell {
				return sshHandler
			}
			return fakeTerminal
		}(),
		Version:         fingerprint,
		PasswordHandler: passwordHandler,
	}
	server.AddHostKey(privateKey)
	log.Println("Started Honeypot SSH server on", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func logData(data loginData) {
	file, err := os.OpenFile("loot.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err.Error())
	}
	defer file.Close()
	loginString := fmt.Sprintf("%s,%s,%s,%s\n", data.username, data.password, data.remoteIP, data.remoteVersion)
	if _, err := file.WriteString(loginString); err != nil {
		log.Println(err.Error())
	}
}

func sshHandler(s ssh.Session) {
	s.Write([]byte("Don't blindly SSH into every VM you see.\n"))
}

func fakeTerminal(s ssh.Session) {
	term := terminal.NewTerminal(s, fmt.Sprintf("%s@kali:~$ ", s.User()))
	for {
		commandLine, _ := term.ReadLine()
		commandLineSlice := strings.Split(commandLine, " ")
		if commandLineSlice[0] == "exit" {
			break
		}
		if commandLineSlice[0] != "" {
			term.Write([]byte(fmt.Sprintf("bash: %s: command not found\n", commandLineSlice[0])))
		}
	}
}

func passwordHandler(context ssh.Context, password string) bool {
	details := fmt.Sprintf(
		"Got Login!\nUsername:\t%s\nPassword:\t%s\nRemote:\t%s\nClient:\t%s\n",
		context.User(), password, context.RemoteAddr(), context.ClientVersion())
	log.Println(details)
	data := loginData{
		username:      context.User(),
		password:      password,
		remoteIP:      context.RemoteAddr().String(),
		remoteVersion: context.ClientVersion()}
	logData(data)
	return true
}
