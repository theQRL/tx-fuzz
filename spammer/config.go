package spammer

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"

	txfuzz "github.com/rgeraldes24/tx-fuzz"
	"github.com/rgeraldes24/tx-fuzz/flags"
	"github.com/rgeraldes24/tx-fuzz/mutator"
	"github.com/theQRL/go-qrllib/dilithium"
	"github.com/theQRL/go-zond/rpc"
	"github.com/theQRL/go-zond/zondclient"
	"github.com/urfave/cli/v2"
)

type Config struct {
	backend *rpc.Client // connection to the rpc provider

	N          uint64                 // number of transactions send per account
	faucetAcc  *dilithium.Dilithium   // dilithium account of the faucet
	accs       []*dilithium.Dilithium // dilithium accounts
	corpus     [][]byte               // optional corpus to use elements from
	accessList bool                   // whether to create accesslist transactions
	gasLimit   uint64                 // gas limit per transaction

	seed int64            // seed used for generating randomness
	mut  *mutator.Mutator // Mutator based on the seed
}

func NewDefaultConfig(rpcAddr string, N uint64, accessList bool, rng *rand.Rand) (*Config, error) {
	// Setup RPC
	backend, err := rpc.Dial(rpcAddr)
	if err != nil {
		return nil, err
	}

	// Setup Accounts
	var accs []*dilithium.Dilithium
	for i := 0; i < len(staticSeeds); i++ {
		acc, err := dilithium.NewDilithiumFromHexSeed(staticSeeds[i][2:])
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}

	faucetAcc, err := dilithium.NewDilithiumFromHexSeed(txfuzz.SEED[2:])
	if err != nil {
		return nil, err
	}

	return &Config{
		backend:    backend,
		N:          N,
		faucetAcc:  faucetAcc,
		accs:       accs,
		corpus:     [][]byte{},
		accessList: accessList,
		gasLimit:   30_000_000,
		seed:       0,
		mut:        mutator.NewMutator(rng),
	}, nil
}

func NewConfigFromContext(c *cli.Context) (*Config, error) {
	// Setup RPC
	rpcAddr := c.String(flags.RpcFlag.Name)
	backend, err := rpc.Dial(rpcAddr)
	if err != nil {
		return nil, err
	}

	// Setup faucet
	faucetAcc, err := dilithium.NewDilithiumFromHexSeed(txfuzz.SEED[2:])
	if err != nil {
		return nil, err
	}
	if seed := c.String(flags.SeedFlag.Name); seed != "" {
		faucetAcc, err = dilithium.NewDilithiumFromHexSeed(seed[2:])
		if err != nil {
			return nil, err
		}
	}

	// Setup Accounts
	var accs []*dilithium.Dilithium
	nSeeds := c.Int(flags.CountFlag.Name)
	if nSeeds == 0 || nSeeds > len(staticSeeds) {
		fmt.Printf("Sanitizing count flag from %v to %v\n", nSeeds, len(staticSeeds))
		nSeeds = len(staticSeeds)
	}
	for i := 0; i < nSeeds; i++ {
		acc, err := dilithium.NewDilithiumFromHexSeed(staticSeeds[i][2:])
		if err != nil {
			return nil, err
		}
		accs = append(accs, acc)
	}

	// Setup gasLimit
	gasLimit := c.Int(flags.GasLimitFlag.Name)

	// Setup N
	N := c.Int(flags.TxCountFlag.Name)
	if N == 0 {
		N, err = setupN(backend, len(accs), gasLimit)
		if err != nil {
			return nil, err
		}
	}

	// Setup seed
	seed := c.Int64(flags.SeedFlag.Name)
	if seed == 0 {
		fmt.Println("No seed provided, creating one")
		rnd := make([]byte, 8)
		crand.Read(rnd)
		seed = int64(binary.BigEndian.Uint64(rnd))
	}

	// Setup Mutator
	mut := mutator.NewMutator(rand.New(rand.NewSource(seed)))

	// Setup corpus
	var corpus [][]byte
	if corpusFile := c.String(flags.CorpusFlag.Name); corpusFile != "" {
		corpus, err = readCorpusElements(corpusFile)
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		backend:    backend,
		N:          uint64(N),
		faucetAcc:  faucetAcc,
		accessList: !c.Bool(flags.NoALFlag.Name),
		gasLimit:   uint64(gasLimit),
		seed:       seed,
		accs:       accs,
		corpus:     corpus,
		mut:        mut,
	}, nil
}

func setupN(backend *rpc.Client, keys int, gasLimit int) (int, error) {
	client := zondclient.NewClient(backend)
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	txPerBlock := int(header.GasLimit / uint64(gasLimit))
	txPerAccount := txPerBlock / keys
	if txPerAccount == 0 {
		return 1, nil
	}
	return txPerAccount, nil
}

func readCorpusElements(path string) ([][]byte, error) {
	stats, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	corpus := make([][]byte, 0, len(stats))
	for _, file := range stats {
		b, err := os.ReadFile(fmt.Sprintf("%v/%v", path, file.Name()))
		if err != nil {
			return nil, err
		}
		corpus = append(corpus, b)
	}
	return corpus, nil
}
