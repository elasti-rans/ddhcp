package ddhcp

import (
	"errors"
	"net"
	"time"

	"github.com/ziutek/utils/netaddr"
)

type lease struct {
	Nic      net.HardwareAddr
	Ip       net.IP
	Expiry   time.Time // When the lease expires
	Duration time.Duration
}

type LeasePool struct {
	startIp     net.IP
	poolCap     uint
	poolOffset  uint
	db          map[string][]lease // replace slice by ring buffer
	releasedIps chan net.IP        // TODO: make this the only way to get an ip, so the thread will check for realesed ip only once realesed
	//LeaseDuration time.Duration
}

func NewLeasePool(startIp net.IP, endIp net.IP) (*LeasePool, error) {
	poolCap, err := netaddr.IPDiff(endIp, startIp)
	if err != nil {
		return nil, err
	}

	var db map[string][]lease
	return &LeasePool{startIp, uint(poolCap), 0, db, make(chan net.IP, 2)}, nil
}

func (l *LeasePool) GetLease(nic net.HardwareAddr) (*lease, error) {
	if lease, exists := l.db[nic.String()]; exists {
		// TODO: need to extend lease ?
		return &lease[len(lease)-1], nil
	}

	return l.getNewLease(nic)
}

func (l *LeasePool) getNewLease(nic net.HardwareAddr) (_ *lease, err error) {
	ip, ok := <-l.releasedIps
	if !ok {
		ip, err = l.nextIp()
		if err != nil {
			return nil, err
		}
	}

	// TODO: expire should come from options
	// TODO: need to protect the db ..., to be done when it will be db of ring buffers
	const leaseDuration = 6 * time.Second
	lease := lease{nic, ip, time.Now().Add(leaseDuration), leaseDuration}
	db_entry := append(l.db[nic.String()], lease)
	l.db[nic.String()] = db_entry
	return &lease, nil
}

func (l *LeasePool) nextIp() (net.IP, error) {
	if l.poolOffset > l.poolCap {
		return nil, errors.New("pool id exusted")
	}

	ip := netaddr.IPAdd(l.startIp, int(l.poolOffset))
	// TODO: need to excelude broadvast ips
	// instead of man len, maybe use end ip ?
	l.poolOffset++
	return ip, nil
}
