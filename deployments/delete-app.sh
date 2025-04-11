#!/bin/bash
echo "########################## deleting share-my-files app "
kubectl delete all --selector=app=share-my-files -n=applications