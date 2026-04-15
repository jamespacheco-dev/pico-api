apiVersion: apps/v1
kind: Deployment
metadata:
  name: pico-api
  namespace: pico
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pico-api
  template:
    metadata:
      labels:
        app: pico-api
    spec:
      containers:
      - name: pico-api
        image: ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
        ports:
          - containerPort: 8080
