package repo

import (
	"fmt"
	"sync"

	"github.com/intellisoftalpin/cardano-wallet-backend/config"
	cwalletapi "github.com/intellisoftalpin/cardano-wallet-backend/cwallet-api"
)

type wallets struct {
	mx      *sync.RWMutex
	wallets map[string]wallet
}

type wallet struct {
	config.WalletConfig
	state cwalletapi.WalletState
}

func (w *wallets) GetWallets() (wallets map[string]wallet) {
	w.mx.RLock()
	defer w.mx.RUnlock()

	return w.wallets
}

func (w *wallets) SetWallets(wallets map[string]wallet) {
	w.mx.Lock()
	defer w.mx.Unlock()

	w.wallets = wallets
}

func (w *wallets) SetWallet(walletID string, wallet wallet) {
	w.mx.Lock()
	defer w.mx.Unlock()

	w.wallets[walletID] = wallet
}

func (w *wallets) GetWallet(walletID string) (wallet wallet, err error) {
	w.mx.RLock()
	defer w.mx.RUnlock()

	wallet, ok := w.wallets[walletID]
	if !ok {
		return wallet, fmt.Errorf("wallet not found")
	}

	return wallet, nil
}

func (w *wallets) GetWalletState(walletID string) (state cwalletapi.WalletState, err error) {
	w.mx.RLock()
	defer w.mx.RUnlock()

	wallet, ok := w.wallets[walletID]
	if !ok {
		return state, fmt.Errorf("wallet not found")
	}

	return wallet.state, nil
}

func (w *wallets) SetWalletState(walletID string, state cwalletapi.WalletState) {
	w.mx.Lock()
	defer w.mx.Unlock()

	w.wallets[walletID] = wallet{
		WalletConfig: w.wallets[walletID].WalletConfig,
		state:        state,
	}
}

func (w *wallets) GetWalletByPolicyID(policyID, assetID string) (wallet wallet, asset config.Asset, err error) {
	w.mx.RLock()
	defer w.mx.RUnlock()

	for _, w := range w.wallets {
		assets := w.Assets
		for _, asset := range assets {
			if asset.PolicyID == policyID && asset.AssetID == assetID {
				return w, asset, nil
			}
		}
	}

	return wallet, asset, fmt.Errorf("wallet not found")
}
