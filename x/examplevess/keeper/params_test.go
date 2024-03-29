package keeper_test

import (
	"testing"

	testkeeper "examplevess/testutil/keeper"
	"examplevess/x/examplevess/types"

	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ExampleVeSsKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
