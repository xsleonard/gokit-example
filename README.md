# gokit-example

This is a demo service for accounts with balances and transfers between those accounts.

An account has an ID, balance and currency type. 
Accounts with the same currency type can transfer amounts between each other, but the account
balance can not go negative. 
All amounts have at most 2 decimals of precision (e.g. "1.23"), regardless of the currency.

<!-- MarkdownTOC levels="1,2" -->

- [Setup](#setup)
- [Running](#running)
- [API Docs](#api-docs)
- [Usage](#usage)
- [Development](#development)

<!-- /MarkdownTOC -->

## Setup

### Setup postgres

You can use docker-compose:

```sh
# start postgres docker container
docker-compose up -d
```

Otherwise, install and run postgres as you wish.
The default database name is `wallet`.

### Install golang-migrate

[migrate](https://github.com/golang-migrate/migrate) is used to manage DB migrations and setup the initial tables.

Follow the installation instructions from https://github.com/golang-migrate/migrate/tree/master/cmd/migrate to install.

### Setup database

If using docker-compose:

```sh
# create database
docker exec -it my_postgres psql -U postgres -c "create database wallet" 
```

Otherwise, create a database in your postgres setup as you wish. The default
database name is `wallet`.

Run the migration tool to initialize the [database schema](./migrations/1_init.up.sql):

```sh
make update-db
```

## Running

By default, the application runs on `localhost:8888` and connects to the local postgres database run by docker-compose, with url `postgresql://postgres@localhost:54320/wallet?sslmode=disable`.

```sh
go run ./cmd/wallet/
```

## API Docs

See the [API Docs](./API.md)

## Usage

### Add test data for manual experimentation

```sh
go run ./cmd/testdata
```

This adds 6 accounts and credits each account with 100 units of their currency.

### Server configuration

By default, the server will connect to the postgres database that is run by docker-compose and a database named `wallet`,
and listens on `localhost:8888`.

```sh
go run ./cmd/wallet -h
```

```
Usage of wallet:
  -addr string
        HTTP listen address (default "localhost:8888")
  -db string
        Postgres DB URL (default "postgresql://postgres@localhost:54320/wallet?sslmode=disable")
```

### Run the server

```sh
go run ./cmd/wallet
```

### List accounts

```sh
curl 'http://localhost:8888/v1/accounts'
```

### Make a transfer

Copy two IDs from the previous response, for accounts with the same currency type, and fill them in the response below.

```sh
curl -X POST 'http://localhost:8888/v1/transfer' -d '{"to":"...","from":"...","amount":"1.23"}'
```

### List payments

See all payments. It will include the original credits to the accounts created by the `testdata` tool, and the new transfer payments.

```
curl 'http://localhost:8888/v1/payments'
```

## Development

### Running tests

Create the test database (if using docker-compose):

```sh
docker exec -it wallet_postgres psql -U postgres -c "create database wallet_test"
```

Run the tests:

```sh
make test
```

### Linting

Linting is done by [golangci-lint](https://github.com/golangci/golangci-lint).

```sh
# Install linters if you don't have them
make install-linters
# Run the linters (uses golangci-lint)
make lint
```
