package alchemy

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	xrplrpc "github.com/Peersyst/xrpl-go/xrpl/rpc"
	xrplrpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	xrpltx "github.com/Peersyst/xrpl-go/xrpl/transaction"
	xrpltxtypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	xrplwallet "github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/ebitezion/Nduracore/internal/integrations"
	"github.com/gagliardetto/solana-go"
	solanasystem "github.com/gagliardetto/solana-go/programs/system"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

func (p *Provider) solanaSignerForNetwork(vaultID, network string) (solana.PrivateKey, solana.PublicKey, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyRaw := p.scopedPrivateKey(vaultID, networkKey)
	if privateKeyRaw == "" {
		return nil, solana.PublicKey{}, fmt.Errorf("private key not configured for network %q", networkKey)
	}

	privateKey, err := parseSolanaPrivateKey(privateKeyRaw)
	if err != nil {
		return nil, solana.PublicKey{}, fmt.Errorf("invalid solana private key for network %q", networkKey)
	}
	return privateKey, privateKey.PublicKey(), nil
}

func parseSolanaPrivateKey(raw string) (solana.PrivateKey, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("empty private key")
	}

	if strings.HasPrefix(trimmed, "[") {
		return solana.PrivateKeyFromSolanaKeygenFileBytes([]byte(trimmed))
	}

	if key, err := solana.PrivateKeyFromBase58(trimmed); err == nil {
		return key, nil
	}

	decoded, err := hex.DecodeString(strings.TrimPrefix(trimmed, "0x"))
	if err != nil {
		return nil, err
	}
	switch len(decoded) {
	case ed25519.PrivateKeySize:
		return solana.PrivateKey(decoded), nil
	case ed25519.SeedSize:
		return solana.PrivateKey(ed25519.NewKeyFromSeed(decoded)), nil
	default:
		return nil, fmt.Errorf("unexpected key length")
	}
}

func (p *Provider) broadcastTransferSolana(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	network := strings.ToLower(strings.TrimSpace(req.Network))
	native, err := p.supportsNativeAsset(ctx, "solana", network, req.Asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !native {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("solana broadcast currently supports native SOL only")
	}
	if req.AmountMinor <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	privateKey, signerAddress, err := p.solanaSignerForNetwork(req.VaultID, network)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	rpcURL, err := p.rpcURLForNetwork(network)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	client := solanarpc.New(rpcURL)

	toAddress, err := solana.PublicKeyFromBase58(strings.TrimSpace(req.ToAddress))
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid destination address")
	}

	blockhash, err := client.GetLatestBlockhash(ctx, solanarpc.CommitmentFinalized)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	ix, err := solanasystem.NewTransferInstruction(uint64(req.AmountMinor), signerAddress, toAddress).ValidateAndBuild()
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix},
		blockhash.Value.Blockhash,
		solana.TransactionPayer(signerAddress),
	)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(signerAddress) {
			return &privateKey
		}
		return nil
	}); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	signature, err := client.SendTransactionWithOpts(ctx, tx, solanarpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: solanarpc.CommitmentConfirmed,
	})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	return integrations.BroadcastTransferResult{TxHash: signature.String()}, nil
}

func (p *Provider) stellarAddressForNetwork(vaultID, network string) (string, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyRaw := p.scopedPrivateKey(vaultID, networkKey)
	if privateKeyRaw == "" {
		return "", fmt.Errorf("private key not configured for network %q", networkKey)
	}

	full, err := keypair.ParseFull(privateKeyRaw)
	if err != nil {
		return "", fmt.Errorf("invalid stellar secret for network %q", networkKey)
	}
	return full.Address(), nil
}

