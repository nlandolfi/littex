//go:build js

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/nlandolfi/lit/web-client/components/app"
	"github.com/spinsrv/browser"
	"github.com/spinsrv/browser/js"
)

const ClientVersion = "0.0.2"
const LocalStorageStateKey = "gba-web-client-state-cache"

func main() {
	// TODO: catch panics
	// js.DefaultBrowser.Document().Body().SetInnerHTML(template.HTML("The app has crashed. Please refresh."))

	var s app.State
	s.Theme = app.DefaultTheme
	s.ClientVersion = ClientVersion

	// try to load state from local storage
	if ss := js.DefaultLocalStorage.Get(LocalStorageStateKey + ":" + ClientVersion); ss != "" {
		if err := json.NewDecoder(bytes.NewBufferString(ss)).Decode(&s); err != nil {
			log.Print("error decoding: %+v", err)
			log.Print("dropping state")
			js.DefaultLocalStorage.Del(LocalStorageStateKey + ":" + ClientVersion)
		}
	}

	if s.PrivateKey != nil && s.PrivateKey.ExpiresAt.Before(time.Now()) {
		log.Printf("key expired; dropping")
		s.PrivateKey = nil
	}

	// TODO: is this right? - NCL 2/1/22
	s.Rewire()

	m := &browser.Mounter{
		Document: js.DefaultBrowser.Document(),
		Root:     js.DefaultBrowser.Document().Body(),
	}

	go browser.Dispatch(app.EventInitialize{})

	for e := range browser.Events {
		s.Handle(e)

		if err := m.Mount(app.View(&s)); err != nil {
			panic(err)
		}

		s.LastWrittenAt = time.Now()
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(&s); err != nil {
			log.Printf("error encoding state to json: %+v", err)
		} else {
			js.DefaultLocalStorage.Put(LocalStorageStateKey+":"+ClientVersion, b.String())
		}
	}

	/*
		for range time.NewTicker(100 * time.Millisecond).C {
			n := app.View(&s)
			var bb bytes.Buffer
			if err := json.NewEncoder(&bb).Encode(&n); err != nil {
				log.Fatal(err)
			}
			//		log.Print(bb.String())
			if err := m.Mount(n); err != nil {
				log.Fatalf("error mounting: %v", err)
			}

			s.LastWrittenAt = time.Now()
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(&s); err != nil {
				log.Printf("error encoding state to json: %+v", err)
			} else {
				js.DefaultLocalStorage.Put(LocalStorageStateKey+":"+ClientVersion, b.String())
			}
		}
	*/
}
