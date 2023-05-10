package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
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

func SubscribeEvent(e string) {
	msg := ENSMsg{msgType: Subscribe}
	copy(msg.name[:], e[:])
	bmsg := msg2bytes(&msg)

	_, _ = ensConn.Write(bmsg)
	//fmt.Println("Subscribe" + e)
}

func UnSubscribeEvent(e string) {
	msg := ENSMsg{msgType: Unsubscribe}
	copy(msg.name[:], e[:])
	bmsg := msg2bytes(&msg)

	_, _ = ensConn.Write(bmsg)
	//fmt.Println("Unsubscribe" + e)
}

func WaitForUpdate(update func(temp_ int, humid_ int, log_ string)) {
	for {
		rsp := make([]byte, ENSMsgSize)
		_, _ = ensConn.Read(rsp)

		ensMsg := bytes2msg(rsp)

		var t string
		switch ensMsg.msgType {
		case Update:
			t = "UPDATE"
		case Publish:
			t = "PUBLISH"
		}
		name := bytes2str(ensMsg.name)
		value, _ := strconv.Atoi(bytes2str(ensMsg.value))
		now := time.Now()
		log := fmt.Sprintf("%s: %s, Event: %s, Value: %d",
			now.Format("15:04:05"), t, bytes2str(ensMsg.name), value)

		switch name {
		case EventTemperature:
			update(value, 0, log)
		case EventHumid:
			update(0, value, log)
		default:
			update(0, 0, "Unknown Event"+name)
		}
	}
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

func bytes2str(b [MaxNameLen]byte) string {
	size := 0
	for i, c := range b {
		if c == 0 {
			size = i
			break
		}
	}
	return string(b[:size])
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

func bytes2msg(b []byte) *ENSMsg {
	var ensMsg ENSMsg
	rrsp := bytes.NewReader(b)
	_ = binary.Read(rrsp, binary.LittleEndian, &ensMsg.msgType)
	_ = binary.Read(rrsp, binary.LittleEndian, &ensMsg.name)
	_ = binary.Read(rrsp, binary.LittleEndian, &ensMsg.value)
	_ = binary.Read(rrsp, binary.LittleEndian, &ensMsg.addr.sinFamily)
	_ = binary.Read(rrsp, binary.LittleEndian, &ensMsg.addr.sinPort)
	_ = binary.Read(rrsp, binary.LittleEndian, &ensMsg.addr.sinAddr)
	return &ensMsg
}
