package useful

import "net"

const CONNECTION_REQUEST = uint64(0)
const MSG_REQUEST = uint64(1)
const FILE_REQUEST = uint64(2)
const GET_ALL_MSGS = uint64(3)
const DISCONNECT = uint64(4) //Last request that should be made
const DEFAULT_DATE = "1997-10-11"

type Client struct {
	NickName   string
	Connection net.Conn
}

type File struct {
	FileName  string
	DataBytes []byte
	DataSize  int64
	FullPath  string
}

type Request struct {
	NickName     string
	RequestMsg   string
	RequestID    uint64
	MsgTimestamp string
	FileInfo     File
}

const PORT = ":8043"