func (p *Provider) broadcastTransferStellar(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	networkName := strings.ToLower(strings.TrimSpace(req.Network))
	native, err := p.supportsNativeAsset(ctx, "stellar", networkName, req.Asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !native {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("stellar broadcast currently supports native XLM only")
	}
	if req.AmountMinor <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	privateKeyRaw := p.scopedPrivateKey(req.VaultID, networkName)
	if privateKeyRaw == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("private key not configured for network %q", networkName)
	}
	kp, err := keypair.ParseFull(privateKeyRaw)
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid stellar secret for network %q", networkName)
	}

	rpcURL, err := p.rpcURLForNetwork(networkName)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	client := horizonclient.Client{
		HorizonURL: rpcURL,
		HTTP:       p.httpClient,
	}

	sourceAccount, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: kp.Address()})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	amountMajor := formatFixedAmount(req.AmountMinor, 7)
	payment := txnbuild.Payment{
		Destination: strings.TrimSpace(req.ToAddress),
		Amount:      amountMajor,
		Asset:       txnbuild.NativeAsset{},
	}

	unsignedTx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount:        &sourceAccount,
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{&payment},
		BaseFee:              txnbuild.MinBaseFee,
		Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
	})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	passphrase := network.TestNetworkPassphrase
	if strings.Contains(networkName, "mainnet") {
		passphrase = network.PublicNetworkPassphrase
	}
	signedTx, err := unsignedTx.Sign(passphrase, kp)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	resp, err := client.SubmitTransaction(signedTx)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	return integrations.BroadcastTransferResult{TxHash: resp.Hash}, nil
}

func (p *Provider) xrplAddressForNetwork(vaultID, network string) (string, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyRaw := p.scopedPrivateKey(vaultID, networkKey)
	if privateKeyRaw == "" {
		return "", fmt.Errorf("private key not configured for network %q", networkKey)
	}
	w, err := xrplwallet.FromSeed(privateKeyRaw, "")
	if err != nil {
		return "", fmt.Errorf("invalid xrpl seed for network %q", networkKey)
	}
	return w.ClassicAddress.String(), nil
}

func (p *Provider) broadcastTransferXRPL(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	networkName := strings.ToLower(strings.TrimSpace(req.Network))
	native, err := p.supportsNativeAsset(ctx, "xrpl", networkName, req.Asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !native {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("xrpl broadcast currently supports native XRP only")
	}
	if req.AmountMinor <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	privateKeyRaw := p.scopedPrivateKey(req.VaultID, networkName)
	if privateKeyRaw == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("private key not configured for network %q", networkName)
	}
	w, err := xrplwallet.FromSeed(privateKeyRaw, "")
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid xrpl seed for network %q", networkName)
	}

	rpcURL, err := p.rpcURLForNetwork(networkName)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	cfg, err := xrplrpc.NewClientConfig(rpcURL)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	client := xrplrpc.NewClient(cfg)

	payment := &xrpltx.Payment{
		BaseTx: xrpltx.BaseTx{
			Account: w.GetAddress(),
		},
		Destination: xrpltxtypes.Address(strings.TrimSpace(req.ToAddress)),
		Amount:      xrpltxtypes.XRPCurrencyAmount(req.AmountMinor),
		DeliverMax:  xrpltxtypes.XRPCurrencyAmount(req.AmountMinor),
	}

	resp, err := client.SubmitTxAndWait(payment.Flatten(), &xrplrpctypes.SubmitOptions{
		Autofill: true,
		Wallet:   &w,
	})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	return integrations.BroadcastTransferResult{TxHash: resp.Hash.String()}, nil
}

func isNativeAssetForFamily(family, asset string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(asset))
	switch family {
	case "solana":
		return normalized == "" || normalized == "SOL"
	case "stellar":
		return normalized == "" || normalized == "XLM"
	case "xrpl":
		return normalized == "" || normalized == "XRP"
	case "tron":
		return normalized == "" || normalized == "TRX"
	case "aptos":
		return normalized == "" || normalized == "APT"
	case "sui":
		return normalized == "" || normalized == "SUI"
	default:
		return false
	}
}

func formatFixedAmount(amountMinor int64, decimals int) string {
	if decimals <= 0 {
		return strconv.FormatInt(amountMinor, 10)
	}
	sign := ""
	value := amountMinor
	if value < 0 {
		sign = "-"
		value = -value
	}

	base := int64(1)
	for i := 0; i < decimals; i++ {
		base *= 10
	}

	whole := value / base
	fraction := value % base
	if fraction == 0 {
		return sign + strconv.FormatInt(whole, 10)
	}
	frac := strconv.FormatInt(fraction, 10)
	if len(frac) < decimals {
		frac = strings.Repeat("0", decimals-len(frac)) + frac
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return sign + strconv.FormatInt(whole, 10)
	}
	return sign + strconv.FormatInt(whole, 10) + "." + frac
}
