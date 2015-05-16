package ddhcp

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
)

type server struct {
	rxAddr net.IP
	conn   net.PacketConn
	leases *LeasePool
}

func New(rxAddr net.IP, leasesPool *LeasePool) (*server, error) {
	// TODO: initialization need to be fixed, to init all attrs
	l, err := net.ListenPacket("udp4", fmt.Sprintf("%s:69", rxAddr.String()))
	if err != nil {
		return nil, err
	}

	return &server{rxAddr, l, leasesPool}, nil
}

func (s *server) Close() {
	s.conn.Close()
}

func Serve(s *server) {
	buffer := make([]byte, 1500)
	for {
		n, addr, err := s.conn.ReadFrom(buffer)
		if err != nil {
			log.Fatal("s.conn.ReadFrom failed with", err)
		}

		m, err := NewMsgFromData(buffer[:n])
		if err != nil {
			log.Println(err)
			continue
		}

		options, err := m.Options()
		if err != nil {
			log.Println(err)
			continue
		}

		replyMsg, err := s.ServeDhcp(m, options) // TODO: should options passed? is exists in msg, maybe not pass req
		if err != nil {
			log.Println(err)
			continue
		}
		if replyMsg != nil {
			// If IP not available, broadcast
			ipStr, portStr, err := net.SplitHostPort(addr.String())
			if err != nil {
				log.Println(err)
				continue
			}

			// TODO: the parse os str is stupied as just a memnt ago it was turned to str
			// TODO: fix the flags / is brodacst api
			if net.ParseIP(ipStr).Equal(net.IPv4zero) || m.Flags()[0] > 127 {
				port, _ := strconv.Atoi(portStr)
				addr = &net.UDPAddr{IP: net.IPv4bcast, Port: port}
			}

			if _, err := s.conn.WriteTo(replyMsg, addr); err != nil {
				log.Println(err)
				continue
			}

			if replyOpts, err := replyMsg.Options(); err != nil {
				log.Println(err)
				continue
			} else {
				log.Println("reply %v was sent to %v", replyOpts[OptionDHCPMessageType], addr)
			}
		}
	}
}

func (s *server) ServeDhcp(req Msg, options Options) (Msg, error) {
	msgType, err := options.MsgType()
	if err != nil {
		return nil, err
	}
	switch msgType {
	case Discover:
		nic, err := req.Chaddr()
		if err != nil {
			return nil, err
		}
		lease, err := s.leases.GetLease(nic)
		if err != nil {
			return nil, err
		}

		options.SetMsgType(Offer)
		options[OptionServerIdentifier] = s.rxAddr
		options.SetLeaseDuration(lease.Duration)

		return NewReplyMsg(req, lease, options)

	case Request:
		return nil, errors.New("not implemented")
	case Release, Decline:
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New(fmt.Sprintf("invalid msgType for dhcp server to get %d", msgType))
	}
}
