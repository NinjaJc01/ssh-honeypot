go get -u "github.com/gliderlabs/ssh"
go get -u "golang.org/x/crypto/ssh"
go get -u "golang.org/x/crypto/ssh/terminal"
ssh-keygen -f ./id_rsa
go build -o server main.go