package app

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	examplevessmodulekeeper "examplevess/x/examplevess/keeper"
	"examplevess/x/examplevess/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	cometrpchttp "github.com/cometbft/cometbft/rpc/client/http"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	authtyps "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmjsonclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
)

type (
	// VoteExtensionHandler defines a dummy vote extension handler for SimApp.
	//
	// NOTE: This implementation is solely used for testing purposes. DO NOT use
	// in a production application!
	VoteExtensionHandler struct {
		exampleVeSsKeeper examplevessmodulekeeper.Keeper
	}

	// VoteExtension defines the structure used to create a dummy vote extension.
	VoteExtension struct {
		Hash   []byte
		Height int64
		Data   []byte
	}
)

func (h *VoteExtensionHandler) SetKeeper(exampleVeSsKeeper examplevessmodulekeeper.Keeper) {
	h.exampleVeSsKeeper = exampleVeSsKeeper
}

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{}
}

func (h *VoteExtensionHandler) SetHandlers(bApp *baseapp.BaseApp) {
	bApp.SetExtendVoteHandler(h.ExtendVote())
	bApp.SetVerifyVoteExtensionHandler(h.VerifyVoteExtension())
}

func submitToLayer1(ctx sdk.Context, keyName string, keyDir string, nodeRpc string, hash string, blockNumber int64) (string, error) {

	httpClient, _ := tmjsonclient.DefaultHTTPClient(nodeRpc)
	cometRpc, err := cometrpchttp.NewWithClient(nodeRpc, "/websocket", httpClient)
	if err != nil {
		ctx.Logger().Error("Error creating client: " + err.Error())
		return "", err
	}

	//create client context
	encodingConfig := MakeEncodingConfig()

	var buffer = bytes.NewBuffer(make([]byte, 0, 1024))
	if buffer == nil {
		ctx.Logger().Error("Buffer is nill")
	}

	clientContext := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithChainID("theta-testnet-001").
		WithInput(os.Stdin).
		WithKeyringDir(keyDir).
		WithClient(cometRpc).
		WithAccountRetriever(authtyps.AccountRetriever{}).
		WithHomeDir("/Users/jacques/.gaia").
		WithBroadcastMode("sync").
		// WithOutputFormat("json").
		WithOutput(buffer)

	// Get address
	keybase, _ := client.NewKeyringFromBackend(clientContext, "test")
	key, err := keybase.Key(keyName)

	if err != nil {
		ctx.Logger().Error("Error getting key from keyring: " + err.Error())
		return "", err
	}

	addr, err := key.GetAddress()

	if err != nil {
		fmt.Println("Error getting key from keyring: " + err.Error())
		return "", err
	}

	toAddr, err := sdk.AccAddressFromBech32("cosmos15cqpk72eh50cdjgmq58ehnhrlvcyk8z8dxz5v3")

	msg := banktypes.NewMsgSend(addr, toAddr, sdk.NewCoins(sdk.NewCoin("uatom", math.NewInt(100))))

	//create factory
	txFactory := clienttx.Factory{}.
		WithAccountRetriever(clientContext.AccountRetriever).
		WithChainID("theta-testnet-001").
		WithMemo(strconv.FormatInt(blockNumber, 10) + "-" + hash).
		WithTxConfig(clientContext.TxConfig).
		WithKeybase(keybase).
		WithGasAdjustment(1.3).
		WithSimulateAndExecute(true)

	ctx.Logger().Error(addr.String())

	err = txFactory.AccountRetriever().EnsureExists(clientContext, addr)

	if err != nil {
		fmt.Println("Error in getting address: " + err.Error())
	}

	initNum, initSeq := txFactory.AccountNumber(), txFactory.Sequence()

	if initNum == 0 || initSeq == 0 {
		num, seq, err := txFactory.AccountRetriever().GetAccountNumberSequence(clientContext, addr)
		if err != nil {
			fmt.Println("Error in getting account seq: " + err.Error())
			return "", err
		}

		if initNum == 0 {
			txFactory = txFactory.WithAccountNumber(num)
		}

		txFactory = txFactory.WithSequence(seq)
	}

	_, adjusted, err := clienttx.CalculateGas(clientContext, txFactory, msg)
	ctx.Logger().Error(strconv.FormatInt(int64(adjusted), 10))
	if err != nil {
		fmt.Println("Error in calculating gas: " + err.Error())
		return "", err
	}

	txFactory = txFactory.WithGas(adjusted).WithGasPrices("0.025uatom")

	clientContext.SkipConfirm = true
	clientContext.FromAddress = addr
	clientContext.GenerateOnly = false
	clientContext.FromName = keyName
	err = clienttx.GenerateOrBroadcastTxWithFactory(clientContext, txFactory, msg)
	if err != nil {
		fmt.Println("Error in broadcasting tx: " + err.Error())
		return "", err
	} else {
		fmt.Println("Broadcasted tx succesfully")
		var data map[string]json.RawMessage

		// Unmarshal the JSON data into the map
		err := json.Unmarshal([]byte(buffer.String()), &data)
		if err != nil {
			fmt.Println("Error parsing TX JSON:", err)
			return "", err
		}

		var txhash string
		err = json.Unmarshal(data["txhash"], &txhash)
		if err != nil {
			fmt.Println("Error parsing name field:", err)
			return "", err
		}
		return txhash, nil

	}
}

func (h *VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {

		ctx = ctx.WithLogger(log.NewLogger(os.Stdout))

		buf := make([]byte, 1024)

		_, err := rand.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random vote extension data: %w", err)
		}

		// HERE query ATOM TX and get the hash of the previous block
		ve := VoteExtension{
			Hash:   req.Hash,
			Height: req.Height,
			Data:   buf,
		}

		// Read config
		configFilePath := DefaultNodeHome + "/config/config.toml"

		// Set the configuration file name and path
		viper.SetConfigFile(configFilePath)

		// Read the configuration file
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Error reading config file:", err)
			return nil, fmt.Errorf("Error reading config file: %w", err)
		}

		rpc := viper.GetString("atom_rpc")
		keyname := viper.GetString("atom_keyname")
		keydir := viper.GetString("atom_keydir")
		ctx.Logger().Error(rpc)
		ctx.Logger().Error(keyname)
		ctx.Logger().Error(keydir)
		txhash, err := submitToLayer1(ctx, keyname, keydir, rpc, hex.EncodeToString(req.Hash), req.Height)
		if err != nil {
			return nil, fmt.Errorf("failed to send to layer1: %w", err)
		}

		hash := types.L1Hash{
			Block: uint64(req.Height),
			Hash:  txhash,
		}
		err = h.exampleVeSsKeeper.AppendHash(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("failed to append hash to keeper: %w", err)
		}
		ctx.Logger().Error("Vote extension height in extend")
		ctx.Logger().Error(strconv.FormatInt(ve.Height, 10))

		bz, err := json.Marshal(ve)
		if err != nil {
			return nil, fmt.Errorf("failed to encode vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		var ve VoteExtension
		ctx.Logger().Error("Vote extension height in verify here")
		ctx.Logger().Error(strconv.FormatInt(ve.Height, 10))

		// Here compare the block hash to the previous block hash
		// return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		// return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, fmt.Errorf("Incorrect hash")

		if err := json.Unmarshal(req.VoteExtension, &ve); err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		switch {
		case req.Height != ve.Height:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		case !bytes.Equal(req.Hash, ve.Hash):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		case len(ve.Data) != 1024:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
