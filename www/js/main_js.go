// +build js

//go:generate gopherjs build main_js.go -o js.js
package main

import (
	"honnef.co/go/js/dom"
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

			win := DOM.NewWindow(dom.GetWindow())
			win.InitApplication(new(AccountApp))
			win.ViewApplication(new(AccountApp))

			if _, err := nativestorage.GetString("SEEDFY"); err == nil {
				win.ViewPage(new(PagePassword))
			}

		}()
	})
}

type AccountApp struct{}

func (c *AccountApp) Name() string {
	return "account"
}

func (c *AccountApp) HaveSidebar() bool {
	return false
}

func (c *AccountApp) Pages() []DOM.Page {
	return []DOM.Page{
		new(PageIndex),
		new(PageGenerate),
		new(PageImport),
		new(PagePassword),
	}
}

type PageIndex struct{}

func (c *PageIndex) Name() string {
	return "index"
}

func (c *PageIndex) OnView(w *DOM.Window, dom *DOM.DOM) {
	// no-op
}

func (c *PageIndex) OnContinue(w *DOM.Window, dom *DOM.DOM, action string) {
	switch action {
	case "genSeed":
		w.ViewPage(new(PageGenerate))
	case "importSeed":
		w.ViewPage(new(PageImport))
	}
}

type PageGenerate struct{}

func (c *PageGenerate) Name() string {
	return "generate"
}

func (c *PageGenerate) OnView(w *DOM.Window, dom *DOM.DOM) {

	seed, err := TwoFactor.NewSeedFY()
	if err != nil {
		return
	}

	textarea, _ := dom.SelectFirstElement(".seed")
	textarea.SetText(seed.String())
	textarea.Apply(DOM.ReadOnlyElement)
}

func (c *PageGenerate) OnContinue(w *DOM.Window, dom *DOM.DOM, action string) {

	seed, err := dom.GetStringValueOf(".seed")
	if strings.TrimSpace(seed) == "" || err != nil {
		dialogs.Alert("Invalid seed", "Error", "Ok")
		return
	}

	err = nativestorage.SetItem("SEEDFY", seed)
	if err != nil {
		dialogs.Alert("Impossible to store the seed", "Error", "Ok")
		return
	}

	w.ViewPage(new(PagePassword))
}

type PageImport struct{}

func (c *PageImport) Name() string {
	return "import"
}

func (c *PageImport) OnView(w *DOM.Window, dom *DOM.DOM) {
	//no-op
}

func (c *PageImport) OnContinue(w *DOM.Window, dom *DOM.DOM, action string) {

	seed, err := dom.GetStringValueOf(".seed")
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

	w.ViewPage(new(PagePassword))
}

type PagePassword struct{}

func (c *PagePassword) Name() string {
	return "password"
}

func (c *PagePassword) OnView(w *DOM.Window, dom *DOM.DOM) {
	// no-op
}

func (c *PagePassword) OnContinue(w *DOM.Window, dom *DOM.DOM, action string) {

	password, err := dom.GetStringValueOf(".password")
	if err != nil || len(password) < 8 {
		dialogs.Alert("Password is too short", "Error", "Ok")
		return
	}

	dom.ApplyFor(".password", DOM.ClearValue)

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
