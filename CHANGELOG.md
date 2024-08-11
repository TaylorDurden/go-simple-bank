# Changelog

### docker run a local postgres database

```bash
$ docker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=simple_bank postgres:14-alpine
```

### [Docker] Write docker-compose file and control service start-up orders with wait-for.sh

- Add docker-compose & Dockerfile to make all services up once
- PR: https://github.com/TaylorDurden/go-simple-bank/pull/14

### [AWS] Store & retrieve production secrets with AWS secrets manager

- Create AWS Secrets Manager to set env vars
- AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
- `openssl rand -hex 64 | head -c 32` to generate 32 bytes rand string for TOKEN_SYMMETRIC_KEY env var
- using AWS Secretsmanager to get secret values

  ```bash
  $ aws secretsmanager get-secret-value --secret-id simple-bank --query SecretString --output text

  {"DB_DRIVER":"postgres","DB_SOURCE":"postgresql://<username>:<password>@<dburl>:5432/simple_bank","SERVER_ADDRESS":"0.0.0.0:8080",
  "TOKEN_SYMMETRIC_KEY":"e4a2f1add2a21faf0373fc19d4ac0b13%",
  "ACCESS_TOKEN_DURATION":"15m","REFRESH_TOKEN_DURATION":"24h"}
  ```

- using `jq` to transform above text to env vars, then overwrite to `app.env` file:

  ```bash
  $ aws secretsmanager get-secret-value --secret-id simple-bank --query SecretString --output text | jq -r 'to_entries|map("\(.key)=\(.value)")|.[]' > app.env

  $ cat app.env

  DB_DRIVER=postgres
  DB_SOURCE=postgresql://<username>:<password>@<dburl>:5432/simple_bank
  SERVER_ADDRESS=0.0.0.0:8080
  TOKEN_SYMMETRIC_KEY=e4a2f1add2a21faf0373fc19d4ac0b13%
  ACCESS_TOKEN_DURATION=15m
  REFRESH_TOKEN_DURATION=24h
  ```

- add `source /app/app.env` in start.sh to effect the env vars

- add a step after `Login to Amazon ECR` in .github/workflows/deploy.yml

  ```yaml
  - name: Load secrets and save to app.env file
    run: aws secretsmanager get-secret-value --secret-id simple-bank --query SecretString --output text | jq -r 'to_entries|map("\(.key)=\(.value)")|.[]' > app.env
  ```

- make PR to build a ECR image
- validate the ECR image locally with aws ecr get-login-password command

  ```bash
  $ aws ecr get-login-password | docker login --username AWS --password-stdin 975049981118.dkr.ecr.us-east-1.amazonaws.com

  Login Succeeded

  # pull ECR image URI
  $ docker pull 975049981118.dkr.ecr.us-east-1.amazonaws.com/simplebank:a3a1068d3d76e6f89de6330ec77c48d0c82dddfe

  # docker run pulled ECR image locally on port 8080 to start api server
  $ docker run -p 8080:8080 975049981118.dkr.ecr.us-east-1.amazonaws.com/simplebank:a3a1068d3d76e6f89de6330ec77c48d0c82dddfe
  ```

- test the apis with postman and check data in production database

### [AWS] Create EKS

- create a new EKS cluster with new IAM role and networking | add-ons settings
- add node group after the EKS cluster was created, and you can edit the node group to scaling
- set IAM user and find the user group to add inline permission policy to grant all EKS access
- use below command in local terminal to add new created context eks cluster

  ```bash
  $ aws eks update-kubeconfig --name simple-bank --region us-east-1

  Added new context arn:aws:eks:us-east-1:975049981118:cluster/simple-bank to /Users/taylor/.kube/config
  ```

- print the eks cluster info

  ```bash
  $ cat ~/.kube/config

  apiVersion: v1
  clusters:
  - cluster:
    certificate-authority: /Users/taylor/.minikube/ca.crt
    server: https://192.168.99.100:8443
    name: minikube
  - cluster:
    certificate-authority-data: \*\*\*
    server: https://ECF10A99902BF198B6F379C5AF459899.gr7.us-east-1.eks.amazonaws.com
    name: arn:aws:eks:us-east-1:975049981118:cluster/simple-bank
  contexts:
  - context:
    cluster: minikube
    user: minikube
    name: minikube
  - context:
    cluster: arn:aws:eks:us-east-1:975049981118:cluster/simple-bank
    user: arn:aws:eks:us-east-1:975049981118:cluster/simple-bank
    name: arn:aws:eks:us-east-1:975049981118:cluster/simple-bank
  current-context: arn:aws:eks:us-east-1:975049981118:cluster/simple-bank
  kind: Config
  preferences: {}
  users:
  - name: minikube
    user:
    client-certificate: /Users/taylor/.minikube/client.crt
    client-key: /Users/taylor/.minikube/client.key
  - name: arn:aws:eks:us-east-1:975049981118:cluster/simple-bank
    user:
    exec:
    apiVersion: client.authentication.k8s.io/v1beta1
    args: - --region - us-east-1 - eks - get-token - --cluster-name - simple-bank - --output - json
    command: aws

  ```

- switch context

  ```bash
  $ kubeclt config use-context arn:aws:eks:us-east-1:975049981118:cluster/simple-bank

  Switched to context "arn:aws:eks:us-east-1:975049981118:cluster/simple-bank".
  ```

