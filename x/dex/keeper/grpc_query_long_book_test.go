package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/sei-protocol/sei-chain/testutil/keeper"
	"github.com/sei-protocol/sei-chain/testutil/nullify"
	"github.com/sei-protocol/sei-chain/x/dex/types"
)

func TestLongBookQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLongBook(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetLongBookRequest
		response *types.QueryGetLongBookResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetLongBookRequest{Price: msgs[0].Price.String(), ContractAddr: TEST_CONTRACT, PriceDenom: TEST_PRICE_DENOM, AssetDenom: TEST_ASSET_DENOM},
			response: &types.QueryGetLongBookResponse{LongBook: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetLongBookRequest{Price: msgs[1].Price.String(), ContractAddr: TEST_CONTRACT, PriceDenom: TEST_PRICE_DENOM, AssetDenom: TEST_ASSET_DENOM},
			response: &types.QueryGetLongBookResponse{LongBook: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetLongBookRequest{Price: "100000", ContractAddr: TEST_CONTRACT, PriceDenom: TEST_PRICE_DENOM, AssetDenom: TEST_ASSET_DENOM},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.LongBook(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestLongBookQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.DexKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNLongBook(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLongBookRequest {
		return &types.QueryAllLongBookRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
			ContractAddr: TEST_CONTRACT,
			PriceDenom:   TEST_PRICE_DENOM,
			AssetDenom:   TEST_ASSET_DENOM,
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LongBookAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LongBook), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LongBook),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LongBookAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.LongBook), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.LongBook),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LongBookAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.LongBook),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LongBookAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
