package wallet

import (
	"context"
	"encoding/json"

	"github.com/intellisoftalpin/cardano-wallet-backend/config"
	"github.com/intellisoftalpin/cardano-wallet-backend/repo"
	walletPB "github.com/intellisoftalpin/proto/proto-gen/wallet"
)

type Server struct {
	walletPB.WalletServer

	TransactionRepo *repo.TransactionRepo
}

func NewServer(config *config.Config) *Server {
	transactionRepo, err := repo.NewTransactionRepo(config)
	if err != nil {
		panic(err)
	}

	return &Server{
		TransactionRepo: transactionRepo,
	}
}

func (s *Server) DecodeTransaction(ctx context.Context, in *walletPB.DecodeTransactionRequest) (*walletPB.DecodeTransactionResponse, error) {
	tx, err := s.TransactionRepo.DecodeTransaction(in.Tx, in.PolicyId, in.AssetId)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return &walletPB.DecodeTransactionResponse{
		DecodedTx: string(b),
		Status:    "OK",
	}, nil
}

func (s *Server) SubmitTransaction(ctx context.Context, in *walletPB.SubmitTransactionRequest) (*walletPB.SubmitTransactionResponse, error) {
	txHash, err := s.TransactionRepo.SubmitExternalTransaction(in.Tx)
	if err != nil {
		return nil, err
	}

	return &walletPB.SubmitTransactionResponse{
		TxHash: txHash,
		Status: "OK",
	}, nil
}

func (s *Server) GetTransaction(ctx context.Context, in *walletPB.GetTransactionRequest) (*walletPB.GetTransactionResponse, error) {
	rawTx, err := s.TransactionRepo.GetTransaction(in.TxHash, in.PolicyId, in.AssetId)
	if err != nil {
		return nil, err
	}

	return &walletPB.GetTransactionResponse{
		// DecodedTx: decodedTx,
		RawTx:  rawTx,
		Status: "OK",
	}, nil
}

func (s *Server) CreateTransaction(ctx context.Context, in *walletPB.CreateTransactionRequest) (*walletPB.CreateTransactionResponse, error) {
	rawTx, txHash, addressTo, transferAmount, assetAmount, assetDecimals, err := s.TransactionRepo.CreateTransaction(in.Tx, in.PolicyId, in.AssetId)
	if err != nil {
		return nil, err
	}

	return &walletPB.CreateTransactionResponse{
		DecodedTx:      string(rawTx),
		AddressTo:      addressTo,
		TransferAmount: transferAmount,
		AssetAmount:    assetAmount,
		AssetDecimals:  assetDecimals,
		TxHash:         txHash,
		Status:         "OK",
	}, nil
}

func (s *Server) CheckTokenBalance(ctx context.Context, in *walletPB.CheckTokenBalanceRequest) (*walletPB.Empty, error) {
	if err := s.TransactionRepo.CheckTokenBalance(in.Tx, in.PolicyId, in.AssetId); err != nil {
		return nil, err
	}

	return &walletPB.Empty{}, nil
}

// ----------------------------------------------------------------------

func (s *Server) GetAllTokens(ctx context.Context, in *walletPB.Empty) (*walletPB.GetAllTokensResponse, error) {
	tokens, err := s.TransactionRepo.GetAllTokens()
	if err != nil {
		return nil, err
	}

	var tokensPB []*walletPB.Token
	for _, token := range tokens {

		tokensPB = append(tokensPB, &walletPB.Token{
			AssetName: token.Metadata.Name,
			PolicyId:  token.PolicyID,
			AssetId:   token.AssetName,
			Ticker:    token.Metadata.Ticker,
			Logo:      token.Metadata.Logo,
			Decimals:  token.Metadata.Decimals,
			Address:   token.Address,
			Price:     &walletPB.Price{Price: token.Price},

			AssetUnit:     token.AssetUnit,
			AssetQuantity: token.AssetQuantity,
			AssetDecimals: token.AssetDecimals,
			Fee:           token.Fee,
			Deposit:       token.Deposit,
			ProcessingFee: token.ProcessingFee,
			TotalQuantity: token.TotalQuantity,
			RewardAddress: token.RewardAddress,
		})
	}

	return &walletPB.GetAllTokensResponse{
		Tokens: tokensPB,
	}, nil
}

