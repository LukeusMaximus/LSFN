package environment

//go:generate mkdir -p ../protobuf
//go:generate protoc3 -I $GOPATH/src/github.com/LSFN/lsfn/protobuf --go_out=../protobuf $GOPATH/src/github.com/LSFN/lsfn/protobuf/shipInput.proto $GOPATH/src/github.com/LSFN/lsfn/protobuf/environmentToVessel.proto $GOPATH/src/github.com/LSFN/lsfn/protobuf/vesselToEnvironment.proto

import (
	"net"

	"github.com/golang/protobuf/proto"

	"github.com/LSFN/lsfn/vessel/protobuf"
)

const (
	MESSAGE_BUFFER_SIZE = 10
)

type conn struct {
	inbound  <-chan *protobuf.EnvironmentToVessel
	outbound chan<- *protobuf.VesselToEnvironment
}

func connectToEnvironment(environmentUDPAddress *net.UDPAddr) (*conn, error) {
	udpConn, err := net.DialUDP("udp", nil, environmentUDPAddress)
	if err != nil {
		return nil, err
	}
	inboundMessages := make(chan *protobuf.EnvironmentToVessel, MESSAGE_BUFFER_SIZE)
	outboundMessages := make(chan *protobuf.VesselToEnvironment, MESSAGE_BUFFER_SIZE)
	environmentConnection := &conn{
		inbound:  inboundMessages,
		outbound: outboundMessages,
	}
	go readFromServer(udpConn, inboundMessages)
	go writeToServer(udpConn, outboundMessages)
	return environmentConnection, nil
}

func readFromServer(udpConn *net.UDPConn, inboundMessages chan<- *protobuf.EnvironmentToVessel) {
	var readBuf []byte
	for {
		_, _, err := udpConn.ReadFromUDP(readBuf)
		if err != nil {
			break
		}
		msg := &protobuf.EnvironmentToVessel{}
		err = proto.Unmarshal(readBuf, msg)
		if err != nil {
			break
		}
		inboundMessages <- msg
	}
}

func writeToServer(conn *net.UDPConn, outboundMessages <-chan *protobuf.VesselToEnvironment) {
	for msg := range outboundMessages {
		buf, err := proto.Marshal(msg)
		if err != nil {
			break
		}
		_, err = conn.Write(buf)
		if err != nil {
			break
		}
	}
}
