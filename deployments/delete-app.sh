#!/bin/bash
echo "########################## deleting share-my-files app "
kubectl delete all -l app=share-my-files -n applications
kubectl delete all,secrets,configmaps,pvc,pv -l app=share-my-files -n applications