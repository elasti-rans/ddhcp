package ddhcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type msgOffset uint

const (
	ofstOp           msgOffset = 0
	ofstHtype        msgOffset = 1
	ofstHlen         msgOffset = 2
	ofstHops         msgOffset = 3
	ofstXidStart     msgOffset = 4
	ofstSecsStart    msgOffset = 8
	ofstFlagsStart   msgOffset = 10
	ofstCiaddrStart  msgOffset = 12
	ofstYiaddrStart  msgOffset = 16
	ofstSiaddrStart  msgOffset = 20
	ofstGiaddrStart  msgOffset = 24
	ofstChaddrStart  msgOffset = 28
	ofstSnameStart   msgOffset = 44
	ofstFileStart    msgOffset = 108
	ofstCookieStart  msgOffset = 236
	ofstOptionsStart msgOffset = 240
	/// options has unknown length
)

type Msg []byte

func NewMsgFromData(buffer []byte) (Msg, error) {
	if len(buffer) < 240 {
		return nil, errors.New(fmt.Sprintf("buffer too small to be dhcp msg %d", len(buffer)))
	}

	m := Msg(buffer)
	if _, err := m.HLen(); err != nil {
		return nil, err
	}

	// TODO: add validation here insteadin getters geeters?
	return m, nil
}

func NewMsg(opCode OpCode, options Options) (Msg, error) {
	opts, err := options.Bytes()
	if err != nil {
		return nil, err
	}
	msg := make(Msg, 240+len(opts))

	msg[ofstOp] = byte(opCode)
	msg[ofstHtype] = byte(1)                                             // Ethernet
	copy(msg[ofstCookieStart:ofstOptionsStart], []byte{99, 130, 83, 99}) // TODO: ????
	copy(msg[ofstOptionsStart:], opts)

	return msg, nil
}

func NewReplyMsg(req Msg, lease *lease, options Options) (Msg, error) {
	m, err := NewMsg(BOOTREPLY, options)
	if err != nil {
		return nil, err
	}

	if hlen, err := req.HLen(); err != nil {
		return nil, err
	} else {
		m[ofstHlen] = hlen
	}

	// TODO: hops is missing
	if xid, err := req.Xid(); err != nil {
		return nil, err
	} else {
		binary.PutUvarint(m[ofstXidStart:ofstSecsStart], uint64(xid))
	}

	copy(m[ofstSecsStart:ofstFlagsStart], req.Secs())
	copy(m[ofstFlagsStart:ofstCiaddrStart], req.Flags())
	copy(m[ofstYiaddrStart:ofstSiaddrStart], lease.Ip)
	// TODOL siaddr is missins
	copy(m[ofstGiaddrStart:ofstChaddrStart], req.Giaddr())
	if hwaddr, err := req.Chaddr(); err != nil {
		return nil, err
	} else {
		copy(m[ofstGiaddrStart:ofstChaddrStart], hwaddr)
	}

	// TODO: sname, file are missings

	return m, nil
}

func (m Msg) OpCode() OpCode { return OpCode(m[ofstOp]) }
func (m Msg) HType() byte    { return m[ofstHtype] }
func (m Msg) HLen() (byte, error) {
	hlen := m[ofstHlen]
	if hlen > 16 {
		return hlen, errors.New(fmt.Sprintf("invalid hlen size %d", hlen))
	}
	return hlen, nil
}
func (m Msg) Hops() byte { return m[ofstHops] }
func (m Msg) Xid() (uint, error) {
	buf := m[ofstXidStart:ofstSecsStart]
	xid, err := binary.ReadUvarint(bytes.NewReader(buf))
	return uint(xid), err
}
func (m Msg) Secs() []byte   { return m[ofstSecsStart:ofstFlagsStart] }
func (m Msg) Flags() []byte  { return m[ofstFlagsStart:ofstCiaddrStart] }
func (m Msg) Ciaddr() []byte { return m[ofstCiaddrStart:ofstYiaddrStart] }
func (m Msg) Yiaddr() []byte { return m[ofstYiaddrStart:ofstSiaddrStart] }
func (m Msg) Siaddr() []byte { return m[ofstSiaddrStart:ofstGiaddrStart] }
func (m Msg) Giaddr() []byte { return m[ofstGiaddrStart:ofstChaddrStart] }
func (m Msg) Chaddr() (net.HardwareAddr, error) {
	hAddrLen, err := m.HLen()
	if err != nil {
		return nil, err
	}
	return net.HardwareAddr(m[ofstChaddrStart : ofstChaddrStart+msgOffset(hAddrLen)]), nil
}
func (m Msg) Sname() []byte  { return m[ofstSnameStart:ofstFileStart] }
func (m Msg) File() []byte   { return m[ofstFileStart:ofstCookieStart] }
func (m Msg) Cookie() []byte { return m[ofstCookieStart:ofstOptionsStart] }
func (m Msg) Options() (Options, error) {
	if msgOffset(len(m)) > ofstOptionsStart {
		return NewOptionsFromData(m[ofstOptionsStart:])
	}
	return nil, errors.New("no options")
}
