package repo

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/intellisoftalpin/cardano-wallet-backend/config"
	cwalletapi "github.com/intellisoftalpin/cardano-wallet-backend/cwallet-api"
)

type TransactionRepo struct {
	// config *config.Config

	// wallets map[string]wallet

	wallets wallets

	CardanoWalletApi *cwalletapi.CardanoWalletApi
}

func NewTransactionRepo(config *config.Config) (t *TransactionRepo, err error) {
	t = &TransactionRepo{
		// config:  config,
		// wallets: make(map[string]wallet),
		wallets: wallets{
			mx:      &sync.RWMutex{},
			wallets: make(map[string]wallet),
		},
	}

	for _, w := range config.Wallets {
		t.wallets.SetWallet(w.ID, wallet{
			WalletConfig: w,
			state: cwalletapi.WalletState{
				Status: "syncing",
			},
		})
	}

	t.CardanoWalletApi, err = cwalletapi.NewCardanoWalletApi(config)
	if err != nil {
		return nil, err
	}

	go func() {
		timer := time.NewTicker(5 * time.Second)

		for range timer.C {
			// get wallet state from cardano-wallet

			wallets := t.wallets.GetWallets()

			for walletID := range wallets {
				wallet, err := t.CardanoWalletApi.GetWalletData(walletID)
				if err != nil {
					t.wallets.SetWalletState(walletID, cwalletapi.WalletState{
						Status: "syncing",
					})
					continue
				}

				t.wallets.SetWalletState(walletID, wallet.State)
			}
		}

	}()

	return t, nil
}

