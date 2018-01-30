package conn

import (
	"UtilsTools/identify_panic"
	"log"
	"seal/kernel"
	"seal/rtmp/pt"
)

type AckWindowSizeS struct {
	AckWindowSize uint32
	HasAckedSize  uint64
}

type RtmpConn struct {
	TcpConn      *kernel.TcpSock
	ChunkStreams map[uint32]*pt.ChunkStream //key:cs id
	InChunkSize  uint32                     //default 128, set by peer
	OutChunkSize uint32                     //default 128, set by config file.
	Pool         *kernel.MemPool
	AckWindow    AckWindowSizeS
	Requests     map[float64]string //key: transactin id, value:command name
	Role         uint8              //publisher or player.
	StreamName   string
	Duration     float64 //for player.used to specified the stop when exceed the duration.
}

func (rc *RtmpConn) Cycle() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	var err error

	err = rc.HandShake()
	if err != nil {
		log.Println("rtmp handshake failed.err=", err)
		return
	}
	log.Println("rtmp handshake success.")

	err = rc.Connect()
	if err != nil {
		log.Println("connect failed. err=", err)
		return
	}
	log.Println("connect success.")

	rc.Loop()

	log.Println("rtmp conn finished, remote=", rc.TcpConn.Conn.RemoteAddr())
}
