package spacedesk

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type SpaceDeskWSMessage map[string]any

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var wsClients = make(map[*websocket.Conn]bool)
var wsMutex sync.Mutex

func broadcastSpaceDeskMessage(msg SpaceDeskWSMessage) {
	wsMutex.Lock()
	defer wsMutex.Unlock()
	for client := range wsClients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(wsClients, client)
		}
	}
}

func SpaceDeskWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Não foi possível fazer upgrade para websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	wsMutex.Lock()
	wsClients[conn] = true
	wsMutex.Unlock()

	for {
		msg := SpaceDeskWSMessage{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		broadcastSpaceDeskMessage(msg)
	}

	wsMutex.Lock()
	delete(wsClients, conn)
	wsMutex.Unlock()
}
