# Go + JWT + Mongo

An example JWT auth server in Go with MongoDB.

## Config

The server is configured via environment variables and with `.env` file.
Predefined env variables has more priority than those from `.env` file (consider it as a sensible defaults).

Refer to [`template.env`](./template.env) for more.

```bash
# Use this template like this
cp template.env .env
```

## Run

`docker-compose.yml` contains sensible env variables (such as `SERVER_JWT_SECRET`) just for demonstration purposes. You can run it without configuring anything.

```bash
docker compose up
```

**It will spin up a...**

- [mongo](https://hub.docker.com/_/mongo) - port `27017`; root username `root`; root password `example`
- [mongo-express](https://hub.docker.com/_/mongo-express) (web ui for mongo) - port `8081`
- JWT (from [Dockerfile](./Dockerfile)) server - port `1234`

## Libraries used

- [koanf](https://github.com/knadh/koanf) - for configuration management
- [echo](https://github.com/labstack/echo) - web framework 
- [mongo-go-driver](https://github.com/mongodb/mongo-go-driver) - mongodb driver
- [jwt-go](https://github.com/golang-jwt/jwt)

And...

- [oapi-codegen](https://github.com/deepmap/oapi-codegen) - generates server boilerplate from [openapi schema](./openapi.yaml).


## Notes

The task requires hashing refresh tokens in DB with `bcrypt` but it can't operate
on passwords longer that 72 bytes which makes it unsuitable for JWT tokens, so
I used `sha512` for that purpose

https://stackoverflow.com/questions/64860460/store-the-hashed-jwt-token-in-the-database
