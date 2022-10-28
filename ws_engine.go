package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type WsEngine struct {
	conns        map[*WsConn]bool
	broadcast    chan []byte
	connected    chan *WsConn
	disconnected chan *WsConn
}

func newWsEngine() *WsEngine {
	return &WsEngine{
		conns:        make(map[*WsConn]bool),
		broadcast:    make(chan []byte),
		connected:    make(chan *WsConn),
		disconnected: make(chan *WsConn),
	}
}

func (wsEngine *WsEngine) sendMessage(msgJson []byte) {
	var msg Message
	if err := json.Unmarshal(msgJson, &msg); err != nil {
		log.Printf("[ERROR] Error on unmarshal JSON message %s", err)
	}

	for wsConn := range wsEngine.conns {
		if wsConn.channelId != msg.ChannelId {
			continue
		}

		select {
		case wsConn.send <- msgJson:
			if err := wsConn.conn.WriteMessage(websocket.TextMessage, msgJson); err != nil {
				return
			}
		default:
			close(wsConn.send)
			delete(wsEngine.conns, wsConn)
		}
	}
}

func (wsEngine *WsEngine) launch() {
	for {
		select {
		case conn := <-wsEngine.connected:
			log.Printf("[INFO] connected")
			wsEngine.conns[conn] = true

			msg := &Message{
				MsgType:   "connected",
				Content:   "进入了房间",
				ChannelId: conn.channelId,
				UserId:    conn.userId,
				Date:      time.Now().Unix(),
			}
			msgBytes, _ := json.Marshal(msg)
			wsEngine.sendMessage(msgBytes)
		case conn := <-wsEngine.disconnected:
			log.Printf("[INFO] disconnected")
			if _, ok := wsEngine.conns[conn]; ok {
				close(conn.send)
				delete(wsEngine.conns, conn)
			}
		case msgBytes := <-wsEngine.broadcast:
			wsEngine.sendMessage(msgBytes)
		}
	}
}
