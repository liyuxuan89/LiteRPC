package codec

import (
	"encoding/gob"
	"io"
	"log"
)

type gobCodec struct {
	conn io.ReadWriteCloser
	enc *gob.Encoder
	dec *gob.Decoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	return &gobCodec{
		conn: conn,
		enc: gob.NewEncoder(conn),
		dec: gob.NewDecoder(conn),
	}
}

func (c *gobCodec) ReadHeader(header *Header) error {
	return c.dec.Decode(header)
}

func (c *gobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(header); err != nil {
		log.Println("codec error: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("codec error: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *gobCodec) Close() error {
	return c.conn.Close()
}