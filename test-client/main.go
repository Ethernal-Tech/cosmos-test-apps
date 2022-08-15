package main

import sdk "github.com/cosmos/cosmos-sdk/types"

// TODO: set these values according to your chain prefix
const (
	Bech32AddrPrefix            = "wasm"
	Bech32PubKeyPrefix          = "wasmpub"
	Bech32ValidatorAddrPrefix   = "wasmvaloper"
	Bech32ValidatorPubKeyPrefix = "wasmvaloperpub"
	Bech32ConsensusAddrPrefix   = "wasmvalcons"
	Bech32ConsensusPubKeyPrefix = "wasmvalconspub"
	ChainId                     = "wasm"
)

func main() {
	setConfig()
	// written using https://docs.cosmos.network/v0.46/run-node/txs.html#using-grpc
	broadcastTx()
}

func setConfig() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(Bech32AddrPrefix, Bech32PubKeyPrefix)
	config.SetBech32PrefixForValidator(Bech32ValidatorAddrPrefix, Bech32ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(Bech32ConsensusAddrPrefix, Bech32ConsensusPubKeyPrefix)
	config.Seal()
}
