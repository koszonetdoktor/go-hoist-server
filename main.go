package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"strconv"

	"github.com/gorilla/websocket"
	zmq "github.com/pebbe/zmq4"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan int)               // broadcast channel
var upgrader = websocket.Upgrader{}

var bytesToSend []uint32

func main() {
	go setupZMQ()
	http.HandleFunc("/", bar)
	http.HandleFunc("/ws", handleConnections)
	http.ListenAndServe(":8080", nil)
	// go publishProcess()
}
func bar(res http.ResponseWriter, rq *http.Request) {
	listOfHoists := make([]int, 22)

	for i := 1; i < 22; i++ {
		listOfHoists[i-1] = i
	}

	tpl, err := template.ParseFiles("index.gohtml")
	if err != nil {
		log.Fatalln(err)
	}
	tpl.ExecuteTemplate(res, "index.gohtml", listOfHoists)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	fmt.Println("hande the connections")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	i := 1
	for {
		i++

		w, err := ws.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		// bs := []byte(strconv.Itoa(i))
		// fmt.Println("bs: ", bytesToSend)
		for _, b := range bytesToSend {
			toStringSend := strconv.Itoa(int(b)) + " "
			toSend := []byte(toStringSend)
			w.Write(toSend)
			if i == 21 {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func setupZMQ() {
	subscriber, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		fmt.Println("Cant make new cokert in setup")
	}
	subscriber.Connect("tcp://192.168.250.55:5556")

	subscriber.SetSubscribe("")

	for {
		msg, err := subscriber.RecvBytes(0)
		if err != nil {
			fmt.Println("ERROR in recv", err)
			break
		}
		var bytes []uint32
		for i := 0; i < len(msg); i += 4 {
			cutLength := msg[i:]
			converted := binary.BigEndian.Uint32(cutLength)
			bytes = append(bytes, converted)
		}
		bytesToSend = bytes
		// fmt.Println(bytesToSend)
	}
}
