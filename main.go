package main

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	DefalutWatchIntervalSeconds   int64         `yaml:"defalutWatchIntervalSeconds"`
	InClusterMode                 bool          `yaml:"inClusterMode"`
	ExternalClusterKubeconfigPath string        `yaml:"externalClusterKubeconfigPath"`
	ExternalClusterHost           string        `yaml:"externalClusterHost"`
	ExternalClusterBearerToken    string        `yaml:"externalClusterBearerToken"`
	Watch                         []WatchConfig `yaml:"watch"`
}

type WatchConfig struct {
	Name                 string `yaml:"name"`
	Namespace            string `yaml:"namespace"`
	SecretType           string `yaml:"type"`
	SecretPath           string `yaml:"secretPath"`
	WatchIntervalSeconds int64  `yaml:"watchIntervalSeconds"`
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	logLevelText, exists := os.LookupEnv("LOG_LEVEL")
	if !exists {
		logLevelText = "info"
	}
	logLevel, err := log.ParseLevel(logLevelText)
	if err != nil {
		panic(err)
	}
	log.SetLevel(logLevel)
}

func main() {

	if len(os.Args) == 1 {
		panic(errors.New("argument filename is required"))
	}

	var c = Config{
		InClusterMode:               true,
		DefalutWatchIntervalSeconds: 120,
	}

	in, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Error(err)
		panic(err.Error())
	}
	err = yaml.Unmarshal(in, &c)
	if err != nil {
		log.Error(err)
		panic(err.Error())
	}

	clientSet, err := makeClusterConfig(&c)
	if err != nil {
		log.Error(err)
		panic(err.Error())
	}

	ctx := context.Background()
	cancelableCtx, cancel := context.WithCancel(ctx)
	for _, w := range c.Watch {
		if w.WatchIntervalSeconds == 0 {
			w.WatchIntervalSeconds = c.DefalutWatchIntervalSeconds
		}
		secretClient := clientSet.CoreV1().Secrets(w.Namespace)
		go watch(cancelableCtx, w, secretClient)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	s := <-sig
	cancel()
	log.Infof("Signal received: %s , cancell goroutines.", s.String())

}

func makeClusterConfig(conf *Config) (*kubernetes.Clientset, error) {

	if conf.InClusterMode {

		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(restConfig)
	} else if len(conf.ExternalClusterKubeconfigPath) > 0 {
		restConfig, err := clientcmd.BuildConfigFromFlags("", conf.ExternalClusterKubeconfigPath)
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(restConfig)
	} else if len(conf.ExternalClusterBearerToken) > 0 {
		restConig := rest.Config{
			Host:        conf.ExternalClusterHost,
			BearerToken: conf.ExternalClusterBearerToken,
		}
		return kubernetes.NewForConfig(&restConig)
	}
	return nil, errors.New("k8s cluster config invalid")
}
