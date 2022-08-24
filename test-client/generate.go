package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/simapp"

	typesTx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	// "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func broadcastTx() error {
	// TODO: can be used to generate new private/public key pair
	//priv1, _, _ := testdata.KeyTestPubAddr()

	// Choose your codec: Amino or Protobuf. Here, we use Protobuf, given by the
	// following function.
	encCfg := simapp.MakeTestEncodingConfig()

	// Create a new TxBuilder.
	txBuilder := encCfg.TxConfig.NewTxBuilder()

	// GENERATING TRANSACTION------------------------------------------------------------------------------------

	// TODO: send some coins to his address through CLI so that it can send txs
	seed := "circle win grain cook zoo aware photo sound grain monkey nothing remain ribbon admit push black name behind pyramid warrior unknown rug public smile"
	priv1 := secp256k1.GenPrivKeyFromSecret([]byte(seed)) // gives the following address: wasm1rr6wy72cc8u8ges9rjvjx4vrnena5mjxhqp0p4

	// TODO: use one of this or add new message types if needed
	//msg := CreateMsgSubmitEvidence(priv1)
	// msg := CreateMsgUnjail(priv1)
	msg := CreateMsgDelegate(priv1)

	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return err
	}

	txBuilder.SetGasLimit(2000000)
	// txBuilder.SetFeeAmount(...)
	// txBuilder.SetMemo(...)
	// txBuilder.SetTimeoutHeight(...)

	// SIGNING TRANSACTION------------------------------------------------------------------------------------

	privs := []cryptotypes.PrivKey{priv1}
	accNums := []uint64{0} // The accounts' account numbers
	accSeqs := []uint64{0} // The accounts' sequence numbers

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  encCfg.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			ChainID:       ChainId,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		sigV2, err := tx.SignWithPrivKey(
			encCfg.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, encCfg.TxConfig, accSeqs[i])
		if err != nil {
			return err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return err
	}

	// EXPORTING TRANSACTION------------------------------------------------------------------------------------

	// Generated Protobuf-encoded bytes.
	txBytes, err := encCfg.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	// Generate a JSON string.
	// txJSONBytes, err := encCfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	// if err != nil {
	// 	return err
	// }
	// txJSON := string(txJSONBytes)

	// BROADCASTING TRANSACTION------------------------------------------------------------------------------------

	// Create a connection to the gRPC server.
	grpcConn, _ := grpc.Dial(
		"localhost:9081",    // Or your gRPC server address.
		grpc.WithInsecure(), // The Cosmos SDK doesn't support any transport security mechanism.
	)
	defer grpcConn.Close()

	// Broadcast the tx via gRPC. We create a new client for the Protobuf Tx
	// service.
	txClient := typesTx.NewServiceClient(grpcConn)
	// We then call the BroadcastTx method on this client.
	grpcRes, err := txClient.BroadcastTx(
		context.Background(),
		&typesTx.BroadcastTxRequest{
			Mode:    typesTx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes, // Proto-binary of the signed transaction, see previous step.
		},
	)
	if err != nil {
		return err
	}

	fmt.Println(grpcRes.TxResponse.Code) // Should be `0` if the tx is successful
	fmt.Println(grpcRes.TxResponse.RawLog)

	return nil
}

func CreateMsgSubmitEvidence(priv1 *secp256k1.PrivKey) sdk.Msg {
	eq := &evidencetypes.Equivocation{
		Height:           5,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: priv1.PubKey().Address().String(),
	}

	addr := sdk.AccAddress(priv1.PubKey().Address())
	msg, err := evidencetypes.NewMsgSubmitEvidence(addr, eq)
	if err != nil {
		panic("failed to create MsgSubmitEvidence")
	}

	return msg
}

func CreateMsgDelegate(priv1 *secp256k1.PrivKey) sdk.Msg {
	accAddr := sdk.AccAddress(priv1.PubKey().Address())
	valAddr := sdk.ValAddress(priv1.PubKey().Address())
	msg := stakingtypes.NewMsgDelegate(accAddr, valAddr, sdk.NewCoin("stake", sdk.NewInt(150)))

	return msg
}

func CreateMsgUnjail(privKey *secp256k1.PrivKey) sdk.Msg {
	valAddr := sdk.ValAddress(privKey.PubKey().Address())
	msg := slashingtypes.NewMsgUnjail(valAddr)

	return msg
}
