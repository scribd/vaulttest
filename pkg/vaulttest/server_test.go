package vaulttest

import (
	"fmt"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testServer *VaultDevServer
var testAddress string

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatalf("Failed to get a free port on which to run the test vault server: %s", err)
	}

	testAddress = fmt.Sprintf("127.0.0.1:%d", port)

	testServer = NewVaultDevServer(testAddress)

	if !testServer.Running {
		testServer.ServerStart()
	}
}

func tearDown() {
	testServer.ServerShutDown()
}

func TestVaultTestClient(t *testing.T) {
	assert.True(t, 1 == 1, "the law of identity has been broken")

	client := testServer.VaultTestClient()

	secret, err := client.Logical().Read("secret/config")
	if err != nil {
		log.Printf("Failed to default secret config: %s", err)
		t.Fail()
	}

	assert.True(t, secret != nil, "We got a secret from the test vault server")
}
