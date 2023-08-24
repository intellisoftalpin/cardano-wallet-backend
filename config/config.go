package config

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"github.com/bykovme/goconfig"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string `json:"server_port"`

	CardanoWalletURL string `json:"cardano_wallet_url"`

	TLS TLSConfig `json:"tls"`

	Wallets map[string]WalletConfig `json:"wallets"`
}

type TLSConfig struct {
	CertPath string   `json:"cert_path"`
	IPs      []string `json:"ips"`
}

type WalletConfig struct {
	Mnemonic string  `json:"mnemonic_sentence"`
	Assets   []Asset `json:"assets"`

	ID         string `json:"-"`
	Passphrase string `json:"-"`
}

type Asset struct {
	PolicyID                  string  `json:"policy_id"`
	AssetID                   string  `json:"asset_id"`
	AssetName                 string  `json:"asset_name"`
	AssetUnit                 string  `json:"asset_unit"`
	PriceLovelace             uint64  `json:"lovelace_quantity"`
	AssetQuantity             float64 `json:"asset_quantity"`
	AssetQuantityWithDecimals uint64  `json:"-"`
	AssetDecimals             uint64  `json:"asset_decimals"`
	Fee                       uint64  `json:"fee"`
	Deposit                   uint64  `json:"deposit"`
	ProcessingFee             uint64  `json:"processing_fee"`

	Buffer        uint64 `json:"buffer"`
	RewardAddress string `json:"reward_address"`
}

type walletsConfig struct {
	Wallets map[string]WalletConfig `json:"wallets"`
}

type InternalConfig struct {
	Wallets map[string]InternalWalletConfig `json:"wallets"`
}

type InternalWalletConfig struct {
	WalletID         string `json:"wallet_id"`
	WalletPassphrase string `json:"wallet_passphrase"`
}

func LoadConfig() (loadedConfig *Config, err error) {
	// loads values from .env into the system
	if err := godotenv.Overload("../.env.local"); err != nil {
		log.Println("No .env file found")
		log.Println(err)
	}

	loadedConfig = &Config{
		ServerPort:       os.Getenv("SERVER_PORT"),
		CardanoWalletURL: os.Getenv("CARDANO_WALLET_URL"),
		TLS: TLSConfig{
			CertPath: os.Getenv("PATH_TO_CERTS"),
			IPs:      strings.Split(strings.ReplaceAll(os.Getenv("IP"), " ", ""), ";"),
		},
	}

	var wallets walletsConfig

	err = goconfig.LoadConfig(os.Getenv("CONFIG_PATH"), &wallets)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	for i := range wallets.Wallets {
		for j := range wallets.Wallets[i].Assets {
			quantity := wallets.Wallets[i].Assets[j].AssetQuantity
			decimals := float64(wallets.Wallets[i].Assets[j].AssetDecimals)
			wallets.Wallets[i].Assets[j].AssetQuantityWithDecimals = uint64(quantity * math.Pow(10, decimals))

			if wallets.Wallets[i].Assets[j].Buffer == 0 {
				wallets.Wallets[i].Assets[j].Buffer = uint64(quantity * math.Pow(10, decimals))
			}
		}
	}

	loadedConfig.Wallets = wallets.Wallets

	log.Println("Loaded config:", loadedConfig)

	return loadedConfig, nil
}
