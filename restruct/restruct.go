package main

import (
	"encoding/binary"
	"io/ioutil"
	"os"

	"gopkg.in/restruct.v1"
)

type Record struct {
	Message string `struct:"[128]byte"`
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

	var w Container
	w.Version = 0xFB000000
	w.NumRecord = 1
	w.Records = ['abc', 'def']
}
