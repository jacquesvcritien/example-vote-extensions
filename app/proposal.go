package app

import (
	"encoding/json"
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProposalHandler struct {
	app App
}

func NewProposalHandler(app App) *ProposalHandler {
	return &ProposalHandler{
		app: app,
	}
}

func (h *ProposalHandler) SetHandlers(app *App) {
	app.SetPrepareProposal(h.PrepareProposal())
	app.SetProcessProposal(h.ProcessProposal())
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		// if len(req.Txs) == 0 {
		// 	return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
		// }

		// var injectedVoteExtTx StakeWeightedPrices
		// if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
		// 	h.logger.Error("failed to decode injected vote extension tx", "err", err)
		// 	return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
		// }

		// XXX: Call ValidateVoteExtensions once 0.48.x is released to verify vote
		// extension signatures and that 2/3 of the voting power is present.
		//
		// baseapp.ValidateVoteExtensions(...)

		// Verify the proposer's stake-weighted oracle prices by computing the same
		// calculation and comparing the results. We omit verification for brevity
		// and demo purposes.
		// stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, injectedVoteExtTx.ExtendedCommitInfo)
		// if err != nil {
		// 	return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
		// }
		// if err := compareOraclePrices(injectedVoteExtTx.StakeWeightedPrices, stakeWeightedPrices); err != nil {
		// 	return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}
		// }

		// // at this point we can persist the stake-weighted oracle prices to state
		// fCtx := h.app.GetFinalizeBlockStateCtx()
		// h.fauxOracleKeeper.SetOraclePrices(fCtx, stakeWeightedPrices)

		// // verify remainder of block proposal, i.e. req.Txs[1:]

		// return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf("Failing for error")
		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

		ci := req.LocalLastCommit

		for _, v := range ci.Votes {
			var voteExt VoteExtension

			if err := json.Unmarshal(v.VoteExtension, &voteExt); err != nil {
				ctx.Logger().Error("failed to decode vote extension", "err", err, "validator", fmt.Sprintf("%x", v.Validator.Address))
				return nil, err
			}

			ctx.Logger().Error("Vote extension height in verify")
			ctx.Logger().Error(strconv.FormatInt(voteExt.Height, 10))

			return nil, fmt.Errorf("Testing breaking consensus")

		}

		return &abci.ResponsePrepareProposal{}, nil
	}
}
