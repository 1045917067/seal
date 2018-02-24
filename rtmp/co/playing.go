package co

import (
	"fmt"
	"log"
	"seal/conf"
	"seal/rtmp/pt"
)

func (rc *RtmpConn) playing(p *pt.PlayPacket) (err error) {

	userSpecifiedDurationToStop := p.Duration > 0
	var startTime int64 = -1

	for {
		//read from client. use short time out.
		//if recv failed, it's ok, not an error.
		if true {
			const timeOutUs = 10 * 1000 //ms
			rc.tcpConn.SetRecvTimeout(timeOutUs)

			var msg pt.Message
			if localErr := rc.RecvMsg(&msg.Header, &msg.Payload); localErr != nil {
				// do nothing, it's ok
			}
			if len(msg.Payload.Payload) > 0 {
				//has recved play control.
				err = rc.handlePlayData(&msg)
				if err != nil {
					log.Println("playing... handle play data faield.err=", err)
					return
				}
			}
		}

		// reset the socket send and recv timeout
		rc.tcpConn.SetRecvTimeout(conf.GlobalConfInfo.Rtmp.TimeOut * 1000 * 1000)
		rc.tcpConn.SetSendTimeout(conf.GlobalConfInfo.Rtmp.TimeOut * 1000 * 1000)

		msg := rc.consumer.dump()
		if nil == msg {
			// wait and try again.
			continue
		} else {
			//send to remote
			// only when user specifies the duration,
			// we start to collect the durations for each message.
			if userSpecifiedDurationToStop {
				if startTime < 0 || startTime > int64(msg.Header.Timestamp) {
					startTime = int64(msg.Header.Timestamp)
				}

				rc.consumer.duration += (float64(msg.Header.Timestamp) - float64(startTime))
				startTime = int64(msg.Header.Timestamp)
			}

			if err = rc.SendMsg(msg, conf.GlobalConfInfo.Rtmp.TimeOut*1000000); err != nil {
				log.Println("playing... send to remote failed.err=", err)
				break
			}
		}
	}

	return
}

func (rc *RtmpConn) handlePlayData(msg *pt.Message) (err error) {

	if nil == msg {
		return
	}

	if msg.Header.IsAmf0Data() || msg.Header.IsAmf3Data() {
		log.Println("play data: has not handled play data amf data. ")
		//todo.
	} else {
		//process user control
		rc.handlePlayUserControl(msg)
	}

	return
}

func (rc *RtmpConn) handlePlayUserControl(msg *pt.Message) (err error) {

	if nil == msg {
		return
	}

	if msg.Header.IsAmf0Command() || msg.Header.IsAmf3Command() {
		//ignore
		log.Println("ignore amf cmd when handle play user control.")
		return
	}

	// for jwplayer/flowplayer, which send close as pause message.
	if true {
		p := pt.CloseStreamPacket{}
		if localError := p.Decode(msg.Payload.Payload); localError != nil {
			//it's ok
		} else {
			err = fmt.Errorf("player ask to close stream,remote=%s", rc.tcpConn.RemoteAddr())
			return
		}
	}

	// call msg,
	// support response null first
	if true {
		p := pt.CallPacket{}
		if localErr := p.Decode(msg.Payload.Payload); localErr != nil {
			// it's ok
		} else {
			pRes := pt.CallResPacket{}
			pRes.CommandName = pt.RTMP_AMF0_COMMAND_RESULT
			pRes.TransactionId = p.TransactionId
			pRes.CommandObjectMarker = pt.RTMP_AMF0_Null
			pRes.ResponseMarker = pt.RTMP_AMF0_Null

			if err = rc.SendPacket(&pRes, 0, conf.GlobalConfInfo.Rtmp.TimeOut*1000000); err != nil {
				return
			}
		}
	}

	//pause or other msg
	if true {
		p := pt.PausePacket{}
		if localErr := p.Decode(msg.Payload.Payload); localErr != nil {
			// it's ok
		} else {
			if err = rc.onPlayClientPause(uint32(rc.defaultStreamID), p.IsPause); err != nil {
				log.Println("play client pause error.err=", err)
				return
			}

			if err = rc.consumer.onPlayPause(p.IsPause); err != nil {
				log.Println("consumer on play pause error.err=", err)
				return
			}
		}
	}

	return
}

func (rc *RtmpConn) onPlayClientPause(streamID uint32, isPause bool) (err error) {

	if isPause {
		// onStatus(NetStream.Pause.Notify)
		p := pt.OnStatusCallPacket{}
		p.CommandName = pt.RTMP_AMF0_COMMAND_ON_STATUS

		p.AddObj(pt.NewAmf0Object(pt.StatusLevel, pt.StatusLevelStatus, pt.RTMP_AMF0_String))
		p.AddObj(pt.NewAmf0Object(pt.StatusCode, pt.StatusCodeStreamPause, pt.RTMP_AMF0_String))
		p.AddObj(pt.NewAmf0Object(pt.StatusDescription, "Paused stream.", pt.RTMP_AMF0_String))

		if err = rc.SendPacket(&p, streamID, conf.GlobalConfInfo.Rtmp.TimeOut*1000000); err != nil {
			return
		}

		// StreamEOF
		if true {
			p := pt.UserControlPacket{}
			p.EventType = pt.SrcPCUCStreamEOF
			p.EventData = streamID

			if err = rc.SendPacket(&p, streamID, conf.GlobalConfInfo.Rtmp.TimeOut*1000000); err != nil {
				log.Println("send PCUC(StreamEOF) message failed.")
				return
			}
		}

	} else {
		// onStatus(NetStream.Unpause.Notify)
		p := pt.OnStatusCallPacket{}
		p.CommandName = pt.RTMP_AMF0_COMMAND_ON_STATUS

		p.AddObj(pt.NewAmf0Object(pt.StatusLevel, pt.StatusLevelStatus, pt.RTMP_AMF0_String))
		p.AddObj(pt.NewAmf0Object(pt.StatusCode, pt.StatusCodeStreamUnpause, pt.RTMP_AMF0_String))
		p.AddObj(pt.NewAmf0Object(pt.StatusDescription, "UnPaused stream.", pt.RTMP_AMF0_String))

		if err = rc.SendPacket(&p, streamID, conf.GlobalConfInfo.Rtmp.TimeOut*1000000); err != nil {
			return
		}

		// StreanBegin
		if true {
			p := pt.UserControlPacket{}
			p.EventType = pt.SrcPCUCStreamBegin
			p.EventData = streamID

			if err = rc.SendPacket(&p, streamID, conf.GlobalConfInfo.Rtmp.TimeOut*1000000); err != nil {
				log.Println("send PCUC(StreanBegin) message failed.")
				return
			}
		}
	}

	return
}
