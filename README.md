# BDJuno
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/forbole/bdjuno/Tests)](https://github.com/forbole/bdjuno/actions?query=workflow%3ATests)
[![Go Report Card](https://goreportcard.com/badge/github.com/forbole/bdjuno)](https://goreportcard.com/report/github.com/forbole/bdjuno)
![Codecov branch](https://img.shields.io/codecov/c/github/forbole/bdjuno/cosmos/v0.40.x)

BDJuno (shorthand for BigDipper Juno) is the [Juno](https://github.com/forbole/juno) implementation
for [BigDipper](https://github.com/forbole/big-dipper).

It extends the custom Juno behavior by adding different handlers and custom operations to make it easier for BigDipper
showing the data inside the UI.

All the chains' data that are queried from the RPC and gRPC endpoints are stored inside
a [PostgreSQL](https://www.postgresql.org/) database on top of which [GraphQL](https://graphql.org/) APIs can then be
created using [Hasura](https://hasura.io/).

## Usage
To know how to setup and run BDJuno, please refer to
the [docs website](https://docs.bigdipper.live/cosmos-based/parser/overview/).

## Local env via Docker
TODO: Add LINK to cudosbuilder init 

## Testing
If you want to test the code, you can do so by running

```shell
$ make test-unit
```

**Note**: Requires [Docker](https://docker.com).

This will:
1. Create a Docker container running a PostgreSQL database.
2. Run all the tests using that database as support.




### Configuration explanation:
1. There are 3 configs that have to be set up: 
    - Inside the BDJuno folder you have:
        - config.yaml - this is the config for the BDJuno - https://docs.bigdipper.live/cosmos-based/parser/config/config
          - ip address of node, db name and password for it are set here
        - genesis.json - this is the genesis file that is going to be parsed before BDJuno starts. It gets by the docker BDJuno docker file
    - .env-bdjuno - this is the env variables for the BDJuno docker (only the ones relevant are listed, leave others as is)
      - HASURA_GRAPHQL_DATABASE_URL - THE URL of the DB for hasura to read from
      - HASURA_GRAPHQL_ADMIN_SECRET - the password of the HASURA 
      - HASURA_GRAPHQL_ENDPOINT_URL - the DNS(IP):PORT of the machine where hasura is hosted
      - HASURA_ACTIONS_GRPC - the IP:9090 of the node that hasura is reading from
      - HASURA_ACTIONS_RPC - the IP:26657 of the node that hasura is reading from
      - HASURA_ACTIONS_PORT - the port of which hasura actions will run 
