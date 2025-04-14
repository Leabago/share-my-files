#!/bin/bash
echo "########################## update secrets"
git pull
sops --decrypt deployments/tls/tls-secret-enc.yaml > deployments/tls/tls-secret.yaml
sops --decrypt deployments/tls/zerossl-enc.yaml > deployments/tls/zerossl.yaml

kubectl apply -f deployments/tls/tls-secret.yaml -n applications
kubectl apply -f deployments/tls/zerossl.yaml -n applications
rm deployments/tls/tls-secret.yaml deployments/tls/zerossl.yaml