func (s *Server) GetToken(ctx context.Context, in *walletPB.TokenID) (*walletPB.GetTokenResponse, error) {
	token, err := s.TransactionRepo.GetTokenData(in.TokenId)
	if err != nil {
		return nil, err
	}

	return &walletPB.GetTokenResponse{
		Token: &walletPB.Token{
			AssetName: token.Metadata.Name,
			PolicyId:  token.PolicyID,
			AssetId:   token.AssetName,
			Ticker:    token.Metadata.Ticker,
			Logo:      token.Metadata.Logo,
			Decimals:  token.Metadata.Decimals,
			Address:   token.Address,
			Price:     &walletPB.Price{Price: token.Price},
		},
	}, nil
}

func (s *Server) GetTokenPrice(ctx context.Context, in *walletPB.TokenID) (*walletPB.GetTokenPriceResponse, error) {
	price, err := s.TransactionRepo.GetTokenPrice(in.TokenId)
	if err != nil {
		return nil, err
	}

	return &walletPB.GetTokenPriceResponse{
		Price: &walletPB.Price{
			Price: price,
		},
	}, nil
}

func (s *Server) GetWalletNetworkInfo(ctx context.Context, in *walletPB.Empty) (*walletPB.GetWalletNetworkInfoResponse, error) {
	networkInfo, err := s.TransactionRepo.GetWalletNetworkInfo()
	if err != nil {
		return nil, err
	}

	return &walletPB.GetWalletNetworkInfoResponse{
		NetworkInfo: &walletPB.NetworkInfo{
			NetworkId:     networkInfo.NetworkInfo.NetworkID,
			ProtocolMagic: networkInfo.NetworkInfo.ProtocolMagic,
		},
		NetworkTip: &walletPB.NetworkTip{
			AbsoluteSlotNumber: networkInfo.NetworkTip.AbsoluteSlotNumber,
			EpochNumber:        networkInfo.NetworkTip.EpochNumber,
			SlotNumber:         networkInfo.NetworkTip.SlotNumber,
			Time:               networkInfo.NetworkTip.Time,
		},
		NextEpoch: &walletPB.NextEpoch{
			EpochNumber:    networkInfo.NextEpoch.EpochNumber,
			EpochStartTime: networkInfo.NextEpoch.EpochStartTime,
		},
		NodeEra: networkInfo.NodeEra,
		NodeTip: &walletPB.NodeTip{
			AbsoluteSlotNumber: networkInfo.NodeTip.AbsoluteSlotNumber,
			EpochNumber:        networkInfo.NodeTip.EpochNumber,
			Height: &walletPB.Quantity{
				Quantity: networkInfo.NodeTip.Height.Quantity,
				Unit:     networkInfo.NodeTip.Height.Unit,
			},
			SlotNumber: networkInfo.NodeTip.SlotNumber,
			Time:       networkInfo.NodeTip.Time,
		},
		SyncProgress: &walletPB.SyncProgress{
			Status: networkInfo.SyncProgress.Status,
			Progress: &walletPB.Progress{
				Quantity: networkInfo.SyncProgress.Progress.Quantity,
				Unit:     networkInfo.SyncProgress.Progress.Unit,
			},
		},
		WalletMode: networkInfo.WalletMode,
	}, nil
}

func (s *Server) GetWalletsState(ctx context.Context, in *walletPB.Empty) (*walletPB.GetWalletsStateResponse, error) {
	walletsState, err := s.TransactionRepo.GetWalletsState()
	if err != nil {
		return nil, err
	}

	var walletsStatePB []*walletPB.WalletState
	for _, walletState := range walletsState {
		walletsStatePB = append(walletsStatePB, &walletPB.WalletState{
			Status: walletState.Status,
			Progress: &walletPB.Progress{
				Quantity: walletState.Progress.Quantity,
				Unit:     walletState.Progress.Unit,
			},
		})
	}

	return &walletPB.GetWalletsStateResponse{
		WalletsState: walletsStatePB,
	}, nil
}
