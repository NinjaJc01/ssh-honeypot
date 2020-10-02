package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/integrii/flaggy"
	_ "github.com/mattn/go-sqlite3"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	database         *sql.DB
	doCommandLogging bool   = false
	hostname         string = "kali"
	message          string = "Don't blindly SSH into every VM you see."
	lootChan         chan (loginData)
	cmdChan          chan (command)
)

const buffSize = 5

type loginData struct {
	username      string
	password      string
	remoteIP      string
	remoteVersion string
	timestamp     string
}

type command struct {
	username  string
	remoteIP  string
	command   string
	timestamp string
}

func main() {
	var (
		lport       uint   = 2222
		lhost       net.IP = net.ParseIP("0.0.0.0")
		keyPath     string = "id_rsa"
		fingerprint string = "OpenSSH_8.2p1 Debian-4"
	)
	databasePointer, err := sql.Open("sqlite3", "honeypot.db")
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Database connection failed")
	}
	database = databasePointer
	lootChan = make(chan (loginData), buffSize)
	cmdChan = make(chan (command), buffSize)
	flaggy.UInt(&lport, "p", "port", "Local port to listen for SSH on")
	flaggy.IP(&lhost, "i", "interface", "IP address for the interface to listen on")
	flaggy.String(&keyPath, "k", "key", "Path to private key for SSH server")
	flaggy.String(&fingerprint, "f", "fingerprint", "SSH Fingerprint, excluding the SSH-2.0- prefix")

	fakeShellSubcommand := flaggy.NewSubcommand("fakeshell")
	fakeShellSubcommand.String(&hostname, "H", "hostname", "Hostname for fake shell prompt")
	fakeShellSubcommand.Bool(&doCommandLogging, "C", "logcmd", "Log user commands within the fake shell?")
	warnSubcommand := flaggy.NewSubcommand("warn")
	warnSubcommand.String(&message, "m", "message", "Warning message to be sent after authentication")

	flaggy.AttachSubcommand(fakeShellSubcommand, 1)
	flaggy.AttachSubcommand(warnSubcommand, 1)
	flaggy.Parse()
	if !fakeShellSubcommand.Used && !warnSubcommand.Used {
		flaggy.ShowHelpAndExit("No subcommand supplied")
	}
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
	go threadsafeLootLogger()
	log.Println("Started loot logger")
	if doCommandLogging {
		go threadsafeCommandLogger()
		log.Println("Started command logger")
	}
	log.Println("Started Honeypot SSH server on", server.Addr)
	log.Fatal(server.ListenAndServe())
}

// These functions ensure only one thread is writing to the DB at once.
// Each handler runs in parallel so we cannot write from that thread safely.
func threadsafeLootLogger() {
	for {
		logLoot(<-lootChan)
	}
}

func threadsafeCommandLogger() {
	for {
		logCommand(<-cmdChan)
	}
}

func logCommand(cmd command) {
	statement, err := database.Prepare(
		"INSERT INTO Command(Username, RemoteIP, Command, Timestamp) values(?,?,?,?)")
	if err != nil {
		log.Println(err.Error())
	}
	_, err = statement.Exec(
		cmd.username,
		cmd.remoteIP,
		cmd.command,
		cmd.timestamp)
	if err != nil {
		log.Println(err.Error())
	}
}

func logLoot(data loginData) { //TODO quoted string username
	statement, err := database.Prepare(
		"INSERT INTO Login(Username, Password, RemoteIP, RemoteVersion, Timestamp) values(?,?,?,?,?)")
	if err != nil {
		log.Println(err.Error())
	}
	_, err = statement.Exec(
		data.username,
		data.password,
		data.remoteIP,
		data.remoteVersion,
		data.timestamp)
	if err != nil {
		log.Println(err.Error())
	}
}

func sshHandler(s ssh.Session) {
	s.Write([]byte(message + "\n"))
}

func fakeTerminal(s ssh.Session) {
	commandLine := s.RawCommand()
	if s.RawCommand() != "" { //If the attacker sets a command with ssh -C
		cmdChan <- command{
			username:  s.User(),
			remoteIP:  s.RemoteAddr().String(),
			command:   commandLine,
			timestamp: fmt.Sprint(time.Now().Unix())}
	}
	term := terminal.NewTerminal(s, fmt.Sprintf("%s@%s:~$ ", s.User(), hostname))
	go func(s ssh.Session) { //timeout sessions to save CPU.
		time.Sleep(time.Second * 30)
		s.Close()
	}(s)
	for {
		commandLine, err := term.ReadLine()
		if err != nil {
			s.Close()
			break
		}
		commandLineSlice := strings.Split(commandLine, " ")
		if commandLineSlice[0] == "exit" {
			break
		}
		if commandLineSlice[0] != "" {
			if doCommandLogging {
				cmdChan <- command{
					username:  s.User(),
					remoteIP:  s.RemoteAddr().String(),
					command:   commandLine,
					timestamp: fmt.Sprint(time.Now().Unix())}
			}
			term.Write([]byte(fmt.Sprintf("bash: %s: command not found\n", commandLineSlice[0])))
		}
	}
	s.Close()
}

func passwordHandler(context ssh.Context, password string) bool {
	data := loginData{
		username:      context.User(),
		password:      password,
		remoteIP:      context.RemoteAddr().String(),
		remoteVersion: context.ClientVersion(),
		timestamp:     fmt.Sprint(time.Now().Unix())}
	//logLoot(data)
	lootChan <- data
	return true
}
