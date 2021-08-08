package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	intV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// check cancel work.
// check secret and file are nothing, secret is nothing yet.
// check nothing file error, but loop continue.
func TestWatchWillBeDoneWhenContextCancel(t *testing.T) {

	timeout := time.After(3 * time.Second)
	done := make(chan bool)

	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	client := fake.NewSimpleClientset()
	sclient := client.CoreV1().Secrets("testnamesapce")
	fixture := WatchConfig{
		WatchIntervalSeconds: 120,
		Name:                 "test",
		Namespace:            "testnamesapce",
		SecretType:           "Opaque",
		SecretPath:           "/tmp/hoge",
	}
	go func() {
		watch(btc, fixture, sclient)
		done <- true
	}()

	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

	// assert secret is emty
	_, err := client.CoreV1().Secrets("testnamesapce").Get(ctx, "test", metav1.GetOptions{})

	if !apiErrors.IsNotFound(err) {
		t.Fatal("this secret must be empty")
	}
}

func TestSecretCreatedWhenSecretIsEmptyAndFileExists(t *testing.T) {

	timeout := time.After(3 * time.Second)
	done := make(chan bool)

	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	client := fake.NewSimpleClientset()
	sclient := client.CoreV1().Secrets("testnamesapce")

	data := map[string]string{"key1": "val1", "key2": "val2"}
	path := mkSecretFile(data)
	fixture := WatchConfig{
		WatchIntervalSeconds: 10,
		Name:                 "test2",
		Namespace:            "testnamesapce",
		SecretType:           "Opaque",
		SecretPath:           path,
	}
	go func() {
		watch(btc, fixture, sclient)
		done <- true
	}()

	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

	act, err := client.CoreV1().Secrets("testnamesapce").Get(ctx, "test2", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if act.Name != "test2" {
		t.Fatalf("secret name registration invalid: %s<=>%s", act.Name, "test2")
	}
	if act.Namespace != "testnamesapce" {
		t.Fatalf("secret namesace registration invalid: %s<=>%s", act.Namespace, "testnamesapce")
	}
	if act.Type != v1.SecretTypeOpaque {
		t.Fatalf("secret type registration invalid: %s<=>%s", act.Type, "Opaque")
	}
	if !reflect.DeepEqual(act.StringData, data) {
		t.Fatalf("secret data registration invalid: %s<=>%s", act.StringData, data)
	}
}

func TestSecretNotUpdateWhenSecretAndFileAreSame(t *testing.T) {

	timeout := time.After(3 * time.Second)
	done := make(chan bool)

	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	client := fake.NewSimpleClientset()
	sclient := client.CoreV1().Secrets("testnamesapce")

	data := map[string]string{"key1": "val1", "key2": "val2"}
	path := mkSecretFile(data)
	fixture := WatchConfig{
		WatchIntervalSeconds: 5,
		Name:                 "test3",
		Namespace:            "testnamesapce",
		SecretType:           "Opaque",
		SecretPath:           path,
	}
	createSecret(sclient, &fixture, data)
	go func() {
		watch(btc, fixture, sclient)
		done <- true
	}()

	time.Sleep(300 * time.Millisecond)
	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

	act, err := client.CoreV1().Secrets("testnamesapce").Get(ctx, "test3", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if act.Name != "test3" {
		t.Fatalf("secret name registration invalid: %s<=>%s", act.Name, "test3")
	}
	if act.Namespace != "testnamesapce" {
		t.Fatalf("secret namesace registration invalid: %s<=>%s", act.Namespace, "testnamesapce")
	}
	if act.Type != v1.SecretTypeOpaque {
		t.Fatalf("secret type registration invalid: %s<=>%s", act.Type, "Opaque")
	}
	if !reflect.DeepEqual(act.StringData, data) {
		t.Fatalf("secret data registration invalid: %s<=>%s", act.StringData, data)
	}
}

func TestSecretUpdateWhenSecretAndFileAreDifferent(t *testing.T) {

	timeout := time.After(3 * time.Second)
	done := make(chan bool)

	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	client := fake.NewSimpleClientset()
	sclient := client.CoreV1().Secrets("testnamesapce")

	data := map[string]string{"key1": "val1", "key3": "val3"}
	path := mkSecretFile(data)
	fixture := WatchConfig{
		WatchIntervalSeconds: 5,
		Name:                 "test4",
		Namespace:            "testnamesapce",
		SecretType:           "Opaque",
		SecretPath:           path,
	}
	createSecret(sclient, &fixture, map[string]string{"key1": "val1", "key2": "val2"})
	go func() {
		watch(btc, fixture, sclient)
		done <- true
	}()

	time.Sleep(300 * time.Millisecond)
	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

	act, err := client.CoreV1().Secrets("testnamesapce").Get(ctx, "test4", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if act.Name != "test4" {
		t.Fatalf("secret name registration invalid: %s<=>%s", act.Name, "test4")
	}
	if act.Namespace != "testnamesapce" {
		t.Fatalf("secret namesace registration invalid: %s<=>%s", act.Namespace, "testnamesapce")
	}
	if act.Type != v1.SecretTypeOpaque {
		t.Fatalf("secret type registration invalid: %s<=>%s", act.Type, "Opaque")
	}
	if !reflect.DeepEqual(act.StringData, data) {
		t.Fatalf("secret data registration invalid: %s<=>%s", act.StringData, data)
	}
}

func TestSecretUpdateWhenSecretAndFileAreDifferentPattern2(t *testing.T) {

	timeout := time.After(3 * time.Second)
	done := make(chan bool)

	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	client := fake.NewSimpleClientset()
	sclient := client.CoreV1().Secrets("testnamesapce")

	data := map[string]string{"key1": "val1", "key2": "val2Update"}
	path := mkSecretFile(data)
	fixture := WatchConfig{
		WatchIntervalSeconds: 5,
		Name:                 "test4",
		Namespace:            "testnamesapce",
		SecretType:           "Opaque",
		SecretPath:           path,
	}
	createSecret(sclient, &fixture, map[string]string{"key1": "val1", "key2": "val2"})
	go func() {
		watch(btc, fixture, sclient)
		done <- true
	}()

	time.Sleep(300 * time.Millisecond)
	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

	act, err := client.CoreV1().Secrets("testnamesapce").Get(ctx, "test4", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if act.Name != "test4" {
		t.Fatalf("secret name registration invalid: %s<=>%s", act.Name, "test4")
	}
	if act.Namespace != "testnamesapce" {
		t.Fatalf("secret namesace registration invalid: %s<=>%s", act.Namespace, "testnamesapce")
	}
	if act.Type != v1.SecretTypeOpaque {
		t.Fatalf("secret type registration invalid: %s<=>%s", act.Type, "Opaque")
	}
	if !reflect.DeepEqual(act.StringData, data) {
		t.Fatalf("secret data registration invalid: %s<=>%s", act.StringData, data)
	}

}

func TestDockerConfigJson(t *testing.T) {

	timeout := time.After(5 * time.Second)
	done := make(chan bool)

	ctx := context.Background()
	btc, canecl := context.WithCancel(ctx)
	client := fake.NewSimpleClientset()
	sclient := client.CoreV1().Secrets("testnamesapce")

	data := map[string]string{"docker-server": "private.example.com:5050", "docker-username": "user", "docker-password": "pass"}
	path := mkSecretFile(data)
	fixture := WatchConfig{
		WatchIntervalSeconds: 1,
		Name:                 "docker-cred",
		Namespace:            "testnamesapce",
		SecretType:           "kubernetes.io/dockerconfigjson",
		SecretPath:           path,
	}
	go func() {
		watch(btc, fixture, sclient)
		done <- true
	}()

	time.Sleep(3000 * time.Millisecond)
	canecl()

	select {
	case <-timeout:
		t.Fatal("Watch method not finished by cancel")
	case <-done:
	}

	act, err := client.CoreV1().Secrets("testnamesapce").Get(ctx, "docker-cred", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if act.Type != v1.SecretTypeDockerConfigJson {
		t.Fatalf("secret type registration invalid: %s<=>%s", act.Type, "kubernetes.io/dockerconfigjson")
	}
	dockercred, ok := act.StringData[".dockerconfigjson"]
	if !ok {
		t.Fatalf(".dockerconfigjson not exists")
	}

	cred := make(map[string]map[string]map[string]string)
	if json.Unmarshal([]byte(dockercred), &cred); err != nil {
		t.Fatal("dockerconfigjson is not json")
	}
	auth, ok := cred["auths"]["https://private.example.com:5050"]
	if !ok {
		t.Fatal("authz/https://private.example.com:5050 not exists")
	}
	if auth["username"] != "user" {
		t.Fatalf("dockerconfig  username invalid: %s", auth["username"])
	}
	if auth["password"] != "pass" {
		t.Fatalf("dockerconfig  username invalid: %s", auth["password"])
	}
	base64auth := base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if auth["auth"] != base64auth {
		t.Fatalf("dockerconfig  username invalid: %s <=> %s ", auth["username"], base64auth)
	}
}

func createSecret(client intV1.SecretInterface, watch *WatchConfig, data map[string]string) {
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      watch.Name,
			Namespace: watch.Namespace,
		},
		Type:       v1.SecretType(watch.SecretType),
		StringData: data,
	}
	_, err := client.Create(context.TODO(), &secret, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func mkSecretFile(data map[string]string) string {

	f, err := ioutil.TempFile(os.TempDir(), "sec")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	bytes, err := yaml.Marshal(data)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(bytes)

	if err != nil {
		panic(err)
	}
	return f.Name()
}
