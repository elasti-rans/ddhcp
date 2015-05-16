package ddhcp

import (
	"errors"
	"fmt"
)

type MsgType byte

const (
	Discover MsgType = 1 // Broadcast Packet From Client - Can I have an IP?
	Offer    MsgType = 2 // Broadcast From Server - Here's an IP
	Request  MsgType = 3 // Broadcast From Client - I'll take that IP (Also start for renewals)
	Decline  MsgType = 4 // Broadcast From Client - Sorry I can't use that IP
	ACK      MsgType = 5 // From Server, Yes you can have that IP
	NAK      MsgType = 6 // From Server, No you cannot have that IP
	Release  MsgType = 7 // From Client, I don't need that IP anymore
	Inform   MsgType = 8 // From Client, I have this IP and there's nothing you can do about it
)

func NewMsgType(data byte) (MsgType, error) {
	msgType := MsgType(data)
	if msgType < Discover || msgType > Inform {
		return msgType, errors.New(fmt.Sprintf("invalid value for MsgType %d", msgType))
	}
	return msgType, nil
}

type (
	OpCode byte
)

const (
	BOOTREQUEST OpCode = 1
	BOOTREPLY   OpCode = 2
)
