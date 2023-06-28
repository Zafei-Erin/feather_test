package bindings_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/Zafei-Erin/feather_test/app"
	bindings "github.com/Zafei-Erin/feather_test/x/tokenfactory/bindings/types"
	"github.com/Zafei-Erin/feather_test/x/tokenfactory/types"
)

func TestCreateDenomMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	lucky := RandomAccountAddress()
	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

	msg := bindings.TokenMsg{CreateDenom: &bindings.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	// query the denom and see if it matches
	query := bindings.TokenQuery{
		FullDenom: &bindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := bindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, fmt.Sprintf("factory/%s/SUN", reflect.String()))
}

func TestMintMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	lucky := RandomAccountAddress()
	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	// Create denom for minting
	msg := bindings.TokenMsg{CreateDenom: &bindings.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdk.NewIntFromString("808010808")
	require.True(t, ok)
	msg = bindings.TokenMsg{MintTokens: &bindings.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 1)
	coin := balances[0]
	require.Equal(t, amount, coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query := bindings.TokenQuery{
		FullDenom: &bindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := bindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// mint the same denom again
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 1)
	coin = balances[0]
	require.Equal(t, amount.MulRaw(2), coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = bindings.TokenQuery{
		FullDenom: &bindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = bindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// now mint another amount / denom
	// create it first
	msg = bindings.TokenMsg{CreateDenom: &bindings.CreateDenom{
		Subdenom: "MOON",
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	moonDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount = amount.SubRaw(1)
	msg = bindings.TokenMsg{MintTokens: &bindings.MintTokens{
		Denom:         moonDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 2)
	coin = balances[0]
	require.Equal(t, amount, coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = bindings.TokenQuery{
		FullDenom: &bindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "MOON",
		},
	}
	resp = bindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// and check the first denom is unchanged
	coin = balances[1]
	require.Equal(t, amount.AddRaw(1).MulRaw(2), coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = bindings.TokenQuery{
		FullDenom: &bindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = bindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)
}

func TestBurnMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	lucky := RandomAccountAddress()
	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	// Create denom for minting
	msg := bindings.TokenMsg{CreateDenom: &bindings.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdk.NewIntFromString("808010808")
	require.True(t, ok)

	msg = bindings.TokenMsg{MintTokens: &bindings.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	// can't burn from different address
	msg = bindings.TokenMsg{BurnTokens: &bindings.BurnTokens{
		Denom:           sunDenom,
		Amount:          amount,
		BurnFromAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.Error(t, err)

	// lucky needs to send balance to reflect contract to burn it
	luckyBalance := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	err = osmosis.BankKeeper.SendCoins(ctx, lucky, reflect, luckyBalance)
	require.NoError(t, err)

	msg = bindings.TokenMsg{BurnTokens: &bindings.BurnTokens{
		Denom:           sunDenom,
		Amount:          amount,
		BurnFromAddress: reflect.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
}

type ReflectExec struct {
	ReflectMsg    *ReflectMsgs    `json:"reflect_msg,omitempty"`
	ReflectSubMsg *ReflectSubMsgs `json:"reflect_sub_msg,omitempty"`
}

type ReflectMsgs struct {
	Msgs []wasmvmtypes.CosmosMsg `json:"msgs"`
}

type ReflectSubMsgs struct {
	Msgs []wasmvmtypes.SubMsg `json:"msgs"`
}

func executeCustom(t *testing.T, ctx sdk.Context, osmosis *app.App, contract sdk.AccAddress, sender sdk.AccAddress, msg bindings.TokenMsg, funds sdk.Coin) error {
	t.Helper()
	wrapped := bindings.TokenFactoryMsg{
		Token: &msg,
	}
	customBz, err := json.Marshal(wrapped)
	require.NoError(t, err)

	reflectMsg := ReflectExec{
		ReflectMsg: &ReflectMsgs{
			Msgs: []wasmvmtypes.CosmosMsg{{
				Custom: customBz,
			}},
		},
	}
	reflectBz, err := json.Marshal(reflectMsg)
	require.NoError(t, err)

	// no funds sent if amount is 0
	var coins sdk.Coins
	if !funds.Amount.IsNil() {
		coins = sdk.Coins{funds}
	}

	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	_, err = contractKeeper.Execute(ctx, contract, sender, reflectBz, coins)
	return err
}
