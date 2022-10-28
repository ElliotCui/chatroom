package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	writeWait      = 120 * time.Second
	pongWait       = 60 * time.Second
	pingPerriod    = pongWait * 9 / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(r *http.Request) bool {
	return true
}

type WsConn struct {
	channelId string
	userId    string
	wsEngine  *WsEngine
	conn      *websocket.Conn
	send      chan []byte
}

type Message struct {
	UserId    string `json:"userId"`
	ChannelId string `json:"channelId"`
	MsgType   string `json:"msgType"`
	Content   string `json:"content"`
	Date      int64  `json:"date"`
}

func (wsConn *WsConn) readPipe() {
	defer func() {
		wsConn.wsEngine.disconnected <- wsConn
		wsConn.conn.Close()
	}()
	wsConn.conn.SetReadLimit(maxMessageSize)
	wsConn.conn.SetReadDeadline(time.Now().Add(pongWait))
	wsConn.conn.SetPongHandler(func(string) error { wsConn.conn.SetReadDeadline((time.Now().Add(pongWait))); return nil })

	for {
		msgType, msg, err := wsConn.conn.ReadMessage()
		log.Printf("[INFO] ReadMessage: %v", msgType)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR] error: %v", err)
			}
			break
		}
		msg = bytes.TrimSpace((bytes.Replace(msg, newline, space, -1)))
		wsConn.wsEngine.broadcast <- msg
	}
}

func (wsConn *WsConn) writePipe() {
	ticker := time.NewTicker(pingPerriod)
	defer func() {
		ticker.Stop()
		wsConn.conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			wsConn.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsConn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(wsEngine *WsEngine, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] error: %v", err)
		return
	}
	params := r.URL.Query()
	channelId := params.Get("channelId")
	userId := params.Get("userId")
	wsConn := &WsConn{channelId: channelId, userId: userId, wsEngine: wsEngine, conn: conn, send: make(chan []byte, 256)}
	wsConn.wsEngine.connected <- wsConn

	go wsConn.writePipe()
	go wsConn.readPipe()
}
