package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type Session struct {
	ID             string
	Inclusions     SearchIncExl
	Exclusions     SearchIncExl
	ActiveCorp     map[string]bool
	VariaOK        bool
	IncertaOK      bool
	SpuriaOK       bool
	AvailDBs       map[string]bool `json:"available"`
	VectorVals     bool
	UISettings     bool
	OutPutSettings bool
	Analogyfinder  bool   `json:"analogyfinder"`
	Authorflagging bool   `json:"authorflagging"`
	Authorssummary bool   `json:"authorssummary"`
	Baggingmethod  string `json:"baggingmethod"`
}

func makedefaultsession(id string) Session {
	var s Session
	s.ID = id
	s.ActiveCorp = map[string]bool{"greekcorpus": true, "latincorpus": true, "inscriptioncorpus": true}
	s.VariaOK = true
	s.IncertaOK = true
	s.SpuriaOK = true
	s.AvailDBs = map[string]bool{"greek_dictionary": true, "greek_lemmata": true, "greek_morphology": true, "latin_dictionary": true, "latin_lemmata": true, "latin_morphology": true, "wordcounts_0": true}
	s.Analogyfinder = false

	return s
}

func RtFrontpage(c echo.Context) error {
	fmt.Println("frontpage")
	id := readUUIDCookie(c)
	fmt.Println(id)
	if _, t := sessions[id]; !t {
		sessions[id] = makedefaultsession(id)
	}
	fmt.Println(sessions[id])
	err := c.File("static/html/frontpage.html")
	if err != nil {
		return nil
	}
	return nil
}

func readUUIDCookie(c echo.Context) string {
	cookie, err := c.Cookie("ID")
	if err != nil {
		id := writeUUIDCookie(c)
		return id
	}
	id := cookie.Value
	return id
}

func writeUUIDCookie(c echo.Context) string {
	cookie := new(http.Cookie)
	cookie.Name = "ID"
	cookie.Value = uuid.New().String()
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)
	return cookie.Value
}
