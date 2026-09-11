package identity

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

var errUnsupportedCBOR = errors.New("unsupported CBOR encoding")

func decodeCBOR(data []byte) (any, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("empty CBOR")
	}
	major := data[0] >> 5
	additional := data[0] & 0x1f
	length, header, err := cborLength(data, additional)
	if err != nil {
		return nil, 0, err
	}
	switch major {
	case 0:
		return int64(length), header, nil
	case 1:
		if length > math.MaxInt64 {
			return nil, 0, errUnsupportedCBOR
		}
		return -1 - int64(length), header, nil
	case 2:
		end := header + int(length)
		if end > len(data) {
			return nil, 0, errors.New("truncated CBOR bytes")
		}
		out := append([]byte(nil), data[header:end]...)
		return out, end, nil
	case 3:
		end := header + int(length)
		if end > len(data) {
			return nil, 0, errors.New("truncated CBOR text")
		}
		return string(data[header:end]), end, nil
	case 4:
		offset := header
		items := make([]any, 0, int(length))
		for i := uint64(0); i < length; i++ {
			item, n, err := decodeCBOR(data[offset:])
			if err != nil {
				return nil, 0, err
			}
			offset += n
			items = append(items, item)
		}
		return items, offset, nil
	case 5:
		offset := header
		m := make(map[any]any, int(length))
		for i := uint64(0); i < length; i++ {
			key, n, err := decodeCBOR(data[offset:])
			if err != nil {
				return nil, 0, err
			}
			offset += n
			value, n, err := decodeCBOR(data[offset:])
			if err != nil {
				return nil, 0, err
			}
			offset += n
			m[key] = value
		}
		return m, offset, nil
	case 6:
		value, n, err := decodeCBOR(data[header:])
		return value, header + n, err
	case 7:
		switch additional {
		case 20:
			return false, 1, nil
		case 21:
			return true, 1, nil
		case 22, 23:
			return nil, 1, nil
		default:
			return nil, 0, errUnsupportedCBOR
		}
	default:
		return nil, 0, errUnsupportedCBOR
	}
}

func cborLength(data []byte, additional byte) (uint64, int, error) {
	switch {
	case additional < 24:
		return uint64(additional), 1, nil
	case additional == 24:
		if len(data) < 2 {
			return 0, 0, errors.New("truncated CBOR length")
		}
		return uint64(data[1]), 2, nil
	case additional == 25:
		if len(data) < 3 {
			return 0, 0, errors.New("truncated CBOR length")
		}
		return uint64(binary.BigEndian.Uint16(data[1:3])), 3, nil
	case additional == 26:
		if len(data) < 5 {
			return 0, 0, errors.New("truncated CBOR length")
		}
		return uint64(binary.BigEndian.Uint32(data[1:5])), 5, nil
	case additional == 27:
		if len(data) < 9 {
			return 0, 0, errors.New("truncated CBOR length")
		}
		return binary.BigEndian.Uint64(data[1:9]), 9, nil
	default:
		return 0, 0, fmt.Errorf("%w: indefinite lengths", errUnsupportedCBOR)
	}
}
