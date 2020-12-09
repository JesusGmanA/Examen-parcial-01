package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"

	"./useful"
)

const SHOW_ALL_MSGS = 1
const SAVE_ALL_MSGS = 2
const TERMINATE_SRVR = 3

type Server struct {
	Clients map[useful.Client]string
	Chat    string //Chat for the server
}

func (s *Server) servidor() {
	server, err := net.Listen("tcp", useful.PORT)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for {
		client, err := server.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		go s.handleClient(client)
	}
}

func (s *Server) handleClient(client net.Conn) {
	var r useful.Request
	continueWithConnection := true
	for continueWithConnection {
		err := gob.NewDecoder(client).Decode(&r)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			nick := r.NickName
			fmt.Println("User: " + nick + " sent a request")
			if s.clientIdExists(client, nick) {
				switch r.RequestID {
				case useful.MSG_REQUEST:
					go s.sendMessageToAllUsers(nick, r)
				case useful.FILE_REQUEST:
					go s.sendFileToAllUsers(nick, r)
				case useful.GET_ALL_MSGS:
					go s.sendHistoricMsgs(nick)
				case useful.DISCONNECT:
					s.removeClientId(client, nick)
					continueWithConnection = false
					fmt.Println("User: " + nick)
					fmt.Println("User: " + nick + " went offline")
				}
			} else {
				s.addNewClient(client, r.NickName)
				fmt.Println("User: " + r.NickName + " joined the server")
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func (s *Server) sendFileToAllUsers(nick string, r useful.Request) {
	var err error
	var name string
	msg := "[" + r.MsgTimestamp + "] - " + nick + " send a file: " + r.FileInfo.FullPath + "\n"
	s.Chat += msg //This is for the printing on the server
	for client, _ := range s.Clients {
		name = client.NickName
		if name == nick { //We don't want to send the info to the user that send it
			s.Clients[client] += "[" + r.MsgTimestamp + "] - You sent a file: " + r.FileInfo.FullPath + "\n" //We save the historic message for the user that send the request differently
			continue
		} else {
			s.Clients[client] += msg
			endUsrReq := useful.Request{NickName: client.NickName, RequestMsg: s.Clients[client], RequestID: useful.FILE_REQUEST, FileInfo: r.FileInfo} //Send all of the chat including the last message to the user
			err = gob.NewEncoder(client.Connection).Encode(&endUsrReq)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Server) sendHistoricMsgs(nick string) {
	for client, _ := range s.Clients {
		if client.NickName == nick {
			endUsrReq := useful.Request{RequestMsg: s.Clients[client], RequestID: useful.GET_ALL_MSGS}
			err := gob.NewEncoder(client.Connection).Encode(&endUsrReq) //This sends the chat to the specified user
			if err != nil {
				fmt.Println(err)
			}
			break
		}
	}
}

func (s *Server) sendMessageToAllUsers(nick string, req useful.Request) {
	var err error
	var name string
	msg := "[" + req.MsgTimestamp + "] - " + nick + ": " + req.RequestMsg + "\n"
	s.Chat += msg //This is for the printing on the server
	for client, _ := range s.Clients {
		name = client.NickName
		if name == nick { //We don't want to send the info to the user that send it
			s.Clients[client] += "[" + req.MsgTimestamp + "] - You: " + req.RequestMsg + "\n" //We save the historic message for the user that send the request differently
			continue
		} else {
			s.Clients[client] += msg
			endUsrReq := useful.Request{RequestMsg: s.Clients[client], RequestID: useful.MSG_REQUEST} //Send all of the chat including the last message to the user
			err = gob.NewEncoder(client.Connection).Encode(&endUsrReq)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (s *Server) clientIdExists(client net.Conn, nick string) bool { //To fix
	aux := useful.Client{NickName: nick, Connection: client}
	_, found := s.Clients[aux]
	if found {
		return true
	}
	return false
}

func (s *Server) printMessagesFromClients() {
	fmt.Println(s.Chat)
}

func (s *Server) removeClientId(client net.Conn, nick string) {
	c := useful.Client{NickName: nick, Connection: client}
	delete(s.Clients, c)
}

func (s *Server) addNewClient(client net.Conn, nick string) {
	c := useful.Client{NickName: nick, Connection: client}
	s.Clients[c] = ""
}

func (s *Server) saveAllMessages() {
	file, err := os.Create("backup.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file.WriteString(s.Chat)
	file.Close()
}

func main() {
	s := Server{Clients: make(map[useful.Client]string), Chat: ""}
	option := 0
	go s.servidor()

	for option != TERMINATE_SRVR {
		option = getMenuOpt()

		switch option {
		case SHOW_ALL_MSGS:
			s.printMessagesFromClients()
		case SAVE_ALL_MSGS:
			s.saveAllMessages()
		}
		if option != TERMINATE_SRVR {
			fmt.Print("Press 'Enter' to continue...")
			fmt.Scanln()
		} else {
			s.shutdownConnections()
			fmt.Print("Shutting down the server...")
		}
	}
}

func (s *Server) shutdownConnections() {
	for client, _ := range s.Clients {
		endUsrReq := useful.Request{RequestID: useful.DISCONNECT}
		err := gob.NewEncoder(client.Connection).Encode(&endUsrReq) //This sends the chat to the specified user
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getMenuOpt() int {
	var opt int
	fmt.Println("------------------Select an option------------------")
	fmt.Println("1. Display all messages")
	fmt.Println("2. Save all messages on the server")
	fmt.Println("3. Shut down server")
	fmt.Print("Option: ")
	fmt.Scanln(&opt)
	return opt
}
