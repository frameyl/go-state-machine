package main

import (
	"fmt"
	"encoding/hex"
	"encoding/binary"
	"gopkg.in/restruct.v1"
)

type Record struct {
	Message string `struct:"[8]byte"`
}

type Container struct {
	Version   int `struct:"int32"`
	NumRecord int `struct:"int32,sizeof=Records"`
	Records   []Record
}

func main() {
	// var c Container

	// file, _ := os.Open("records")
	// defer file.Close()
	// data, _ := ioutil.ReadAll(file)

	// restruct.Unpack(data, binary.LittleEndian, &c)

	var ws Container
	ws.Version = 0xFB000000
	ws.NumRecord = 1
	ws.Records = []Record{Record{"abcde"}, Record{"abdcd"}}

	var output []byte

	output, _ = restruct.Pack(binary.LittleEndian, &ws)
	fmt.Printf("%s", hex.Dump(output))

	var rs Container
	restruct.Unpack(output, binary.LittleEndian, &rs)
	fmt.Printf("Version %x, #Rec %d, Rec#1 %s, Rec#2 %s", rs.Version, rs.NumRecord, rs.Records[0].Message, rs.Records[1].Message)

	fmt.Printf("Invalid index %s", rs.Records[2].Message)
}
