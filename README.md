# tx-fuzz

tx-fuzz is a package containing helpful functions to create random transactions. 
It can be used to easily access fuzzed transactions from within other programs.

## Usage

```
cd cmd/livefuzzer
go build
```

Run a Gzond execution layer client locally in a standalone bash window.
Tx-fuzz sends transactions to port `8545` by default.

```
gzond --http --http.port 8545
```

Run livefuzzer.

```
./livefuzzer spam
```

tx-fuzz allows for an optional seed parameter to get reproducible fuzz transactions

## Advanced usage
You can optionally specify a seed parameter or a secret key to use as a faucet

```
./livefuzzer spam --seed <seed> --sk <SK>
```

You can set the RPC to use with `--rpc <RPC>`.
