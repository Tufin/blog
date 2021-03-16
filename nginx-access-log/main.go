package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {

	if err := serve(context.Background(), ":6060"); err != nil {
		os.Exit(1)
	}
}

// serve is capable of answering to a single client at a time
func serve(ctx context.Context, address string) error {

	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		log.Errorf("failed to UDP listen on '%s' with '%v'", address, err)
		return err
	}
	defer func() {
		if err := pc.Close(); err != nil {
			log.Errorf("failed to close packet connection with '%v'", err)
		}
	}()

	errChan := make(chan error, 1)
	// maxBufferSize specifies the size of the buffers that
	// are used to temporarily hold data from the UDP packets
	// that we receive.
	buffer := make([]byte, 2048)
	go func() {
		for {
			_, _, err := pc.ReadFrom(buffer)
			if err != nil {
				errChan <- err
				return
			}

			request, err := toAccessLog(buffer)
			if err != nil {
				errChan <- err
				return
			}
			log.Info("%+v", request)
		}
	}()

	var ret error
	select {
	case <-ctx.Done():
		ret = ctx.Err()
		log.Infof("cancelled with '%v'", err)
	case ret = <-errChan:
	}

	return ret
}

type accessLog struct {
	MethodPathProtocol string `json:"request"`
	StatusCode         string `json:"status_code"`
	Connection         string `json:"connection"`
}

func toAccessLog(accessLogRequest []byte) (*accessLog, error) {

	const substr = `{"time":`
	start := strings.Index(string(accessLogRequest), substr)
	if start < 0 {
		msg := fmt.Sprintf("failed to find access-log request JSON '%s' starting with '%s'", string(accessLogRequest), substr)
		log.Error(msg)
		return nil, errors.New(msg)
	}
	var ret accessLog
	err := json.Unmarshal(bytes.Trim([]byte(string(accessLogRequest)[start:]), "\x00"), &ret)
	if err != nil {
		log.Errorf("failed to unmarshal access-log '%s' with '%v'", string(accessLogRequest)[start:], err)
		return nil, err
	}

	return &ret, nil
}
