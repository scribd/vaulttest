# vaulttest

[![Circle CI](https://circleci.com/gh/scribd/vaulttest.svg?style=shield)](https://circleci.com/gh/scribd/vaulttest)

Library for spinning up test instances of Hashicorp Vault for use in integration tests locally and in CI systems.

Hashicorp Vault is cool, and can do lots of shtuff, it even has a nice UI, but *managing* it really needs more than pointing and clicking in a UI, or running vault commands against the server.

A much better way is to write some code that instruments your Vault in a predictable manner.

But how do you *test* said code?  You really need a test Vault or better yet a fleet of them to test changes in parallel.

Unfortunately Hashicorp Vault's source code is not organized/ exported in a way to make it's internal api easily adapted to a fully code defined, in memory Vault Dev server.

So instead, we have this package which will spin one up- provided the `vault` binary is on the system somewhere.

We'll test the system to find a free port, spin up vault in dev mode on that port, do our tests against it, and shut it down politely once we're done.  Voila!

# Prerequisites

* Hashicorp Vault, installed on your system somewhere in the PATH.  https://www.vaultproject.io/downloads.html

* This library: `go get github.com/scribd/vaulttest`

# Usage

Include the following in your test code:

    var testServer *vaulttest.VaultDevServer
    var testClient *api.Client

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

        testAddress := fmt.Sprintf("127.0.0.1:%d", port)

        testServer = vaulttest.NewVaultDevServer(testAddress)

        if !testServer.Running {
            testServer.ServerStart()
            client := testServer.VaultTestClient()

            // set up some secret engines
            for _, endpoint := range []string{
                "prod",
                "stage",
                "dev",
            } {
                data := map[string]interface{}{
                    "type":        "kv-v2",
                    "description": "Production Secrets",
                }
                _, err := client.Logical().Write(fmt.Sprintf("sys/mounts/%s", endpoint), data)
                if err != nil {
                    log.Fatalf("Unable to create secret engine %q: %s", endpoint, err)
                }
            }

            // setup a PKI backend
            data := map[string]interface{}{
                "type":        "pki",
                "description": "PKI backend",
            }
            
            _, err := client.Logical().Write("sys/mounts/pki", data)
            if err != nil {
                log.Fatalf("Failed to create pki secrets engine: %s", err)
            }

            data = map[string]interface{}{
                "common_name": "test-ca",
                "ttl":         "43800h",
            }
            
            _, err = client.Logical().Write("pki/root/generate/internal", data)
            if err != nil {
                log.Fatalf("Failed to create root cert: %s", err)
            }

            data = map[string]interface{}{
                "max_ttl":         "24h",
                "ttl":             "24h",
                "allow_ip_sans":   true,
                "allow_localhost": true,
                "allow_any_name":  true,
            }
            
            _, err = client.Logical().Write("pki/roles/foo", data)
            if err != nil {
                log.Fatalf("Failed to create cert issuing role: %s", err)
            }

            data = map[string]interface{}{
                "type":        "cert",
                "description": "TLS Cert Auth endpoint",
            }

            _, err = client.Logical().Write("sys/auth/cert", data)
            if err != nil {
                log.Fatalf("Failed to enable TLS cert auth: %s", err)
            }
            
            ... Do other setup stuff ...
            
            testClient = client
        }
    }

    func tearDown() {
        if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
            os.Remove(tmpDir)
        }

        testServer.ServerShutDown()
    }
    
    func TestNewNamespace(t *testing.T) {
        path := "dev/foo/bar"
        secret, err := testClient.Logical().Read(path)
        if err != nil {
            log.Printf("Unable to read %q: %s\n", path, err)
            t.Fail()
        }
        
        if secret == nil {
            log.Print("Nil Secret")
            t.fail() 
        }
        
        assert.True(t, secret.Data["foo"].(string) == "bar", "Successfully returned secret")
    }
