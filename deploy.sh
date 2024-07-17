#!/bin/bash

docker build -t trace-test:tracer-test:v1.0.0 .

# Log in to the Azure cli.
az login

docker tag trace-test:tracer-test:v1.0.0 kopsdev26bf3467a7874ff19f0965e516c2918a.azurecr.io/redis-app/trace-test:tracer-test:v1.0.0

docker push kopsdev26bf3467a7874ff19f0965e516c2918a.azurecr.io/redis-app/trace-test:tracer-test:v1.0.0

# Update image on deployment.
az aks get-credentials --name sporetm-aks --resource-group kops

kubectl set image deployment/trace-test trace-test=kopsdev26bf3467a7874ff19f0965e516c2918a.azurecr.io/redis-app/trace-test:tracer-test:v1.0.0 --namespace redis-app

