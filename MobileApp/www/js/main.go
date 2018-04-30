// +build js

//go:generate gopherjs build -o js.js
package main

import (
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
	"github.com/brokenbydefault/MFA"
	"github.com/brokenbydefault/Nanollet/GUI/guitypes"
	"github.com/brokenbydefault/Nanollet/GUI/Front"
	"github.com/brokenbydefault/Nanollet/GUI/App/DOM"
	"github.com/brokenbydefault/Nanollet/Wallet"
	"github.com/jaracil/goco/nativestorage"
	"github.com/jaracil/goco/barcodescanner"
	"strings"
	"github.com/brokenbydefault/Nanollet/RPC"
	"github.com/brokenbydefault/Nanollet/RPC/Connectivity"
	"github.com/brokenbydefault/Nanollet/Util"
	"github.com/jaracil/goco"
	"github.com/jaracil/goco/dialogs"
)

func OnDeviceReady(cb func()) {
	js.Global.Get("document").Call("addEventListener", "deviceready", cb, false)
}

var root = dom.GetWindow().Document()

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

	seed, err := MFA.NewSeedFY()
	if err != nil {
		return
	}

	textarea, _ := page.SelectFirstElement(w, ".seed")
	textarea.SetTextContent(seed.Encode())
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
		dialogs.Alert("Imposible to store the seed", "Error", "Ok")
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
		dialogs.Alert("Imposible to store the seed", "Error", "Ok")
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
	seedfy, err := MFA.ReadSeedFY(seed, password)
	if err != nil {
		return
	}

	token, err := MFA.RecoverToken(seedfy, 0)
	if err != nil {
		return
	}

	devicehex, _ := nativestorage.GetString("DEVICE")
	device, ok := Util.SecureHexDecode(devicehex)
	if !ok {
		return
	}

	sender, err := MFA.NewSender(MFA.NewDevice(device))
	if err != nil {
		return
	}

	qrcode, err := barcodescanner.NewScanner().Scan()
	if err != nil {
		dialogs.Alert("Impossible to use the scanner", "Error", "Ok")
		return
	}

	receiver, ok := Util.SecureHexDecode(qrcode.Text)
	if len(receiver) != 32 || !ok {
		dialogs.Alert("Impossible to decode the key", "Error", "Ok")
		return
	}

	env, err := sender.CreateEnvelope(receiver, token)
	if err != nil {
		return
	}

	err = RPCClient.SendToken(Connectivity.Socket, receiver, env)
	if err != nil {
		dialogs.Alert("Impossible to connect with our server", "Error", "Ok")
		return
	}

	return
}

func main() {
	goco.OnDeviceReady(func() {
		go func() {
			if _, err := nativestorage.GetString("DEVICE"); err != nil {
				nativestorage.SetItem("DEVICE", Util.SecureHexEncode(MFA.GenerateDevice().SecretKey))
			}

			InitApplication(root, &AccountApp{})
			ViewApplication(root, &AccountApp{})

			if _, err := nativestorage.GetString("SEEDFY"); err == nil {
				ViewPage(root, &PagePassword{})
			}

			Connectivity.Socket.StartWebsocket()
		}()
	})
}
