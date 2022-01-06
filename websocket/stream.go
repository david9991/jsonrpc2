// Package websocket provides WebSocket transport support for JSON-RPC
// 2.0.
package websocket

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
)

// A ObjectStream is a jsonrpc2.ObjectStream that uses a WebSocket to
// send and receive JSON-RPC 2.0 objects.
type ObjectStream struct {
	conn *ws.Conn
}

// NewObjectStream creates a new jsonrpc2.ObjectStream for sending and
// receiving JSON-RPC 2.0 objects over a WebSocket.
func NewObjectStream(conn *ws.Conn) ObjectStream {
	return ObjectStream{conn: conn}
}

// WriteObject implements jsonrpc2.ObjectStream.
func (t ObjectStream) WriteObject(obj interface{}) error {
	msg, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	t.conn.WriteMessage(websocket.TextMessage, append([]byte("Content-Length: "), append([]byte(strconv.Itoa(len(msg))), []byte("\r\n\r\n")...)...))
	return t.conn.WriteMessage(websocket.TextMessage, msg)
}

// ReadObject implements jsonrpc2.ObjectStream.
func (t ObjectStream) ReadObject(v interface{}) error {
	_, cl, err := t.conn.ReadMessage()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(cl), "Content-Length:") {
		return errors.New("invalid state")
	}
	err = t.conn.ReadJSON(v)
	if e, ok := err.(*ws.CloseError); ok {
		if e.Code == ws.CloseAbnormalClosure && e.Text == io.ErrUnexpectedEOF.Error() {
			// Suppress a noisy (but harmless) log message by
			// unwrapping this error.
			err = io.ErrUnexpectedEOF
		}
	}
	return err
}

// Close implements jsonrpc2.ObjectStream.
func (t ObjectStream) Close() error {
	return t.conn.Close()
}
