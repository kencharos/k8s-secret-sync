package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// check k8s secret and local file periodically. If these is different, replace k8s secret.
func watch(ctx context.Context, watch WatchConfig, secretClient intV1.SecretInterface) {

	getOpt := metav1.GetOptions{}

	initSecret, err := secretClient.Get(ctx, watch.Name, getOpt)
	cachedData := make(map[string][]byte)
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			log.Errorf("unexpected error when get secret of %s/%s, but continute goroutine..", watch.Namespace, watch.Name)
			log.Error(err)
		}
		log.Debugf("secret %s/%s does not exists", watch.Namespace, watch.Name)
	} else {
		log.Debugf("secret %s/%s already exists", watch.Namespace, watch.Name)
		// notice! StringData is write only. empty.
		cachedData = initSecret.Data
	}

	log.Infof("start watch %s/%s from %s with interval %d sec", watch.Namespace, watch.Name, watch.SecretPath, watch.WatchIntervalSeconds)
	for {

		log.Debugf("start watch loop %s/%s from %s with interval %d sec", watch.Namespace, watch.Name, watch.SecretPath, watch.WatchIntervalSeconds)

		fileData, err := readSecretFile(&watch)
		if err != nil {
			log.Errorf("unexpected error on read file but continue next loop, %s/%s from %s", watch.Namespace, watch.Name, watch.SecretPath)
			log.Error(err)
		} else {
			if !reflect.DeepEqual(cachedData, fileData) {
				err = replaceSecret(ctx, secretClient, &watch, fileData, len(cachedData) == 0)
				if err != nil {
					log.Errorf("unexpected error on replace secret but continue next loop, %s/%s from %s", watch.Namespace, watch.Name, watch.SecretPath)
					log.Error(err)
				} else {
					log.Infof("secret %s/%s type=%s was replaced!", watch.Namespace, watch.Name, watch.SecretType)
					cachedData = fileData
				}
			}
		}
		log.Debugf("end watch loop %s/%s from %s, next loop after %d sec", watch.Namespace, watch.Name, watch.SecretPath, watch.WatchIntervalSeconds)

		select {
		case <-ctx.Done():
			log.Warnf("watch %s/%s: canceld", watch.Namespace, watch.Name)
			return
		case <-time.After(time.Duration(watch.WatchIntervalSeconds) * time.Second):
			continue
		}
	}

}

func readSecretFile(watch *WatchConfig) (map[string][]byte, error) {
	bytes, err := ioutil.ReadFile(watch.SecretPath)
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)
	err = yaml.Unmarshal(bytes, data)
	if err != nil {
		return nil, err
	}

	converted, err := convertData(watch.SecretType, data)
	if err != nil {
		return nil, err
	}
	return converted, nil
}

func convertData(secretType string, data map[string]string) (map[string][]byte, error) {
	newData := make(map[string][]byte)
	if v1.SecretType(secretType) == v1.SecretTypeDockerConfigJson {
		//generate .dockerconfigjson docker-server, docker-username, docker-password
		dockerServer, ok := data["docker-server"]
		if !ok {
			return nil, errors.New("docker-server key not exist")
		}

		username, ok := data["docker-username"]
		if !ok {
			return nil, errors.New("docker-server key not exist")
		}
		password, ok := data["docker-password"]
		if !ok {
			return nil, errors.New("docker-server key not exist")
		}

		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		dockerconfigjson := fmt.Sprintf(
			"{\"auths\":{\"https://%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}", dockerServer, username, password, auth)
		newData[".dockerconfigjson"] = []byte(dockerconfigjson)
	} else {
		for k, v := range data {
			newData[k] = []byte(v)
		}
	}
	return newData, nil
}

func replaceSecret(ctx context.Context, secretClient intV1.SecretInterface, watch *WatchConfig, newData map[string][]byte, create bool) error {
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      watch.Name,
			Namespace: watch.Namespace,
		},
		Type: v1.SecretType(watch.SecretType),
		Data: newData,
	}
	if create {
		log.Debugf("Create secret %s/%s", watch.Namespace, watch.Name)
		_, err := secretClient.Create(ctx, &secret, metav1.CreateOptions{})
		return err
	} else {
		log.Debugf("Update secret %s/%s", watch.Namespace, watch.Name)
		_, err := secretClient.Update(ctx, &secret, metav1.UpdateOptions{})
		return err
	}
}
