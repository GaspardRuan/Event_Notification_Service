package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
)

const (
	SrvName    = "Event_NS"
	BroadCast  = "255.255.255.255"
	ToDSPort   = 8008
	FromDSPort = 8002

	MaxNameLen  = 50
	MaxValueLen = 50

	DSReplySize = 8
	ENSMsgSize  = 120

	EventTemperature = "TemperatureChange"
	EventHumid       = "HumidityChange"
)

const (
	Subscribe = iota
	Unsubscribe
	Publish
	Update
)

type DSReply struct {
	ip   [4]byte
	port uint32
}

type SockAddrIn struct {
	sinFamily uint16
	sinPort   uint16
	sinAddr   [4]byte
	sinZero   [8]byte
}

type ENSMsg struct {
	msgType uint8
	name    [MaxNameLen]byte
	value   [MaxValueLen]byte
	addr    SockAddrIn
}

var (
	ensConn   net.Conn
	LocalAddr string
)

func InitNetCfg() {
	LocalAddr = *flag.String("lip", "192.168.137.1", "本机地址，若有虚拟机或vpn，可能有多个，跟成功跑起来的服务端一致就没问题")
	flag.Parse()
}

func InitConn() (error, string) {
	ip, port, err := resolveService(SrvName)
	if err != nil {
		return err, ""
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	ensConn, err = net.Dial("tcp", addr)
	if err != nil {
		return err, ""
	}
	return nil, addr
}

func PublishEvent(name string, value string) {
	msg := ENSMsg{msgType: Publish}
	copy(msg.name[:], name[:])
	copy(msg.value[:], value[:])
	bmsg := msg2bytes(&msg)

	_, _ = ensConn.Write(bmsg)
	//fmt.Println("Publish" + e)
}

func UpdateEvent(name string, value string) {
	msg := ENSMsg{msgType: Update}
	copy(msg.name[:], name[:])
	copy(msg.value[:], value[:])
	bmsg := msg2bytes(&msg)

	_, _ = ensConn.Write(bmsg)
	//fmt.Println("Publish" + e)
}

func resolveService(srv string) (string, int, error) {
	sraddr := fmt.Sprintf("%s:%d", BroadCast, ToDSPort)
	raddr, err := net.ResolveUDPAddr("udp", sraddr)
	if err != nil {
		return "", 0, err
	}

	sladdr := fmt.Sprintf("%s:%d", LocalAddr, FromDSPort)
	laddr, err := net.ResolveUDPAddr("udp", sladdr)
	if err != nil {
		return "", 0, err
	}

	conn, _ := net.ListenUDP("udp", laddr)
	defer conn.Close()

	msg := "RESOLVE:" + srv
	if _, err = conn.WriteTo([]byte(msg), raddr); err != nil {
		return "", 0, err
	}

	rsp := make([]byte, DSReplySize)
	if _, err = conn.Read(rsp); err != nil {
		return "", 0, err
	}

	rrsp := bytes.NewReader(rsp)
	var dsRsp DSReply
	_ = binary.Read(rrsp, binary.LittleEndian, &dsRsp.ip)
	_ = binary.Read(rrsp, binary.LittleEndian, &dsRsp.port)

	ip := fmt.Sprintf("%d.%d.%d.%d", dsRsp.ip[0], dsRsp.ip[1], dsRsp.ip[2], dsRsp.ip[3])

	return ip, int(dsRsp.port), nil
}

func msg2bytes(msg *ENSMsg) []byte {
	buf := &bytes.Buffer{}
	_ = binary.Write(buf, binary.LittleEndian, msg)
	zero := byte(0)
	for buf.Len() < ENSMsgSize {
		_ = binary.Write(buf, binary.LittleEndian, zero)
	}
	return buf.Bytes()
}
