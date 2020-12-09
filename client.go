package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"./useful"
)

const SEND_MSG = 1
const SEND_FILE = 2
const SHOW_CHAT = 3
const EXIT = 4

var online = true

type Client struct {
	NickName   string
	Connection net.Conn
	Req        useful.Request
}

func (c *Client) client() {
	r := useful.Request{
		NickName:     c.NickName,
		RequestMsg:   "",
		RequestID:    useful.CONNECTION_REQUEST,
		MsgTimestamp: useful.DEFAULT_DATE}
	err := gob.NewEncoder(c.Connection).Encode(r) //Initial request sends the nickname only
	if err != nil {
		fmt.Println(err)
		return
	}
	for online {
		var req useful.Request
		err = gob.NewDecoder(c.Connection).Decode(&req)
		if err != nil {
			fmt.Println(err)
			return
		}
		switch req.RequestID {
		case useful.MSG_REQUEST:
			getMessage(req)
		case useful.FILE_REQUEST:
			getFile(req)
		case useful.GET_ALL_MSGS:
			printMsgs(req)
		case useful.DISCONNECT:
			fmt.Println("Server is offline!")
			online = false
		}
	}
}

func printMsgs(req useful.Request) {
	fmt.Println(req.RequestMsg)
}

func getMessage(req useful.Request) {
	fmt.Println(req.RequestMsg)
}

func getFile(req useful.Request) {
	userName := req.NickName
	if _, err := os.Stat(userName); os.IsNotExist(err) { //We create a specific directory for the user
		err = os.Mkdir(userName, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}
	filePath := userName + "\\" + req.FileInfo.FileName
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	file.Write(req.FileInfo.DataBytes)
	file.Close()
	fmt.Println(req.RequestMsg)
}

func (c *Client) deleteClient() {
	req := useful.Request{NickName: c.NickName, RequestID: useful.DISCONNECT, MsgTimestamp: useful.DEFAULT_DATE}
	err := gob.NewEncoder(c.Connection).Encode(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Logged off")
	c.Connection.Close()
}

func main() {
	scanner := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter a username: ")
	aux, _, _ := scanner.ReadLine()
	nick := string(aux)
	client, err := net.Dial("tcp", useful.PORT) //we make a shared connection for all services
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Trying to connect to the server...")
	fmt.Println("Connected...")
	c := Client{NickName: nick, Connection: client, Req: useful.Request{RequestMsg: "", RequestID: useful.CONNECTION_REQUEST}}
	go c.client()
	continueRunning(c, scanner)
}

func continueRunning(c Client, s *bufio.Reader) {
	option := 0
	for online {
		option = getMenuOption()

		switch option {
		case SEND_MSG:
			c.sendMessage(s)
		case SEND_FILE:
			c.sendFile(s)
		case SHOW_CHAT:
			c.showChat()
		case EXIT:
			online = false
		}
		if option != EXIT || online {
			fmt.Println("Press 'Enter' to continue...")
			fmt.Scanln()
			clearScreen()
		}
	}
	fmt.Println("Logging out...")
	c.deleteClient()
}

func (c *Client) sendMessage(s *bufio.Reader) { //Working on this
	var msg string
	fmt.Print("Enter a message: ")
	aux, _, _ := s.ReadLine()
	msg = string(aux)
	req := useful.Request{NickName: c.NickName, RequestMsg: msg, RequestID: useful.MSG_REQUEST, MsgTimestamp: getCurrTime()}
	err := gob.NewEncoder(c.Connection).Encode(req) //Sending msg to the server
	fmt.Print(err)
}

func (c *Client) sendFile(s *bufio.Reader) {
	var req useful.Request
	var path string

	req.NickName = c.NickName
	req.RequestID = useful.FILE_REQUEST
	fmt.Print("Write the path to the file: ")
	aux, _, _ := s.ReadLine()
	path = string(aux)
	fmt.Println(path)
	// Leer archivo
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("The system couldn't find the file on the specified path:", err)
		return
	}
	defer file.Close()
	stats, err := file.Stat()
	if err != nil {
		fmt.Println("File information couldn't be retrieved:", err)
		return
	}
	_, fileName := filepath.Split(path)
	fmt.Println(file)
	req.FileInfo.FileName = fileName
	req.FileInfo.DataSize = stats.Size()
	req.FileInfo.DataBytes = make([]byte, stats.Size())
	req.FileInfo.FullPath = path
	file.Read(req.FileInfo.DataBytes) //we retrieve the data from the file into our byte slice

	err = gob.NewEncoder(c.Connection).Encode(req) //We send the file to our server through the request
	if err != nil {
		fmt.Println(err)
		return
	}

}

func getCurrTime() string {
	t := time.Now()
	return t.Format(time.RFC822)
}

func (c *Client) showChat() {
	var err error
	sendReq := useful.Request{NickName: c.NickName,
		RequestID: useful.GET_ALL_MSGS} //We just need the nickname for this request and ReqID all of the other fields we leave them as default

	err = gob.NewEncoder(c.Connection).Encode(sendReq) //Sending the request to retrieve the chat
	if err != nil {
		fmt.Print(err)
	}
}

func getMenuOption() int {
	var opt int
	fmt.Println("------------------Select an option------------------")
	fmt.Println("1. Send a message")
	fmt.Println("2. Send a file")
	fmt.Println("3. Display chat")
	fmt.Println("4. Log out")
	fmt.Print("Option: ")
	fmt.Scanln(&opt)
	return opt
}

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
