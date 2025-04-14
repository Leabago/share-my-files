## run share-my-files app
to run the share-my-files application use the script ./deployments/go-run.sh which sets the environment variables and run the go application

## manual deploy on kubernetes
to deploy the share-my-files application on kubernetes run the script ./deployments/update-image.sh which create image and apply deployment.yaml file

## deploy on kubernetes with helm
helm secrets install share-my-files -f share-my-files/values.yaml -f share-my-files/credentials.yaml share-my-files --namespace=applications
to deploy the share-my-files with helm