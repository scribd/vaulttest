package vault_test_server

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testServer *VaultDevServer

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	testServer = NewVaultDevServer("")

	if !testServer.Running {
		testServer.Start()
	}
}

func tearDown() {
	testServer.ShutDown()
}

func TestVaultTestClient(t *testing.T) {
	assert.True(t, 1 == 1, "the law of identity has been broken")

	client := testServer.VaultClient()

	secret, err := client.Logical().Read("secret/config")
	if err != nil {
		log.Printf("Failed to default secret config: %s", err)
		t.Fail()
	}

	assert.True(t, secret != nil, "We got a secret from the test vault server")
}
