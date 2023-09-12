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

## Testing
If you want to test the code, you can do so by running

```shell
$ make test-unit
```

**Note**: Requires [Docker](https://docker.com).

This will:
1. Create a Docker container running a PostgreSQL database.
2. Run all the tests using that database as support.

# Cudos Fork

The fork is based on branch cosmos/v0.47.x. It is at 67a1737418672d5357e72731b5e99a4c460c19df, which is 4 commit ahead of official version 4.0.0. These 4 commits (actually the last one) adds support to cosmos-sdk v0.47 proposals.

## Migrating from v2.0.2 (Cudos fork is tags as 1.6.x) to v4.0.0+ version

This version includes hasura actions as module so they do not need to be started separately. 

### config.yaml
- remove history module
- add actions, feegrant, group modules
- add database -> url to be the connection string from the .env plus appending "?sslmode=disable&search_path=public"
- add database -> partition_size: 100000
- add database -> partition_batch: 1000
- add actions -> port: 3286 (don't change the port because exactly this port is exposed in docker-compose.yaml)

we must upgrade "start_height" to be equal at the height when the chain is migrated to cosmos-sdk 0.47

### .env

- check for missing values in .env based on .env-prod.sample

### database

There are duplicates in "messages" table in existing databases. Check them and delete them beforing proceeding with the upgrade.

- Check for duplicates

```sql
SELECT message.transaction_hash, message.index, count(*) FROM message Group by message.transaction_hash, message.index HAVING count(*) > 1
```

or

```sql
SELECT * FROM message T1, message T2
WHERE  T1.ctid    < T2.ctid       -- delete the "older" ones
  AND  T1.transaction_hash    = T2.transaction_hash       -- list columns that define duplicates
  AND  T1.index = T2.index;
```

- Remove duplicates all at once
```sql
DELETE FROM message T1 USING message T2
WHERE  T1.ctid    < T2.ctid       -- delete the "older" ones
  AND  T1.transaction_hash    = T2.transaction_hash       -- list columns that define duplicates
  AND  T1.index = T2.index;
```

- Remove duplicates one by one
```sql
DELETE FROM message WHERE ctid IN ('(10469,7)', '(10264,6)')
```

## Explanation of some changes

### Daily fetcher

The BdJuno could skip a block. In order to fix that there is "original" daily_fetcher module but it only fetches blocks for last 24h. We have removed this module and replaced it by "fix_blocks_worker.go" which always check all blocks from <startHeight> to <last known height - 10>.