func (t *TransactionRepo) DecodeTransaction(txHash, policyID, assetID string) (tx cwalletapi.Transaction, err error) {
	wallet, _, err := t.wallets.GetWalletByPolicyID(policyID, assetID)
	if err != nil {
		return tx, err
	}

	tx, err = t.CardanoWalletApi.DecodeTransaction(wallet.ID, txHash)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func (t *TransactionRepo) SubmitExternalTransaction(tx string) (txHash string, err error) {
	txHash, err = t.CardanoWalletApi.SubmitExternalTransaction(tx)
	if err != nil {
		return txHash, err
	}

	return txHash, nil
}

func (t *TransactionRepo) GetTransaction(txHash, policyID, assetID string) (tx []byte, err error) {
	wallet, _, err := t.wallets.GetWalletByPolicyID(policyID, assetID)
	if err != nil {
		return tx, err
	}

	tx, err = t.CardanoWalletApi.GetTransaction(wallet.ID, txHash)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func (t *TransactionRepo) CreateTransaction(txCBOR, policyID, assetID string) (rawTx []byte, txHash, addressTo, transferAmount, assetAmount, assetDecimals string, err error) {
	wallet, asset, err := t.wallets.GetWalletByPolicyID(policyID, assetID)
	if err != nil {
		return rawTx, txHash, addressTo, transferAmount, assetAmount, assetDecimals, err
	}

	assetDecimals = fmt.Sprint(asset.AssetDecimals)

	tx, err := t.CardanoWalletApi.DecodeTransaction(wallet.ID, txCBOR)
	if err != nil {
		return rawTx, txHash, addressTo, transferAmount, assetAmount, assetDecimals, err
	}

	req, err := t.ConstructCreateTransactionRequest(tx, wallet.Passphrase, asset)
	if err != nil {
		return rawTx, txHash, addressTo, transferAmount, assetAmount, assetDecimals, err
	}

	addressTo = req.Payments[0].Address

	transferAmount = fmt.Sprintf("%d", req.Payments[0].Amount.Quantity)
	assetAmount = fmt.Sprintf("%d", req.Payments[0].Assets[0].Quantity)

	rawTx, newTx, err := t.CardanoWalletApi.CreateTransaction(wallet.ID, req)
	if err != nil {
		return rawTx, txHash, addressTo, transferAmount, assetAmount, assetDecimals, err
	}

	txHash = newTx.ID

	return rawTx, txHash, addressTo, transferAmount, assetAmount, assetDecimals, err
}

func (t *TransactionRepo) CheckTokenBalance(txCBOR, policyID, assetID string) error {
	wallet, walletAsset, err := t.wallets.GetWalletByPolicyID(policyID, assetID)
	if err != nil {
		return err
	}

	tx, err := t.CardanoWalletApi.DecodeTransaction(wallet.ID, txCBOR)
	if err != nil {
		return err
	}

	policyID = tx.Metadata["1002"].String
	assetID = tx.Metadata["1003"].String
	qty := tx.Metadata["1004"].Int

	if policyID == "" || assetID == "" || qty == 0 {
		return fmt.Errorf("invalid metadata")
	}

	walletData, err := t.CardanoWalletApi.GetWalletData(wallet.ID)
	if err != nil {
		return err
	}

	for _, asset := range walletData.Assets.Available {
		if asset.PolicyID == walletAsset.PolicyID &&
			asset.AssetName == walletAsset.AssetID &&
			asset.Quantity >= walletAsset.Buffer+walletAsset.AssetQuantityWithDecimals { // check if token balance is sufficient
			return nil
		}
	}

	return fmt.Errorf("insufficient balance")
}

func (t *TransactionRepo) GetAllTokens() (walletAssets []cwalletapi.WalletAsset, err error) {
	wallets := t.wallets.GetWallets()

	for _, w := range wallets {
		walletID := w.ID

		if w.state.Status != "ready" {
			continue
		}

		address, err := t.CardanoWalletApi.GetAddress(walletID)
		if err != nil {
			return nil, err
		}

		for _, a := range w.Assets {
			token, err := t.CardanoWalletApi.GetToken(walletID, a.PolicyID, a.AssetID)
			if err != nil {
				return nil, err
			}

			token.Address = address
			token.Price = a.PriceLovelace
			token.AssetUnit = a.AssetUnit
			token.AssetQuantity = a.AssetQuantityWithDecimals
			token.AssetDecimals = a.AssetDecimals
			token.Fee = a.Fee
			token.Deposit = a.Deposit
			token.ProcessingFee = a.ProcessingFee
			token.RewardAddress = a.RewardAddress

			walletData, err := t.CardanoWalletApi.GetWalletData(walletID)
			if err != nil {
				return nil, err
			}

			for _, asset := range walletData.Assets.Available {
				if asset.PolicyID == a.PolicyID &&
					asset.AssetName == a.AssetID &&
					asset.Quantity >= a.Buffer+a.AssetQuantityWithDecimals { // check if token balance is sufficient
					token.TotalQuantity = asset.Quantity - a.Buffer
				}
			}

			if token.TotalQuantity == 0 {
				continue
			}

			walletAssets = append(walletAssets, token)
		}
	}

	return walletAssets, err
}

func (t *TransactionRepo) GetTokenData(tokenID string) (token cwalletapi.WalletAsset, err error) {
	tID := strings.Split(tokenID, ".")
	if len(tID) != 2 {
		return token, fmt.Errorf("invalid tokenID")
	}

	policyID, assetID := tID[0], tID[1]

	wallet, asset, err := t.wallets.GetWalletByPolicyID(policyID, assetID)
	if err != nil {
		return token, err
	}

	walletID := wallet.ID

	walletData, err := t.CardanoWalletApi.GetWalletData(walletID)
	if err != nil {
		return token, err
	}

	address, err := t.CardanoWalletApi.GetAddress(walletID)
	if err != nil {
		return token, err
	}

	for _, a := range walletData.Assets.Available {
		if a.PolicyID == policyID && a.AssetName == assetID {
			token, err := t.CardanoWalletApi.GetToken(walletID, a.PolicyID, a.AssetName)
			if err != nil {
				return token, err
			}

			token.Address = address
			token.Price = asset.PriceLovelace

			return token, nil
		}
	}

	return token, err
}

func (t *TransactionRepo) GetTokenPrice(tokenID string) (price uint64, err error) {
	// parse tokenID. tokenID = "policyID.assetName"
	tID := strings.Split(tokenID, ".")
	if len(tID) != 2 {
		return price, fmt.Errorf("invalid tokenID")
	}

	policyID, assetID := tID[0], tID[1]

	_, asset, err := t.wallets.GetWalletByPolicyID(policyID, assetID)
	if err != nil {
		return price, err
	}

	return asset.PriceLovelace, err
}

// ----------------------------------------------------------------------

func (c *TransactionRepo) ConstructCreateTransactionRequest(tx cwalletapi.Transaction, passphrase string, asset config.Asset) (req cwalletapi.CreateTransactionRequest, err error) {
	address := tx.Metadata["1010"].String + tx.Metadata["1011"].String
	policyID := tx.Metadata["1002"].String
	assetID := tx.Metadata["1003"].String
	qty := tx.Metadata["1004"].Int

	if address == "" || policyID == "" || assetID == "" || qty == 0 {
		return req, fmt.Errorf("invalid metadata")
	}

	var totalOutputAmountQuantity uint64

	for _, output := range tx.Outputs {
		totalOutputAmountQuantity += output.Amount.Quantity
	}

	deposit := asset.Deposit

	if asset.PriceLovelace > totalOutputAmountQuantity-deposit-asset.ProcessingFee {
		return req, fmt.Errorf("insufficient balance")
	}

	finalQty := asset.AssetQuantityWithDecimals

	req = cwalletapi.CreateTransactionRequest{
		Passphrase: passphrase,
		Payments: []cwalletapi.Payment{
			{
				Address: address,
				Amount: cwalletapi.Quantity{
					Quantity: deposit,
					Unit:     "lovelace",
				},
				Assets: []cwalletapi.Asset{
					{
						PolicyID:  policyID,
						AssetName: assetID,
						Quantity:  finalQty,
					},
				},
			},
		},
		Withdrawal: "self",
		TimeToLive: cwalletapi.Quantity{
			Quantity: 3600, // 1 hour
			Unit:     "second",
		},
	}

	return req, nil
}

func (c *TransactionRepo) GetWalletNetworkInfo() (networkInfo cwalletapi.NetworkInfo, err error) {
	networkInfo, err = c.CardanoWalletApi.GetWalletNetworkInformation()
	if err != nil {
		return networkInfo, err
	}

	return networkInfo, err
}
