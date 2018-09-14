// +build js

//go:generate gopherjs build -o js.js
package main

import (
	"honnef.co/go/js/dom"
	"github.com/brokenbydefault/Nanollet/GUI/guitypes"
	"github.com/brokenbydefault/Nanollet/GUI/Front"
	"github.com/brokenbydefault/Nanollet/GUI/App/DOM"
	"github.com/brokenbydefault/Nanollet/Wallet"
	"github.com/jaracil/goco/nativestorage"
	"github.com/jaracil/goco/barcodescanner"
	"strings"
	"github.com/brokenbydefault/Nanollet/Util"
	"github.com/jaracil/goco"
	"github.com/jaracil/goco/dialogs"
	"github.com/brokenbydefault/Nanollet/TwoFactor"
	"encoding/binary"
	"bytes"
)

var root = dom.GetWindow().Document()

func main() {
	goco.OnDeviceReady(func() {
		go func() {
			if _, err := nativestorage.GetString("DEVICE"); err != nil {
				_, sk, _ := Wallet.GenerateRandomKeyPair()
				nativestorage.SetItem("DEVICE", Util.SecureHexEncode(sk[:]))
			}

			InitApplication(root, &AccountApp{})
			ViewApplication(root, &AccountApp{})

			if _, err := nativestorage.GetString("SEEDFY"); err == nil {
				ViewPage(root, &PagePassword{})
			}

		}()
	})
}

type AccountApp guitypes.App

func (c *AccountApp) Name() string {
	return "account"
}

func (c *AccountApp) HaveSidebar() bool {
	return false
}

func (c *AccountApp) Display() Front.HTMLPAGE {
	return ""
}

func (c *AccountApp) Pages() []guitypes.Page {
	return []guitypes.Page{
		&PageIndex{},
		&PageGenerate{},
		&PageImport{},
		&PagePassword{},
	}
}

type PageIndex guitypes.Sector

func (c *PageIndex) Name() string {
	return "index"
}

func (c *PageIndex) OnView(w dom.Document) {
	// no-op
}

func (c *PageIndex) OnContinue(w dom.Document, action string) {
	switch action {
	case "genSeed":
		ViewPage(w, &PageGenerate{})
	case "importSeed":
		ViewPage(w, &PageImport{})
	}
}

type PageGenerate guitypes.Sector

func (c *PageGenerate) Name() string {
	return "generate"
}

func (c *PageGenerate) OnView(w dom.Document) {
	page := DOM.SetSector(c)

	seed, err := TwoFactor.NewSeedFY()
	if err != nil {
		return
	}

	textarea, _ := page.SelectFirstElement(w, ".seed")
	textarea.SetTextContent(seed.String())
	DOM.ReadOnlyElement(textarea)
}

func (c *PageGenerate) OnContinue(w dom.Document, _ string) {
	page := DOM.SetSector(c)

	seed, err := page.GetStringValue(w, ".seed")
	if strings.TrimSpace(seed) == "" || err != nil {
		dialogs.Alert("Invalid seed", "Error", "Ok")
		return
	}

	err = nativestorage.SetItem("SEEDFY", seed)
	if err != nil {
		dialogs.Alert("Impossible to store the seed", "Error", "Ok")
		return
	}

	ViewPage(w, &PagePassword{})
}

type PageImport guitypes.Sector

func (c *PageImport) Name() string {
	return "import"
}

func (c *PageImport) OnView(w dom.Document) {
	//no-op
}

func (c *PageImport) OnContinue(w dom.Document, _ string) {
	page := DOM.SetSector(c)

	seed, err := page.GetStringValue(root, ".seed")
	if err != nil || seed == "" {
		return
	}

	sf, err := Wallet.ReadSeedFY(seed)
	if err != nil {
		dialogs.Alert("Invalid seed", "Error", "Ok")
		return
	}

	if ok := sf.IsValid(Wallet.Version(sf.Version), Wallet.MFA); !ok {
		dialogs.Alert("Unsupported seed", "Error", "Ok")
		return
	}

	if err = nativestorage.SetItem("SEEDFY", seed); err != nil {
		dialogs.Alert("Impossible to store the seed", "Error", "Ok")
		return
	}

	ViewPage(w, &PagePassword{})
}

type PagePassword guitypes.Sector

func (c *PagePassword) Name() string {
	return "password"
}

func (c *PagePassword) OnView(w dom.Document) {
	// no-op
}

func (c *PagePassword) OnContinue(w dom.Document, _ string) {
	page := DOM.SetSector(c)

	password, err := page.GetStringValue(w, ".password")
	if err != nil || len(password) < 8 {
		dialogs.Alert("Password is too short", "Error", "Ok")
		return
	}

	DOM.ApplyForIt(w, ".password", DOM.ClearValue)

	seed, _ := nativestorage.GetString("SEEDFY")
	token, err := TwoFactor.NewToken(seed, []byte(password))
	if err != nil {
		return
	}

	devicehex, _ := nativestorage.GetString("DEVICE")
	deviceb, ok := Util.SecureHexDecode(devicehex)
	if !ok {
		return
	}

	device := Wallet.NewSecretKey(deviceb)

	qrcode, err := barcodescanner.NewScanner().Scan()
	if err != nil {
		dialogs.Alert("Impossible to use the scanner", "Error", "Ok")
		return
	}

	request := TwoFactor.Request{}
	receiver, ok := Util.SecureHexDecode(qrcode.Text)
	if err := binary.Read(bytes.NewReader(receiver), binary.BigEndian, &request); err != nil || !ok {
		dialogs.Alert("Impossible to decode the key", "Error", "Ok")
		return
	}

	if err := TwoFactor.ReplyRequest(&device, token, request); err != nil {
		dialogs.Alert("Impossible to connect with our server", "Error", "Ok")
		return
	}

	return
}
