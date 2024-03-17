package web

import (
	"fmt"
	"github.com/e-gun/HipparchiaGoServer/internal/launch"
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

// ReadUUIDCookie - find the ID of the client
func ReadUUIDCookie(c echo.Context) string {
	cookie, err := c.Cookie("ID")
	if err != nil {
		id := writeUUIDCookie(c)
		return id
	}
	id := cookie.Value

	if !vaults.AllSessions.IsInVault(id) {
		vaults.AllSessions.InsertSess(launch.MakeDefaultSession(id))
	}

	return id
}

// writeUUIDCookie - set the ID of the client
func writeUUIDCookie(c echo.Context) string {
	cookie := new(http.Cookie)
	cookie.Name = "ID"
	cookie.Path = "/"
	cookie.Value = uuid.New().String()
	cookie.Expires = time.Now().Add(4800 * time.Hour)
	c.SetCookie(cookie)
	msg.TMI(fmt.Sprintf("writeUUIDCookie() - new ID set: %s", cookie.Value))
	return cookie.Value
}
