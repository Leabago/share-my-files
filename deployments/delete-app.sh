#!/bin/bash
echo "########################## deleting share-my-files app "
kubectl delete all -l app=share-my-files -n applications

# do not delete secrets,configmaps,pvc,pv
# kubectl delete secrets,configmaps,pvc,pv -l app=share-my-files -n applications
