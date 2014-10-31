package sia

// wallet.go contains things like signatures and scans the blockchain for
// available funds that can be spent.

import (
	"crypto/ecdsa"
	"errors"
)

type Wallet struct {
	SecretKey       *ecdsa.PrivateKey
	SpendConditions SpendConditions
	CoinAddress     CoinAddress         // 1 of 1 spend using the secret key, no timelock.
	OwnedOutputs    map[OutputID]Output // All outputs to CoinAddress
}

func CreateWallet() (w *Wallet, err error) {
	w = new(Wallet)

	var pk PublicKey
	w.SecretKey, pk, err = GenerateKeyPair()
	w.SpendConditions.PublicKeys = append(w.SpendConditions.PublicKeys, pk)
	w.SpendConditions.NumSignatures = 1
	// w.CoinAddress = sc.GetAddress()

	return
}

// Scans all unspent transactions and adds the ones that are spendable by this
// wallet.
func (w *Wallet) Scan(state *State) {
	w.OwnedOutputs = make(map[OutputID]Output)
	for id, output := range state.ConsensusState.UnspentOutputs {
		if output.SpendHash == w.CoinAddress {
			w.OwnedOutputs[id] = output
		}
	}
}

// Takes a new address, and an amount to send, and adds outputs until the
// amount is reached. Then sends leftovers back to self.
//
// Should probably call scan before doing this...?
func (w *Wallet) SpendCoins(amount Currency, address CoinAddress, state *State) (t Transaction, err error) {
	if amount == Currency(0) {
		err = errors.New("Cannot send 0 coins")
		return
	}

	// Scan blockchain for outputs.
	w.Scan(state)

	t.Version = 1

	total := Currency(0)
	for id, output := range w.OwnedOutputs {
		if total > amount {
			break
		}

		newInput := Input{
			OutputID:        id,
			SpendConditions: w.SpendConditions,
		}
		t.Inputs = append(t.Inputs, newInput)

		total += output.Value
	}

	if total < amount {
		err = errors.New("insufficient funds")
		return
	}

	t.Outputs = []Output{
		{Value: amount, SpendHash: address},
		{Value: total - amount, SpendHash: w.CoinAddress},
	}

	// Still need to sign the transaction.

	return
}
