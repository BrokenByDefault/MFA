// +build ignore

package MFA

import (
	"github.com/brokenbydefault/Nanollet/Wallet"
	"github.com/brokenbydefault/Nanollet/Util"
	"errors"
	"bytes"
)

type VERSION uint8

const (
	V1 VERSION = iota
)

type Capsule struct {
	Version   VERSION
	Signature []byte
	Device    Wallet.PublicKey
	Sender    [32]byte
	Receiver  [32]byte
	Token     []byte
}

func EncapsulateMFA(seed Wallet.Seed, receiver [32]byte) (capsule *Capsule, err error) {

	token, err := RecoverToken(seed)
	if err != nil {
		return capsule, err
	}

	ephemeral, err := NewEphemeral()
	if err != nil {
		return nil, err
	}

	capsule = &Capsule{}
	capsule.Version = V1
	capsule.Receiver = receiver
	capsule.Sender = ephemeral.PublicKey()
	capsule.Token = ephemeral.Crypt(ephemeral.NonceSender(receiver), token, receiver)

	return
}

// HEX( VERSION || SIGNATURE || ED25519 PUBLIC-KEY || X25519sender PUBLIC-KEY || X25519receiver PUBLIC-KEY || ENCRYPTED-TOKEN )
func (capsule *Capsule) Encode(device Wallet.SecretKey) (hex string, err error) {
	var caps = make([]byte, 193)

	capsule.Device, err = device.PublicKey()
	if err != nil {
		return
	}

	copy(caps, []byte{byte(capsule.Version)})
	copy(caps[65:], capsule.Device)
	copy(caps[97:], capsule.Sender[:])
	copy(caps[129:], capsule.Receiver[:])
	copy(caps[161:], capsule.Token)

	capsule.Signature, err = device.CreateSignature(caps[65:])
	if err != nil {
		return
	}

	copy(caps[1:], capsule.Signature)

	return Util.SecureHexEncode(caps), nil
}

func UncapsulateMFA(receiver EphemeralKey, hex string) (capsule *Capsule, err error) {
	var caps = make([]byte, 193)
	var receiverpk = receiver.PublicKey()

	decoded, _ := Util.SecureHexDecode(hex)
	copy(caps, decoded)

	capsule = &Capsule{}
	capsule.Version = VERSION(caps[0])
	capsule.Signature = caps[1:65]
	capsule.Device = Wallet.PublicKey(caps[65:97])

	copy(capsule.Sender[:], caps[97:129])
	copy(capsule.Receiver[:], caps[129:161])

	if !capsule.Device.CompareSignature(caps[65:], capsule.Signature) {
		return nil, errors.New("invalid signature")
	}

	// Public-Data doesn't need to be constant time
	if !bytes.Equal(capsule.Receiver[:], receiverpk[:]) {
		return nil, errors.New("invalid receiver")
	}

	capsule.Token = receiver.Crypt(receiver.NonceReceiver(capsule.Sender), caps[161:], capsule.Sender)
	return
}
