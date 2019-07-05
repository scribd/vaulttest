package vaultintegtest

import (
	"bufio"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type TestServer struct {
	Command       *exec.Cmd
	Running       bool
	UnsealKey     string
	RootToken     string
	UserToken     string
	UserTokenFile string
	Address       string
}

func NewTestServer(address string) *TestServer {
	if address == "" {
		address = "127.0.0.1:8200"
	}

	testServer := TestServer{
		Address: address,
	}

	return &testServer
}

func (t *TestServer) Start() {
	// find the user's vault token file if it exists
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Unable to determine user's home dir: %s", err)
	}

	t.UserTokenFile = fmt.Sprintf("%s/.vault-token", homeDir)

	// read it into memory cos the test server is gonna overwrite it
	if _, err := os.Stat(t.UserTokenFile); !os.IsNotExist(err) {
		tokenBytes, err := ioutil.ReadFile(t.UserTokenFile)
		if err == nil {
			t.UserToken = string(tokenBytes)
		}
	}

	vault, err := exec.LookPath("vault")
	if err != nil {
		log.Fatal("'vault' is not installed and available on the path")
	}

	t.Command = exec.Command(vault, "server", "-dev", "-address", t.Address)

	t.Command.Stderr = os.Stderr
	out, err := t.Command.StdoutPipe()
	if err != nil {
		log.Fatalf("unable to connect to testserver's stdout: %s", err)
	}

	err = t.Command.Start()

	scanner := bufio.NewScanner(out)

	unsealPattern := regexp.MustCompile(`^Unseal Key:.+`)
	rootTokenPattern := regexp.MustCompile(`^Root Token:.+`)

	for t.UnsealKey == "" || t.RootToken == "" {
		scanner.Scan()
		line := scanner.Text()

		if t.UnsealKey == "" && unsealPattern.MatchString(line) {
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				t.UnsealKey = parts[1]
				strings.TrimRight(t.UnsealKey, "\n")
				strings.TrimLeft(t.UnsealKey, " ")
			}

			continue
		}

		if t.RootToken == "" && rootTokenPattern.MatchString(line) {
			parts := strings.Split(line, ": ")
			if len(parts) > 1 {
				t.RootToken = parts[1]
				strings.TrimRight(t.RootToken, "\n")
				strings.TrimLeft(t.RootToken, " ")
			}

			continue
		}

	}

	t.Running = true
}

func (t *TestServer) ShutDown() {
	if t.Running {
		t.Command.Process.Kill()
	}

	// restore the user's vault token when we're done.
	if t.UserToken != "" {
		_ = ioutil.WriteFile(t.UserTokenFile, []byte(t.UserToken), 0600)
	}
}

// VaultTestClient returns a configured vault client for the test vault server.  By default the client returned has the root token for the test vault instance set.  If you want something else, you will need to reconfigure it.
func (t *TestServer) VaultClient() *api.Client {
	config := api.DefaultConfig()

	err := config.ReadEnvironment()
	if err != nil {
		log.Fatalf("failed to inject environment into test vault client config")
	}

	config.Address = "http://127.0.0.1:8200"

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("failed to create test vault api client: %s", err)
	}

	client.SetToken(t.RootToken)

	return client
}
