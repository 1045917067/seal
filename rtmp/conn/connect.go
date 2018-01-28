package conn

import (
	"UtilsTools/identify_panic"
	"log"
)

func (rc *RtmpConn) Connect() (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err, ",panic at ", identify_panic.IdentifyPanic())
		}
	}()

	return
}
