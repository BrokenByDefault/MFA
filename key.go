package MFA

import (
	"golang.org/x/crypto/curve25519"
	"crypto/rand"
	"github.com/brokenbydefault/Nanollet/Wallet"
)

type DeviceSecret struct {
	Wallet.SecretKey
}

func GenerateDevice() *DeviceSecret {
	_, sk, err := Wallet.GenerateRandomKeyPair()
	if err != nil {
		panic(err)
	}

	return &DeviceSecret{sk}
}

func NewDevice(key []byte) (*DeviceSecret) {
	return &DeviceSecret{Wallet.SecretKey(key)}
}

type EphemeralSecret struct {
	sk [32]byte
}

func NewEphemeralSecret() *EphemeralSecret {
	e := &EphemeralSecret{}

	if _, err := rand.Read(e.sk[:]); err != nil {
		panic("impossible create key")
	}

	return e
}

func (e *EphemeralSecret) PublicKey() (pk [32]byte) {
	curve25519.ScalarBaseMult(&pk, &e.sk)
	return pk
}

func (e *EphemeralSecret) Exchange(partner []byte) (sharedSecretKey [32]byte) {
	var p [32]byte
	copy(p[:], partner)

	curve25519.ScalarMult(&sharedSecretKey, &e.sk, &p)
	return sharedSecretKey
}
