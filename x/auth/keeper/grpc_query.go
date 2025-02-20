package keeper

import (
	"context"
	"errors"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var _ types.QueryServer = AccountKeeper{}

func (ak AccountKeeper) Accounts(c context.Context, req *types.QueryAccountsRequest) (*types.QueryAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(ak.key)
	accountsStore := prefix.NewStore(store, types.AddressStoreKeyPrefix)

	var accounts []*codectypes.Any
	pageRes, err := query.Paginate(accountsStore, req.Pagination, func(key, value []byte) error {
		account := ak.decodeAccount(value)
		any, err := codectypes.NewAnyWithValue(account)
		if err != nil {
			return err
		}

		accounts = append(accounts, any)
		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "paginate: %v", err)
	}

	return &types.QueryAccountsResponse{Accounts: accounts, Pagination: pageRes}, err
}

// Account returns account details based on address
func (ak AccountKeeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "Address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	addr, err := sdk.AccAddressFromBech32(req.Address)

	if err != nil {
		return nil, err
	}
	account := ak.GetAccount(ctx, addr)
	if account == nil {
		fmt.Println("here 1")
		fmt.Println("%+v\n", req)
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.Address)
	}

	any, err := codectypes.NewAnyWithValue(account)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryAccountResponse{Account: any}, nil
}

// Params returns parameters of auth module
func (ak AccountKeeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	params := ak.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ModuleAccounts returns all the existing Module Accounts
func (ak AccountKeeper) ModuleAccounts(c context.Context, req *types.QueryModuleAccountsRequest) (*types.QueryModuleAccountsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	modAccounts := make([]*codectypes.Any, 0, len(ak.permAddrs))

	for moduleName := range ak.permAddrs {
		account := ak.GetModuleAccount(ctx, moduleName)
		if account == nil {
			return nil, status.Errorf(codes.NotFound, "account %s not found", moduleName)
		}
		any, err := codectypes.NewAnyWithValue(account)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		modAccounts = append(modAccounts, any)
	}

	return &types.QueryModuleAccountsResponse{Accounts: modAccounts}, nil
}

func (ak AccountKeeper) Bech32Prefix(ctx context.Context, req *types.Bech32PrefixRequest) (*types.Bech32PrefixResponse, error) {
	bech32Prefix, err := ak.getBech32Prefix()
	if err != nil {
		return nil, err
	}

	return &types.Bech32PrefixResponse{Bech32Prefix: bech32Prefix}, nil
}

func (ak AccountKeeper) AddressBytesToString(ctx context.Context, req *types.AddressBytesToStringRequest) (*types.AddressBytesToStringResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.AddressBytes) == 0 {
		return nil, errors.New("empty address bytes is not allowed")
	}

	text, err := ak.addressCdc.BytesToString(req.AddressBytes)
	if err != nil {
		return nil, err
	}

	return &types.AddressBytesToStringResponse{AddressString: text}, nil
}

func (ak AccountKeeper) AddressStringToBytes(ctx context.Context, req *types.AddressStringToBytesRequest) (*types.AddressStringToBytesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(strings.TrimSpace(req.AddressString)) == 0 {
		return nil, errors.New("empty address string is not allowed")
	}

	bz, err := ak.addressCdc.StringToBytes(req.AddressString)
	if err != nil {
		return nil, err
	}

	return &types.AddressStringToBytesResponse{AddressBytes: bz}, nil
}
