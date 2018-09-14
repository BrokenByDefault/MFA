// +build js

package main

import (
	"github.com/brokenbydefault/Nanollet/GUI/guitypes"
	"github.com/brokenbydefault/Nanollet/GUI/App/DOM"
	"strings"
	"honnef.co/go/js/dom"
)

func InitApplication(w dom.Document, app guitypes.Application) {
	StartApplication(w, app)

	if !app.HaveSidebar() {
		return
	}

	controlbar, _ := DOM.SelectFirstElement(w, ".control")
	modulebutton := DOM.CreateElementAppendTo("button", "", "", "", controlbar)

	DOM.CreateElementAppendTo("span", strings.Title(app.Name()), "title", "", modulebutton)
	DOM.CreateElementAppendTo("span", "", "pointer", "", modulebutton)

	modulebutton.AddEventListener("click", false, func(_ dom.Event) {
		ViewApplication(w, app)
	})

	aside := DOM.CreateElementAppendTo("aside", "", "application", app.Name(), controlbar)

	for _, p := range app.Pages() {
		page := p

		controlbutton := DOM.CreateElementAppendTo("button", "", strings.Title(page.Name()), "", aside)
		block := DOM.CreateElementAppendTo("span", "", "block", "", controlbutton)

		DOM.CreateElementAppendTo("icon", "", "icon-"+page.Name(), "", block)
		DOM.CreateElementAppendTo("span", strings.Title(page.Name()), "title", "", block)
		DOM.CreateElementAppendTo("span", "", "pointer", "", block)

		controlbutton.AddEventListener("click", false, func(_ dom.Event) {
			ViewPage(w, page)
		})
	}
}

func ViewApplication(w dom.Document, app guitypes.Application) error {
	DOM.ApplyForAll(w, ".application, [page]", DOM.HideElement)

	if app.HaveSidebar() {
		el, _ := DOM.SelectFirstElement(w, "body")
		el.SetAttribute("class", "")
	}

	DOM.ApplyForAll(w, ".application#"+app.Name(), DOM.ShowElement)

	return ViewPage(w, app.Pages()[0])
}

func StartApplication(w dom.Document, app guitypes.Application) {

	for _, p := range app.Pages() {
		sector := DOM.SetSector(p)
		page := p

		btns, err := sector.SelectAllElement(w, `button, input[type="submit"]`)
		if err != nil {
			return
		}

		for _, b := range btns {
			btn := b
			btnClass := btn.GetAttribute("class")
			btn.AddEventListener("click", false, func(_ dom.Event) {
				go func() {
					sector.ApplyForAll(w, `button, input[type="submit"]`, DOM.DisableElement)
					page.OnContinue(w, btnClass)
					sector.ApplyForAll(w, `button, input[type="submit"]`, DOM.EnableElement)
				}()
			})
		}
	}

	return
}

func ViewPage(w dom.Document, page guitypes.Page) error {

	DOM.ApplyForAll(w, ".control button", DOM.UnvisitedElement)
	DOM.ApplyForIt(w, ".control button."+strings.Title(page.Name()), DOM.VisitedElement)

	DOM.ApplyForAll(w, "[page]", DOM.HideElement)
	page.OnView(w)

	DOM.ApplyForIt(w, "[page=\""+page.Name()+"\"]", DOM.ShowElement)

	return nil
}
