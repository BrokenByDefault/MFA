package MFA

import (
	"testing"
	"bytes"
	"fmt"
)

func TestEncapsulateMFA(t *testing.T) {
	smartphoneDeviceKey := GenerateDevice()

	computer := NewReceiver()
	smartphone, err := NewSender(smartphoneDeviceKey)
	if err != nil {
		t.Error(err)
	}

	seedfy, err := NewSeedFY()
	if err != nil {
		t.Error(err)
	}

	seed, err := ReadSeedFY(seedfy.Encode(), "123456789")
	if err != nil {
		t.Error(err)
	}

	token, err := RecoverToken(seed, 0)
	if err != nil {
		t.Error(err)
	}

	computerpk := computer.Ephemeral.PublicKey()
	env, err := smartphone.CreateEnvelope(computerpk[:], token)
	if err != nil {
		t.Error(err)
	}

	ctoken, pk, err := computer.OpenEnvelope(env)
	if err != nil {
		t.Error(err)
	}

	smartphoneDeviceKeyPK, _ := smartphoneDeviceKey.PublicKey()
	fmt.Println(bytes.Equal(ctoken, token), bytes.Equal(pk, smartphoneDeviceKeyPK))
}
