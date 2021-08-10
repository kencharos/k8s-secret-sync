k8s sample
--------

## notice

I checked this process on M1-mac, docker k8s(docker-desktop)

## process

1. run vault server and vault-auto-injector
2. vault k8s auth setup
3. run k8s-secret-sync pod

## run vault server and vault-auto-injector

use helm to insall vault-sever and vault-auto-injector

```
helm repo add hashicorp https://helm.releases.hashicorp.com

# helm show values hashicorp/vault
helm install vault hashicorp/vault --set="server.dev.enabled=true" --set="ui.enabled=true"

```

access vault api & ui 

```
kubectl port-forward svc/vault 8200:8200 
```

## vault k8s auth setup

see also -> https://www.vaultproject.io/docs/auth/kubernetes and https://learn.hashicorp.com/tutorials/vault/agent-kubernetes?in=vault/kubernetes
docker-desktop cluster host is https://kubernetes.docker.internal:6443/

use setup from UI or vault command. 

```
# run `kubectl port-forward svc/vault 8200:8200 ` in other trminal

## vault base config
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root

## setup k8s auth method setting
kubectl create serviceaccount vault-auth
kubectl apply --filename vault-auth-service-account.yaml

export VAULT_SA_NAME=$(kubectl get sa vault-auth \
    --output jsonpath="{.secrets[*]['name']}")
export SA_JWT_TOKEN=$(kubectl get secret $VAULT_SA_NAME \
    --output 'go-template={{ .data.token }}' | base64 --decode)
export SA_CA_CRT=$(kubectl config view --raw --minify --flatten \
    --output 'jsonpath={.clusters[].cluster.certificate-authority-data}' | base64 --decode)
export K8S_HOST=$(kubectl config view --raw --minify --flatten \
    --output 'jsonpath={.clusters[].cluster.server}')

vault auth enable kubernetes
# issuer need in docker-desktop
vault write auth/kubernetes/config \
        issuer="https://kubernetes.default.svc.cluster.local" \
        token_reviewer_jwt="$SA_JWT_TOKEN" \
        kubernetes_host="$K8S_HOST" \
        kubernetes_ca_cert="$SA_CA_CRT"

## for k8s-sync config 

vault policy write secret-sync-policy - <<EOF
path "secret/data/*" {
    capabilities = ["read", "list"]
}
EOF

vault write auth/kubernetes/role/k8s-secret-sync \
        bound_service_account_names=k8s-secret-sync-sa \
        bound_service_account_namespaces=default \
        policies=secret-sync-policy \
        ttl=24h

## setup initial secret
vault kv put secret/opaque_sample \
      sample1='secret1' \
      secret2='secret2'

vault kv put secret/docker_cred_sample \
      docker_host='private.example.com:5050' \
      docker_username='user' \
      docker_password='pass' \


vault kv put secret/tls_sample @tlscert.json

```

## run k8s-secret-sync pod

```
kubectl apply -f k8s-secret-sync.yml
```