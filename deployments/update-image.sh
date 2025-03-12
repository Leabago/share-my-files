#!/bin/bash
echo "########################## build image"
minikube image build -t share-my-files -f ./Dockerfile .
echo "########################## apply deployment"
kubectl apply -f ./deployments/deployment.yaml 
echo "########################## restart deployment"
kubectl rollout restart deployment --selector=app=share-my-files -n=applications
