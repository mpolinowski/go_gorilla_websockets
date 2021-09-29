package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var wsChn = make(chan WsPayload)
var clients = make(map[WebSocketConnection] string)

var upgradeConnection = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {return true},
}

// Use the renderPage function to render the Home page
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

// Define the response returned from the websocket
type WsJsonResponse struct {
	Action 			string `json: "action"`
	Message 		string `json: "message"`
	MessageType 	string `json: "message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// Websocket type provided by the websocket package
type WebSocketConnection struct {
	*websocket.Conn
}

// Define the response returned from the websocket
type WsPayload struct {
	Action 		string `json: "action"`
	Username	string `json: "username"`
	Message 	string `json: "message"`
	Conn WebSocketConnection `json: "-"`
}

// Upgrade http connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	
	log.Println("Client Connected to Endpoint")

	var response WsJsonResponse
	response.Message = `<em><small>Connected to Server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

func ListenForWs(conn *WebSocketConnection) {
	//If listener crashes restart it
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			wsChn <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		// forward the payload stored in `wsChn` to the broadcast function
		e := <- wsChn

		// Do different things based on the action that triggered you
		switch e.Action {
		// If action is `username` send it to `getUserList` and broadcast the return
		case "username":
			// get a list of all users and send it to the broadcast function
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)
		// If action is `left` delete user from list that send the message
		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)
		// If action is `broadcast` receive message and broadcast to all connected users
		case "broadcast":
			// Broadcast sends the username and a message
			response.Action = "broadcast"
			// Prepend Username in front of message
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			// And send to everyone
			broadcastToAll(response)

		}
	}
}


// Collect all connected user's names and return a sorted list
func getUserList() []string {

	var userList []string

	for _, x := range clients {
		if x != "" {
			// If user name is not empty string append it
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList
}

func broadcastToAll(response WsJsonResponse) {
	// Broadcast payload to all connected clients
	for client := range clients {
		err := client.WriteJSON(response)
		// If you encounter an error delete the client
		if err != nil {
			log.Println("Websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}


// Render the jet template
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}