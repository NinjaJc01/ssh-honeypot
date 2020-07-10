package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/integrii/flaggy"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	hostname string = "kali"
	message  string = "Don't blindly SSH into every VM you see."
)

type loginData struct {
	username      string
	password      string
	remoteIP      string
	remoteVersion string
}

func main() {
	var (
		lport       uint   = 2222
		lhost       net.IP = net.ParseIP("0.0.0.0")
		keyPath     string = "id_rsa"
		fingerprint string = "OpenSSH_8.2p1 Debian-4"
	)

	flaggy.UInt(&lport, "p", "port", "Local port to listen for SSH on")
	flaggy.IP(&lhost, "i", "interface", "IP address for the interface to listen on")
	flaggy.String(&keyPath, "k", "key", "Path to private key for SSH server")
	flaggy.String(&fingerprint, "f", "fingerprint", "")

	fakeShellSubcommand := flaggy.NewSubcommand("fakeshell")
	fakeShellSubcommand.String(&hostname, "H", "hostname", "Hostname for fake shell prompt")
	warnSubcommand := flaggy.NewSubcommand("warn")
	warnSubcommand.String(&message, "m", "message", "Warning message to be sent after authentication")
	
	flaggy.AttachSubcommand(fakeShellSubcommand,1)
	flaggy.AttachSubcommand(warnSubcommand,1)
	flaggy.Parse()
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
		Addr: fmt.Sprintf("%s:%v", lhost.String(), lport),
		Handler: func() ssh.Handler {
			if warnSubcommand.Used {
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
	s.Write([]byte(message + "\n"))
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
