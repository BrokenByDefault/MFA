package MFA

import (
	"github.com/brokenbydefault/Nanollet/Util"
	"crypto/subtle"
	"github.com/aead/chacha20poly1305"
	"github.com/brokenbydefault/Nanollet/Wallet"
	"errors"
	"github.com/Inkeliz/blakEd25519"
)

var (
	ErrInvalidKey      = errors.New("invalid key")
	ErrInvalidEnvelope = errors.New("invalid envelope")
)

const (
	LengthSealedEnvelope = 208
)

type Sender struct {
	Device    *DeviceSecret
	Ephemeral *EphemeralSecret
}

func NewSender(device *DeviceSecret) (*Sender, error) {
	if len(device.SecretKey) != blakEd25519.PrivateKeySize {
		return nil, ErrInvalidKey
	}

	return &Sender{
		Device:    device,
		Ephemeral: NewEphemeralSecret(),
	}, nil
}

// Envelope = [Sender PK] [Receiver PK] [Encrypted-Capsule] [Tag]
func (sender *Sender) CreateEnvelope(receiverPK []byte, token Token) ([]byte, error) {
	sharedKey := sender.Ephemeral.Exchange(receiverPK)

	cipher, err := chacha20poly1305.NewXCipher(sharedKey[:])
	if err != nil {
		return nil, err
	}

	devicePK, err := sender.Device.PublicKey()
	if err != nil {
		return nil, err
	}

	senderPK := sender.Ephemeral.PublicKey()

	envelope := make([]byte, 64)
	copy(envelope[0:32], senderPK[:])
	copy(envelope[32:64], receiverPK[:])

	capsule := make([]byte, 128)
	copy(capsule[0:32], devicePK[:])
	copy(capsule[32:64], token[:])

	signature, err := sender.Device.CreateSignature(append(envelope[:64], capsule[:64]...))
	if err != nil {
		return nil, err
	}

	copy(capsule[64:128], signature)

	return cipher.Seal(envelope, Util.CreateHash(24, envelope[:64]), capsule, envelope), nil
}

type Receiver struct {
	Ephemeral *EphemeralSecret
}

func NewReceiver() *Receiver {
	return &Receiver{
		Ephemeral: NewEphemeralSecret(),
	}
}

func (receiver *Receiver) OpenEnvelope(envelope []byte) (Token, Wallet.PublicKey, error) {

	// If envelope have enough size
	if len(envelope) != LengthSealedEnvelope {
		return nil, nil, ErrInvalidEnvelope
	}

	receiverPK := receiver.Ephemeral.PublicKey()

	// If destination of the envelope doesn't meets the receiver
	if subtle.ConstantTimeCompare(envelope[32:64], receiverPK[:]) == 0 {
		return nil, nil, ErrInvalidEnvelope
	}

	sharedKey := receiver.Ephemeral.Exchange(envelope[:32])
	cipher, err := chacha20poly1305.NewXCipher(sharedKey[:])
	if err != nil {
		return nil, nil, ErrInvalidEnvelope
	}

	// Retrieve the capsule from encrypted-capsule, error if the poly1305 is invalid
	capsule, err := cipher.Open(nil, Util.CreateHash(24, envelope[:64]), envelope[64:], envelope[:64])
	if err != nil {
		return nil, nil, ErrInvalidEnvelope
	}

	device := Wallet.PublicKey(capsule[:32])
	if !device.CompareSignature(append(envelope[:64], capsule[:64]...), capsule[64:]) {
		return nil, nil, ErrInvalidEnvelope
	}

	return Token(capsule[32:64]), device, nil
}
