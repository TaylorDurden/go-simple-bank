# Changelog

### [Docker] Write docker-compose file and control service start-up orders with wait-for.sh

- Add docker-compose & Dockerfile to make all services up once
- PR: https://github.com/TaylorDurden/go-simple-bank/pull/14

### [AWS] Store & retrieve production secrets with AWS secrets manager

- Create AWS Secrets Manager to set env vars
- AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
- `openssl rand -hex 64 | head -c 32` to generate 32 bytes rand string for TOKEN_SYMMETRIC_KEY env var
- using AWS Secretsmanager to get secret values

```bash
aws secretsmanager get-secret-value --secret-id simple-bank --query SecretString --output text
{"DB_DRIVER":"postgres","DB_SOURCE":"postgresql://<username>:<password>@<dburl>:5432/simple_bank","SERVER_ADDRESS":"0.0.0.0:8080",
"TOKEN_SYMMETRIC_KEY":"e4a2f1add2a21faf0373fc19d4ac0b13%",
"ACCESS_TOKEN_DURATION":"15m","REFRESH_TOKEN_DURATION":"24h"}
```

- using `jq` to transform above text to env vars, then overwrite to `app.env` file:

  ```bash
  aws secretsmanager get-secret-value --secret-id simple-bank --query SecretString --output text | jq -r 'to_entries|map("\(.key)=\(.value)")|.[]' > app.env

  cat app.env

  DB_DRIVER=postgres
  DB_SOURCE=postgresql://<username>:<password>@<dburl>:5432/simple_bank
  SERVER_ADDRESS=0.0.0.0:8080
  TOKEN_SYMMETRIC_KEY=e4a2f1add2a21faf0373fc19d4ac0b13%
  ACCESS_TOKEN_DURATION=15m
  REFRESH_TOKEN_DURATION=24h
  ```

- add a step after `Login to Amazon ECR` in .github/workflows/deploy.yml

  ```yaml
  - name: Load secrets and save to app.env file
    run: aws secretsmanager get-secret-value --secret-id simple-bank --query SecretString --output text | jq -r 'to_entries|map("\(.key)=\(.value)")|.[]' > app.env
  ```

- make PR to build a ECR image
