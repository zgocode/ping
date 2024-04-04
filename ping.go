package ping

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"
)

type rTd = time.Duration

type rPing struct {
	addr    string
	conn    net.Conn
	data    []byte
	timeout time.Duration
}

type rICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func CheckSum(data []byte) uint16 {
	var sum uint32
	var length = len(data)
	var index int

	for length > 1 { // 溢出部分直接去除
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length == 1 {
		sum += uint32(data[index])
	}
	sum = uint32(uint16(sum>>16) + uint16(sum))
	sum = uint32(uint16(sum>>16) + uint16(sum))
	return uint16(^sum)
}

func New(addr string, size uint, timeout uint) *rPing {
	mt := time.Duration(timeout) * time.Millisecond
	ping := &rPing{addr: addr, timeout: mt}
	ping.Size(size)
	return ping
}

func (ping *rPing) Size(size uint) {
	var icmp = rICMP{
		Type:        8,
		Code:        0,
		Checksum:    0,
		Identifier:  1,
		SequenceNum: 1,
	}
	icmp.SequenceNum = uint16(rand.Intn(math.MaxUint16))
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, icmp)
	data := make([]byte, size)
	buffer.Write(data)
	ping.data = buffer.Bytes()
	ping.data[2] = byte(0)
	ping.data[3] = byte(0)
	ping.data[6] = byte(icmp.SequenceNum >> 8)
	ping.data[7] = byte(icmp.SequenceNum)
	icmp.Checksum = CheckSum(ping.data)
	ping.data[2] = byte(icmp.Checksum >> 8)
	ping.data[3] = byte(icmp.Checksum)
}

func (ping *rPing) Address(addr string) {
	ping.close()
	ping.addr = addr
}

func (ping *rPing) connect() error {
	if ping.conn != nil {
		return nil
	}
	conn, err := net.DialTimeout("ip:icmp", ping.addr, ping.timeout)
	if err != nil {
		return err
	}
	ping.conn = conn
	return nil
}

func (ping *rPing) close() error {
	if ping.conn != nil {
		ping.conn.Close()
		ping.conn = nil
	}
	return nil
}

func (ping *rPing) handleClosed(rc bool, err error) (rTd, error) {
	if strings.Index(err.Error(), "closed") >= 0 {
		ping.close()
		if rc {
			return ping.send(false)
		}
	}
	return 0, err
}

func (ping *rPing) send(rc bool) (rTd, error) {
	if err := ping.connect(); err != nil {
		return 0, err
	}
	t1 := time.Now()
	ping.conn.SetDeadline(t1.Add(ping.timeout))
	_, err := ping.conn.Write(ping.data)
	if err != nil {
		return ping.handleClosed(rc, err)
	}
	buf := make([]byte, 65535)
	_, err = ping.conn.Read(buf)
	if err != nil {
		return ping.handleClosed(rc, err)
	}
	return time.Since(t1), nil
}

func (ping *rPing) Get() int {
	t, err := ping.send(true)
	if err != nil {
		return -1
	}
	return int(t.Milliseconds())
}
