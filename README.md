# DataRobot batch-scoring tool rewritten with golang

## Setup

```bash
go install github.com/nimnull/go_scoring
```

```bash
git clone https://github.com/nimnull/go_scoring
cd go_scoring
go build
```

## Usage

```bash
./go_scoring standalone --help                                                                                                        master ✚ ✱ ◼
Use: go_scoring standalone [flags] <import_id> <dataset path>
```

Currently supports only `standalone` subcommand with following flags:

- `--host`: specifies host to request for predictions in format <proto>//:<ip|fqdn>[:port].


Arguments:

- `import_id`: cluster-defined unique identifier for imported model
- `dataset path`: csv file to score

