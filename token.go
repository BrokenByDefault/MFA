package MFA

import (
	"github.com/brokenbydefault/Nanollet/Wallet"
	"errors"
)

type Token []byte

func NewSeedFY() (Wallet.SeedFY, error) {
	return Wallet.NewCustomFY(Wallet.V0, Wallet.MFA, 2, 10)
}

func ReadSeedFY(seedfy string, pass string) (Wallet.Seed, error) {
	sf, err := Wallet.ReadSeedFY(seedfy)
	if err != nil {
		return nil, err
	}

	if !sf.IsValid(Wallet.V0, Wallet.MFA) {
		return nil, errors.New("seedfy not intended to be used in MFA")
	}

	return sf.RecoverSeed(pass, nil), nil
}

func RecoverToken(seed Wallet.Seed, index uint32) (token Token, err error) {
	_, sk, err := seed.CreateKeyPair(Wallet.Base, index)

	return Token(sk[:32]), err
}
