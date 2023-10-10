package cwalletapi

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bykovme/goconfig"
	"github.com/intellisoftalpin/cardano-wallet-backend/config"
	"github.com/intellisoftalpin/cardano-wallet-backend/helpers"
)

type CardanoWalletApi struct {
	url string

	wallets map[string]config.WalletConfig
}

func NewCardanoWalletApi(config *config.Config) (*CardanoWalletApi, error) {
	wallets, err := GetWalletsPasswords(config.CardanoWalletURL, config.Wallets)
	if err != nil {
		return nil, err
	}

	config.Wallets = wallets

	c := &CardanoWalletApi{
		url:     config.CardanoWalletURL,
		wallets: wallets,
	}

	return c, nil
}

// Send POST request to cwallet-api
// Return decoded tx
func (c *CardanoWalletApi) DecodeTransaction(walletID, txCBOR string) (tx Transaction, err error) {
	body, err := json.Marshal(decodeTxRequest{Transaction: txCBOR})
	if err != nil {
		return tx, err
	}

	resp, err := http.Post(c.url+"/v2/wallets/"+walletID+"/transactions-decode", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return tx, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return tx, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return tx, fmt.Errorf("tx not decoded: %s - %s", resp.Status, string(b))
	}

	if err = json.Unmarshal(b, &tx); err != nil {
		return tx, err
	}

	return tx, nil
}

// Submit External Transaction
func (c *CardanoWalletApi) SubmitExternalTransaction(txCBOR string) (string, error) {
	b, err := hex.DecodeString(txCBOR)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(c.url+"/v2/proxy/transactions", "application/octet-stream", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("tx not submitted: %s - %s", resp.Status, string(body))
	}

	var txID Transaction
	if err = json.Unmarshal(body, &txID); err != nil {
		return "", err
	}

	return txID.ID, nil
}

// Get transaction by id
func (c *CardanoWalletApi) GetTransaction(walletID, txID string) ([]byte, error) {
	resp, err := http.Get(c.url + "/v2/wallets/" + walletID + "/transactions/" + txID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tx not found: %s - %s", resp.Status, string(b))
	}

	return b, nil
}

// Create transaction
func (c *CardanoWalletApi) CreateTransaction(walletID string, req CreateTransactionRequest) (rawTx []byte, tx Transaction, err error) {
	body, err := json.Marshal(req)
	if err != nil {
		log.Println(err)
		return rawTx, tx, err
	}

	resp, err := http.Post(c.url+"/v2/wallets/"+walletID+"/transactions", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Println(err)
		return rawTx, tx, err
	}
	defer resp.Body.Close()

	rawTx, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return rawTx, tx, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return rawTx, tx, fmt.Errorf("tx not created: %s - %s", resp.Status, string(rawTx))
	}

	if err = json.Unmarshal(rawTx, &tx); err != nil {
		return rawTx, tx, err
	}

	return rawTx, tx, nil
}

// Get wallet by walletID
func (c *CardanoWalletApi) GetWalletData(walletID string) (wallet WalletResponse, err error) {
	resp, err := http.Get(c.url + "/v2/wallets/" + walletID)
	if err != nil {
		return wallet, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return wallet, err
	}

	if resp.StatusCode != http.StatusOK {
		return wallet, fmt.Errorf("wallet not found: %s - %s", resp.Status, string(b))
	}

	if err = json.Unmarshal(b, &wallet); err != nil {
		return wallet, err
	}

	return wallet, nil
}

func (c *CardanoWalletApi) GetAddress(walletID string) (address string, err error) {
	resp, err := http.Get(c.url + "/v2/wallets/" + walletID + "/addresses")
	if err != nil {
		return address, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return address, err
	}

	if resp.StatusCode != http.StatusOK {
		return address, fmt.Errorf("wallet not found: %s - %s", resp.Status, string(b))
	}

	var walletAddresses []WalletAddress
	if err = json.Unmarshal(b, &walletAddresses); err != nil {
		return address, err
	}

	return walletAddresses[0].ID, nil
}

// --------------------------------------------------------

func (c *CardanoWalletApi) GetToken(walletID, policyID, assetName string) (token WalletAsset, err error) {
	resp, err := http.Get(c.url + "/v2/wallets/" + walletID + "/assets/" + policyID + "/" + assetName)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return token, err
	}

	if resp.StatusCode != http.StatusOK {
		return token, fmt.Errorf("wallet not found: %s - %s", resp.Status, string(b))
	}

	if err = json.Unmarshal(b, &token); err != nil {
		return token, err
	}

	if token.Metadata.Name == "" {
		// hex to string
		b, err := hex.DecodeString(token.AssetName)
		if err != nil {
			return token, err
		}

		token.Metadata.Name = string(b)
	}

	return token, nil
}

// --------------------------------------------------------

// Create and restore a wallet from a mnemonic sentence or account public key.
func CreateWallet(url string, req CreateWalletRequest) (wallet WalletResponse, err error) {
	body, err := json.Marshal(req)
	if err != nil {
		return wallet, err
	}

	resp, err := http.Post(url+"/v2/wallets", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return wallet, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return wallet, err
	}

	if resp.StatusCode != http.StatusCreated {
		return wallet, fmt.Errorf("wallet not created: %s - %s", resp.Status, string(b))
	}

	if err = json.Unmarshal(b, &wallet); err != nil {
		return wallet, err
	}

	return wallet, nil
}

// --------------------------------------------------------

func GetWalletsPasswords(url string, wallets map[string]config.WalletConfig) (fullWallet map[string]config.WalletConfig, err error) {
	internalConf := config.InternalConfig{
		Wallets: make(map[string]config.InternalWalletConfig),
	}

	fullWallet = make(map[string]config.WalletConfig)

	// loads config from volume mounted to container
	if err = goconfig.LoadConfig("/data/config.json", &internalConf); err != nil {
		log.Println("No internal config found")
		log.Println("Error loading config: ", err)
	}

	for i, wallet := range wallets {
		iConf, ok := internalConf.Wallets[i]
		if ok {
			wallet.ID = iConf.WalletID
			wallet.Passphrase = iConf.WalletPassphrase
			fullWallet[i] = wallet
			continue
		}

		if wallet.Mnemonic != "" {
			mnemonic := strings.Split(wallet.Mnemonic, " ")
			passphrase := helpers.GeneratePassword(20, 1, 1, 1)

			// Create wallet
			w, err := CreateWallet(url, CreateWalletRequest{
				Name:           "wallet " + i,
				Mnemonic:       mnemonic,
				Passphrase:     passphrase,
				AddressPoolGap: 20,
			})
			if err != nil {
				log.Println("Error creating wallet: ", err)
				continue
			}

			wallet.ID = w.ID
			wallet.Passphrase = passphrase
			fullWallet[i] = wallet

			internalConf.Wallets[i] = config.InternalWalletConfig{
				WalletID:         w.ID,
				WalletPassphrase: passphrase,
			}
		}
	}

	if err = goconfig.SaveConfig("/data/config.json", internalConf); err != nil {
		log.Println("Error saving config: ", err)
	}

	return fullWallet, nil
}

// --------------------------------------------------------

func (c *CardanoWalletApi) GetWalletNetworkInformation() (info NetworkInfo, err error) {
	resp, err := http.Get(c.url + "/v2/network/information")
	if err != nil {
		return info, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, err
	}

	if resp.StatusCode != http.StatusOK {
		return info, fmt.Errorf("network info not found: %s - %s", resp.Status, string(b))
	}

	if err = json.Unmarshal(b, &info); err != nil {
		return info, err
	}

	return info, err
}
