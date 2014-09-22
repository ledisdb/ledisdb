package ledis

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	kTypeDeleteEvent uint8 = 0
	kTypePutEvent    uint8 = 1
)

var (
	errInvalidPutEvent    = errors.New("invalid put event")
	errInvalidDeleteEvent = errors.New("invalid delete event")
	errInvalidEvent       = errors.New("invalid event")
)

type eventBatch struct {
	bytes.Buffer
}

func (b *eventBatch) Put(key []byte, value []byte) {
	l := uint32(len(key) + len(value) + 1 + 2)
	binary.Write(b, binary.BigEndian, l)
	b.WriteByte(kTypePutEvent)
	keyLen := uint16(len(key))
	binary.Write(b, binary.BigEndian, keyLen)
	b.Write(key)
	b.Write(value)
}

func (b *eventBatch) Delete(key []byte) {
	l := uint32(len(key) + 1)
	binary.Write(b, binary.BigEndian, l)
	b.WriteByte(kTypeDeleteEvent)
	b.Write(key)
}

type eventWriter interface {
	Put(key []byte, value []byte)
	Delete(key []byte)
}

func decodeEventBatch(w eventWriter, data []byte) error {
	for {
		if len(data) == 0 {
			return nil
		}

		if len(data) < 4 {
			return io.ErrUnexpectedEOF
		}

		l := binary.BigEndian.Uint32(data)
		data = data[4:]
		if uint32(len(data)) < l {
			return io.ErrUnexpectedEOF
		}

		if err := decodeEvent(w, data[0:l]); err != nil {
			return err
		}
		data = data[l:]
	}
}

func decodeEvent(w eventWriter, b []byte) error {
	if len(b) == 0 {
		return errInvalidEvent
	}

	switch b[0] {
	case kTypePutEvent:
		if len(b[1:]) < 2 {
			return errInvalidPutEvent
		}

		keyLen := binary.BigEndian.Uint16(b[1:3])
		b = b[3:]
		if len(b) < int(keyLen) {
			return errInvalidPutEvent
		}

		w.Put(b[0:keyLen], b[keyLen:])
	case kTypeDeleteEvent:
		w.Delete(b[1:])
	default:
		return errInvalidEvent
	}

	return nil
}

func formatEventKey(buf []byte, k []byte) ([]byte, error) {
	if len(k) < 2 {
		return nil, errInvalidEvent
	}

	buf = append(buf, fmt.Sprintf("DB:%2d ", k[0])...)
	buf = append(buf, fmt.Sprintf("%s ", TypeName[k[1]])...)

	db := new(DB)
	db.index = k[0]

	//to do format at respective place

	switch k[1] {
	case KVType:
		if key, err := db.decodeKVKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
		}
	case HashType:
		if key, field, err := db.hDecodeHashKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendQuote(buf, String(field))
		}
	case HSizeType:
		if key, err := db.hDecodeSizeKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
		}
	case ListType:
		if key, seq, err := db.lDecodeListKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendInt(buf, int64(seq), 10)
		}
	case LMetaType:
		if key, err := db.lDecodeMetaKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
		}
	case ZSetType:
		if key, m, err := db.zDecodeSetKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendQuote(buf, String(m))
		}
	case ZSizeType:
		if key, err := db.zDecodeSizeKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
		}
	case ZScoreType:
		if key, m, score, err := db.zDecodeScoreKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendQuote(buf, String(m))
			buf = append(buf, ' ')
			buf = strconv.AppendInt(buf, score, 10)
		}
	case BitType:
		if key, seq, err := db.bDecodeBinKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendUint(buf, uint64(seq), 10)
		}
	case BitMetaType:
		if key, err := db.bDecodeMetaKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
		}
	case SetType:
		if key, member, err := db.sDecodeSetKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendQuote(buf, String(member))
		}
	case SSizeType:
		if key, err := db.sDecodeSizeKey(k); err != nil {
			return nil, err
		} else {
			buf = strconv.AppendQuote(buf, String(key))
		}
	case ExpTimeType:
		if tp, key, t, err := db.expDecodeTimeKey(k); err != nil {
			return nil, err
		} else {
			buf = append(buf, TypeName[tp]...)
			buf = append(buf, ' ')
			buf = strconv.AppendQuote(buf, String(key))
			buf = append(buf, ' ')
			buf = strconv.AppendInt(buf, t, 10)
		}
	case ExpMetaType:
		if tp, key, err := db.expDecodeMetaKey(k); err != nil {
			return nil, err
		} else {
			buf = append(buf, TypeName[tp]...)
			buf = append(buf, ' ')
			buf = strconv.AppendQuote(buf, String(key))
		}
	default:
		return nil, errInvalidEvent
	}

	return buf, nil
}