- check current context user

  ```bash
  $ aws sts get-caller-identity

  # which is not the root user who created the EKS cluster, so this user does not have the access to EKS cluster
  # we need to grant this user
  {
    "UserId": "AIDA6GBMB7S7AORDWULNW",
    "Account": "975049981118",
    "Arn": "arn:aws:iam::975049981118:user/github-actions"
  }
  ```

- to grant the user, login to AWS and click on your account - Security credentials - create access key

  ```bash
  # set default profile to your new created aws credentials
  # `aws configure` use this to create aws credential and region info if you do not this file
  $ vim ~/.aws/credentials

  [default]
  aws_access_key_id = <your-key>
  aws-secret_access_key = <your-secret>

  [github]
  you can set your original key & secret here
  ```

- check cluster info if it works

  ```bash
  $ kubectl get pods

  No resources found in default namespace

  $ kubectl cluster-info

  Kubernetes control plane is running at https://ECF10A99902BF198B6F379C5AF459899.gr7.us-east-1.eks.amazonaws.com
  CoreDNS is running at https://ECF10A99902BF198B6F379C5AF459899.gr7.us-east-1.eks.amazonaws.com/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
  ```

- grant github-actions user by ConfigMap by RBAC and use github actions CI to deploy

  ```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: aws-auth
    namespace: kube-system
  data:
    mapUsers: |
      - userarn: arn:aws:iam::975049981118:user/github-actions
        username: github-actions
        groups:
          - system:masters
  ```

  ```bash
  # match the aws credential profile
  $ export AWS_PROFILE=default

  # apply the config by root user
  $ kubectl apply -f eks/aws-auth.  yaml

  # swithc to granted github user
  $ export AWS_PROFILE=github
  $ kubectl cluster-info

  Kubernetes control plane is running at https://ECF10A99902BF198B6F379C5AF459899.gr7.us-east-1.eks.amazonaws.com
  CoreDNS is running at https://ECF10A99902BF198B6F379C5AF459899.gr7.us-east-1.eks.amazonaws.com/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
  ```

- use [k9s](https://k9scli.io/) to interact with kubernetes nicely
- create a k8s deployment file

  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: simple-bank-api-deployment
    labels:
      app: simple-bank-api
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: simple-bank-api
    template:
      metadata:
        labels:
          app: simple-bank-api
      spec:
        containers:
          - name: simple-bank-api
            image: 975049981118.dkr.ecr.us-east-1.amazonaws.com/simplebank:1e828020197c773bd9b51b39a11f2640a8b02ac8
            ports:
              - containerPort: 8080
        restartPolicy: Always
  ```

- apply the deployment and check it by k9s

```bash
# -f to specify the file location
kubectl apply -f eks/deployment.yaml
```

- create a service to expose container port to service port

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: simple-bank-api-service
  spec:
    selector:
      app: simple-bank-api
    ports:
      - protocol: TCP
        port: 80
        targetPort: 8080
    type: ClusterIP
  ```

- k9s to inspect the LoadBalancer domain

  ```bash
  $ nslookup aa80c1853c2d1459fb25ca1b02073524-1903647260.us-east-1.elb.amazonaws.com

  Server:		192.168.5.1
  Address:	192.168.5.1#53

  Non-authoritative answer:
  Name:	aa80c1853c2d1459fb25ca1b02073524-1903647260.us-east-1.elb.amazonaws.com
  Address: 54.208.108.67
  Name:	aa80c1853c2d1459fb25ca1b02073524-1903647260.us-east-1.elb.amazonaws.com
  Address: 52.7.200.173
  ```

- change postman request url to domain to test
- change to `type: ClusterIP` in service.yaml add a ingress.yml

  ```yaml
  apiVersion: networking.k8s.io/v1
  kind: IngressClass
  metadata:
    name: nginx
  spec:
    controller: k8s.io/ingress-nginx
  ---
  apiVersion: networking.k8s.io/v1
  kind: Ingress
  metadata:
    name: simple-bank-ingress
    annotations:
      kubernetes.io/ingress.class: nginx
  spec:
    ingressClassName: nginx
    rules:
      - host: "api.reshopping.life"
        http:
          paths:
            - pathType: Prefix
              path: /
              backend:
                service:
                  name: simple-bank-api-service
                  port:
                    number: 80
  ```

  ```bash
  $ kubectl apply -f eks/service.yaml
  $ kubectl apply -f eks/ingress.yaml
  ```

- add a [nginx ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/#aws) to set an address to aws ingress

  ```bash
  $ kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.1/deploy/static/provider/aws/deploy.yaml

  #check the ingress controller in pods
  ```

### Configure route53 to transfer to your domain to ingress service

### Configure https TLS cert

- install [cert manager](https://cert-manager.io/docs/installation/kubectl/)
  ```bash
  $ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.2/cert-manager.yaml
  ```
- install [ACME issuer](https://cert-manager.io/docs/configuration/acme/)

  ```yaml
  # eks/issuer.yaml

  apiVersion: cert-manager.io/v1
  kind: ClusterIssuer
  metadata:
    name: letsencrypt
  spec:
    acme:
      # You must replace this email address with your own.
      # Let's Encrypt will use this to contact you about expiring
      # certificates, and issues related to your account.
      email: user@example.com
      server: https://acme-v02.api.letsencrypt.org/directory
      privateKeySecretRef:
        # Secret resource that will be used to store the account's private key.
        name: example-issuer-account-key
      # Add a single challenge solver, HTTP01 using nginx
      solvers:
        - http01:
            ingress:
              ingressClassName: nginx
  ```

- kubectl apply -f eks/issuer.yaml

### Configure github actions to auto deploy to EKS

- check out the file: `.github/workflows/deploy-eks.yml`
