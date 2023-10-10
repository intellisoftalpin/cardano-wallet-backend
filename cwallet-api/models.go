package cwalletapi

type decodeTxRequest struct {
	Transaction string `json:"transaction"`
}

type RestoreWalletRequest struct {
	Name           string   `json:"name"`
	Mnemonic       []string `json:"mnemonic_sentence"`
	Passphrase     string   `json:"passphrase"`
	AddressPoolGap uint64   `json:"address_pool_gap"`
}

type WalletResponse struct {
	ID             string     `json:"id"`
	AddressPoolGap uint64     `json:"address_pool_gap"`
	Balance        Balance    `json:"balance"`
	Assets         Assets     `json:"assets"`
	Delegation     Delegation `json:"delegation"`
	Name           string     `json:"name"`
	Passphrase     struct {
		LastUpdatedAt string `json:"last_updated_at"`
	} `json:"passphrase"`
	State WalletState `json:"state"`
	Tip   Tip         `json:"tip"`
}

type WalletState struct {
	Status   string   `json:"status"`
	Progress Quantity `json:"progress"`
}

type Balance struct {
	Available Quantity `json:"available"`
	Reward    Quantity `json:"reward"`
	Total     Quantity `json:"total"`
}

type Quantity struct {
	Quantity uint64 `json:"quantity"`
	Unit     string `json:"unit"`
}

type Assets struct {
	Available []Asset `json:"available"`
	Total     []Asset `json:"total"`
}

type Asset struct {
	PolicyID  string `json:"policy_id"`
	AssetName string `json:"asset_name"`
	Quantity  uint64 `json:"quantity"`
}

type Delegation struct {
	Active struct {
		Status string `json:"status"`
		Target string `json:"target"`
	} `json:"active"`
	Next []struct {
		Status    string `json:"status"`
		ChangesAt struct {
			EpochNumber    uint64 `json:"epoch_number"`
			EpochStartTime string `json:"epoch_start_time"`
		} `json:"changes_at"`
	} `json:"next"`
}

type Tip struct {
	AbsoluteSlotNumber uint64   `json:"absolute_slot_number"`
	SlotNumber         uint64   `json:"slot_number"`
	EpochNumber        uint64   `json:"epoch_number"`
	Time               string   `json:"time"`
	Height             Quantity `json:"height"`
}

// -----------------------------------------

type Transaction struct {
	ID                string           `json:"id"`
	Amount            Quantity         `json:"amount"`
	Fee               Quantity         `json:"fee"`
	DepositTaken      Quantity         `json:"deposit_taken"`
	DepositReturned   Quantity         `json:"deposit_returned"`
	InsertedAt        Tip              `json:"inserted_at"`
	ExpiresAt         Tip              `json:"expires_at"`
	PendingSince      Tip              `json:"pending_since"`
	Depth             Quantity         `json:"depth"`
	Direction         string           `json:"direction"`
	Inputs            []Input          `json:"inputs"`
	Outputs           []Payment        `json:"outputs"`
	Collaterals       []Collateral     `json:"collateral"`
	CollateralOutputs []Payment        `json:"collateral_outputs"`
	Withdrawals       []Withdrawal     `json:"withdrawals"`
	Status            string           `json:"status"`
	Metadata          Metadata         `json:"metadata"`
	ScriptValidity    string           `json:"script_validity"`
	Certificates      []Certificate    `json:"certificates"`
	Mint              Mint             `json:"mint"`
	Burn              Burn             `json:"burn"`
	ValidityInterval  ValidityInterval `json:"validity_interval"`
	ScriptIntegrity   []string         `json:"script_integrity"`
	ExtraSignatures   []string         `json:"extra_signatures"`
}

type Input struct {
	Payment
	ID    string `json:"id"`
	Index uint64 `json:"index"`
}

type Collateral struct {
	Address string   `json:"address"`
	Amount  Quantity `json:"amount"`
	ID      string   `json:"id"`
	Index   uint64   `json:"index"`
}

type Withdrawal struct {
	StakeAddress string   `json:"stake_address"`
	Amount       Quantity `json:"amount"`
}

type Certificate struct {
	CertificateType   string   `json:"certificate_type"`
	Pool              string   `json:"pool"`
	RewardAccountPath []string `json:"reward_account_path"`
}

type Mint struct {
	Tokens               []Token `json:"tokens"`
	WalletPolicyKeyHash  string  `json:"wallet_policy_key_hash"`
	WalletPolicyKeyIndex string  `json:"wallet_policy_key_index"`
}

type Token struct {
	PolicyID     string       `json:"policy_id"`
	PloicyScript PloicyScript `json:"ploicy_script"`
	Assets       []TokenAsset `json:"assets"`
}

