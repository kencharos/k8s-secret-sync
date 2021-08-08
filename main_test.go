package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// check not raise panic when config valid.
func TestMainConfigParseProcess(t *testing.T) {

	config := `---
defaultWatchIntervalSeconds: 1
inClusterMode: false
externalClusterBearerToken: "token"
externalClusterHost: "https://localhost:5050/api"
watch:
- name: test
  namespace: testns
  type: Opaque
  secretPath: /tmp/sample.yml
`

	configPath := mkYamlFile(config)

	os.Args = []string{"test", configPath}
	go func() { main() }()
	time.Sleep(1 * time.Second)
}

func mkYamlFile(raw string) string {

	f, err := ioutil.TempFile(os.TempDir(), "sec")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(raw)

	if err != nil {
		panic(err)
	}
	return f.Name()
}
