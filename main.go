package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

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

func main() {

	var c = Config{
		InCluster:                   true,
		DefalutWatchIntervalSeconds: 120,
	}

	in, err := ioutil.ReadFile("./example.yaml")
	if err != nil {
		panic(err.Error())
	}
	fmt.Print(c)
	err = yaml.Unmarshal(in, &c)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(c)

	ctx := context.Background()
	cancelableCtx, cancel := context.WithCancel(ctx)
	for _, w := range c.Watch {
		if w.WatchIntervalSeconds == 0 {
			w.WatchIntervalSeconds = c.DefalutWatchIntervalSeconds
		}
		go watch(cancelableCtx, w)
	}
	fmt.Println("goroutine call done")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	s := <-sig
	cancel()
	fmt.Printf("Signal received: %s , cancell goroutine.\n", s.String())
	//time.Sleep(30 * time.Second)

	//cancel()
	//fmt.Println("goroutine cancell done")
	//time.Sleep(5 * time.Second)
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//panic(err.Error())
	//}
	// creates the clientset
	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//panic(err.Error())
	//}
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