type PloicyScript struct {
	ScriptType string    `json:"script_type"`
	Script     string    `json:"script"`
	Reference  Reference `json:"reference"`
}

type Reference struct {
	ID    string `json:"id"`
	Index uint64 `json:"index"`
}

type TokenAsset struct {
	AssetName   string `json:"asset_name"`
	Quantity    uint64 `json:"quantity"`
	Fingerprint string `json:"fingerprint"`
}

type Burn struct {
	Mint
}

type ValidityInterval struct {
	InvalidBefore    Quantity `json:"invalid_before"`
	InvalidHereafter Quantity `json:"invalid_hereafter"`
}

// -----------------------------------------

type CreateTransactionRequest struct {
	Passphrase string    `json:"passphrase"` // required
	Payments   []Payment `json:"payments"`   // required
	Withdrawal string    `json:"withdrawal"`
	Metadata   Metadata  `json:"metadata"`
	TimeToLive Quantity  `json:"time_to_live"`
}

type Payment struct {
	Address        string   `json:"address"`
	Amount         Quantity `json:"amount"`
	Assets         []Asset  `json:"assets"`
	DerivationPath []string `json:"derivation_path"`
}

type Metadata map[string]MetadataValue

type MetadataValue struct {
	String string          `json:"string,omitempty"`
	Int    uint64          `json:"int,omitempty"`
	Bytes  string          `json:"bytes,omitempty"`
	List   []MetadataValue `json:"list,omitempty"`
	Map    []MetadataMap   `json:"map,omitempty"`
}

type MetadataMap struct {
	K MetadataValue `json:"k"`
	V MetadataValue `json:"v"`
}

type WalletAsset struct {
	PolicyID      string              `json:"policy_id"`
	AssetName     string              `json:"asset_name"`
	Fingerprint   string              `json:"fingerprint"`
	Metadata      WalletAssetMetadata `json:"metadata"`
	MetadataError string              `json:"metadata_error"`

	Address string `json:"address"`
	Price   uint64 `json:"price"`

	AssetUnit     string `json:"asset_unit"`
	AssetQuantity uint64 `json:"asset_quantity"`
	AssetDecimals uint64 `json:"asset_decimals"`
	Fee           uint64 `json:"fee"`
	Deposit       uint64 `json:"deposit"`
	ProcessingFee uint64 `json:"processing_fee"`

	TotalQuantity uint64 `json:"total_quantity"`

	RewardAddress string `json:"reward_address"`
}

type WalletAssetMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Ticker      string `json:"ticker"`
	Decimals    uint32 `json:"decimals"`
	Url         string `json:"url"`
	Logo        string `json:"logo"`
}

type WalletAddress struct {
	ID             string   `json:"id"`
	State          string   `json:"state"`
	DerivationPath []string `json:"derivation_path"`
}

type CreateWalletRequest struct {
	Name     string   `json:"name"`
	Mnemonic []string `json:"mnemonic_sentence"`
	// MnemonicSecondFactor string `json:"mnemonic_second_factor"`
	Passphrase     string `json:"passphrase"`
	AddressPoolGap uint64 `json:"address_pool_gap"`
}

// --------------------------------------------------------

type NetworkInfo struct {
	NetworkInfo struct {
		NetworkID     string `json:"network_id"`
		ProtocolMagic uint64 `json:"protocol_magic"`
	} `json:"network_info"`
	NetworkTip struct {
		AbsoluteSlotNumber uint64 `json:"absolute_slot_number"`
		EpochNumber        uint64 `json:"epoch_number"`
		SlotNumber         uint64 `json:"slot_number"`
		Time               string `json:"time"`
	} `json:"network_tip"`
	NextEpoch struct {
		EpochNumber    uint64 `json:"epoch_number"`
		EpochStartTime string `json:"epoch_start_time"`
	} `json:"next_epoch"`
	NodeEra string `json:"node_era"`
	NodeTip struct {
		AbsoluteSlotNumber uint64   `json:"absolute_slot_number"`
		EpochNumber        uint64   `json:"epoch_number"`
		Height             Quantity `json:"height"`
		SlotNumber         uint64   `json:"slot_number"`
		Time               string   `json:"time"`
	} `json:"node_tip"`
	SyncProgress struct {
		Status   string   `json:"status"`
		Progress Quantity `json:"progress"`
	} `json:"sync_progress"`
	WalletMode string `json:"wallet_mode"`
}

type Wallets []WalletResponse

// type Wallet struct {
// 	ID string `json:"id"`
// 	// Address    string  `json:"address"`
// 	// Passphrase string  `json:"-"`
// 	// Assets     []Asset `json:"assets"`
// 	State struct {
// 		Status   string   `json:"status"`
// 		Progress Quantity `json:"state"`
// 	}
// }
