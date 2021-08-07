package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	InCluster                   bool          `yaml:"inCluster"`
	DefalutWatchIntervalSeconds int64         `yaml:"defalutWatchIntervalSeconds"`
	KubeconfigPath              string        `yaml:"kubeconfigPath"`
	Watch                       []WatchConfig `yaml:"watch"`
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
	log.SetLevel(log.InfoLevel)
}

func makeClusterConfig(conf *Config) *kubernetes.Clientset {

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func main() {

	var c = Config{
		InCluster:                   true,
		DefalutWatchIntervalSeconds: 120,
	}

	in, err := ioutil.ReadFile("./example.yaml")
	if err != nil {
		panic(err.Error())
	}
	err = yaml.Unmarshal(in, &c)
	if err != nil {
		panic(err.Error())
	}

	clientSet := makeClusterConfig(&c)

	ctx := context.Background()
	cancelableCtx, cancel := context.WithCancel(ctx)
	for _, w := range c.Watch {
		if w.WatchIntervalSeconds == 0 {
			w.WatchIntervalSeconds = c.DefalutWatchIntervalSeconds
		}
		secretClient := clientSet.CoreV1().Secrets(w.Namespace)
		go watch(cancelableCtx, w, &secretClient)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	s := <-sig
	cancel()
	log.Info("Signal received: %s , cancell goroutines.", s.String())
	//time.Sleep(30 * time.Second)

	//cancel()
	//fmt.Println("goroutine cancell done")
	//time.Sleep(5 * time.Second)

	//for {
	/*
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "example-xxxxx", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod example-xxxxx not found in default namespace\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found example-xxxxx pod in default namespace\n")
		}

		time.Sleep(10 * time.Second)
	*/
	//}
}
