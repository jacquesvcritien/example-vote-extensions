package examplevess_test

import (
	"testing"

	keepertest "examplevess/testutil/keeper"
	"examplevess/testutil/nullify"
	"examplevess/x/examplevess"
	"examplevess/x/examplevess/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ExampleVeSsKeeper(t)
	examplevess.InitGenesis(ctx, *k, genesisState)
	got := examplevess.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
