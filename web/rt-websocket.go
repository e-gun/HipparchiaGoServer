//    HipparchiaGoServer
//    Copyright: E Gunderson 2022-24
//    License: GNU GENERAL PUBLIC LICENSE 3
//        (see LICENSE in the top level directory of the distribution)

package web

import (
	"github.com/e-gun/HipparchiaGoServer/internal/vaults"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	Upgrader = websocket.Upgrader{}
)

//
// THE ROUTE
//

// RtWebsocket - progress info for a search (multiple clients client at a time)
func RtWebsocket(c echo.Context) error {
	const (
		FAILCON = "RtWebsocket(): ws connection failed"
	)

	user := ReadUUIDCookie(c)
	if !vaults.AllAuthorized.Check(user) {
		return nil
	}

	ws, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		msg.NOTE(FAILCON)
		return nil
	}

	progresspoll := &vaults.WSClient{
		Conn: ws,
		Pool: vaults.WebsocketPool,
	}

	vaults.WebsocketPool.Add <- progresspoll
	progresspoll.ReceiveID()
	progresspoll.WSMessageLoop()
	vaults.WebsocketPool.Remove <- progresspoll
	return nil
}
