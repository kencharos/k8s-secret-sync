K8S secret sync
===============

This tool watch files periodically, then sync file contents to k8s secret resources.
This is suited for working with hashicorp k8s-vault. 

## usage

build source `go build`
then, `run ./k8s-secret-sync config.yaml` 

or you can use dockerhub image `<WIP>`

you must write k8s cluster config and watching file/secret config in configuration yaml file(config yaml).

### config.yaml

sample is example_config.yaml .

#### basic config

+ defaultWatchIntervalSeconds: default 120. Watching period to check secret file.
+ inClusterMode: default "true". 
    + if true use k8s pod's service account to call k8s API. This service account has secret read,create,update role. see example k8s manifest.
    + if false you must specify k8s cluster config.
+ externalClusterKubeconfigPath: if inClusterMode is false and you want to use kubeconfig, set kubeconfig path.
+ externalClusterBearerToken, externalClusterHost: if inClusterMode is false and you want to connect any k8s cluster, set k8s cluster host and bearer token that can manage secret resources.
+ watch: list of secret sync setting.

#### secret sync setting

+ name, namespace: k8s secret resource name and its namespace.
+ type: secret type of secret. see -> https://kubernetes.io/ja/docs/concepts/configuration/secret/#secret-types
+ secretPath: yaml file path that describe secret.
+ watchIntervalSeconds: if omitted, use defaultWatchIntervalSeconds. specifc watch interval

#### secret file path.

secret file format is must be yaml. and value is must be string.
this tool register k8s secret resouce data from yaml key value pair.

if secret type is `kubernetes.io/dockerconfigjson`, this tool convert yaml value to `.dockerconfigjson`

example is follwing.



watch sync config is,

```
watch:
  - name: docker-cred
    namespace: sample-ns
    type: kubernetes.io/dockerconfigjson
    secretPath: /secret/docker-cred.yml
```

/secret/docker-cred.yml is here,

```
docker-server: private-repo.example.com:5050
docker-username: private-user
docker-password: private-pass
```

then, this tool write `.dockerconfigjson` from docker-server, docker-username and docker-password.

```
apiVersion: v1
kind: Secret
metadata:
  name: docker-cred
  namespace: sample-ns
type: kubernetes.io/dockercfg
data:
  .dockerconfigjson: e2F1dGh6....
```


### environment variables

+ LOG_LEVEL: default "info".  "panic", "error", "fatal", "warn", "debug" are OK.