package keeper

import (
	"encoding/binary"
	"examplevess/x/examplevess/types"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AppendHash(ctx sdk.Context, hash types.L1Hash) error {
	// Get the store
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.L1HashKey))

	// Convert the category ID into bytes
	byteKey := make([]byte, 64)
	binary.BigEndian.PutUint64(byteKey, hash.Block)

	// Marshal the category into bytes
	appendedValue := []byte(hash.Hash)

	// Insert the category bytes using category ID as a key
	store.Set(byteKey, appendedValue)

	return nil
}
