package bencode

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// Bencode types
type BValue interface{}

type BString string
type BInt int
type BList []BValue
type BDict map[string]BValue

type Serializer interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

type BencodeSerializer struct {
	data []byte
	pos  int
}

func NewBencodeSerializer() *BencodeSerializer {
	return &BencodeSerializer{}
}

// Marshal function
func (s *BencodeSerializer) Marshal(v interface{}) ([]byte, error) {
	bValue, err := marshal(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	return s.encode(bValue)
}

// Unmarshal function
func (s *BencodeSerializer) Unmarshal(data []byte, v interface{}) error {
	s.data = data
	s.pos = 0
	bValue, err := s.decode()
	if err != nil {
		return err
	}
	return unmarshal(bValue, reflect.ValueOf(v).Elem())
}

// Helper function to marshal Go value to BValue
func marshal(val reflect.Value) (BValue, error) {
	switch val.Kind() {
	case reflect.String:
		return BString(val.String()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return BInt(val.Int()), nil
	case reflect.Array:
		var list BList
		for i := 0; i < val.Len(); i++ {
			elem, err := marshal(val.Index(i))
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		return list, nil
	case reflect.Slice:
		switch val.Type().String() {
		case "[]uint8":
			// special case as byte-string
			return BString(val.String()), nil
		default:
			var list BList
			for i := 0; i < val.Len(); i++ {
				elem, err := marshal(val.Index(i))
				if err != nil {
					return nil, err
				}
				list = append(list, elem)
			}
			return list, nil
		}

	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return nil, errors.New("map keys must be strings")
		}
		dict := make(BDict)
		for _, key := range val.MapKeys() {
			elem, err := marshal(val.MapIndex(key))
			if err != nil {
				return nil, err
			}
			dict[key.String()] = elem
		}
		return dict, nil
	case reflect.Struct:
		dict := make(BDict)
		t := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("bencode")
			if tag == "" {
				tag = field.Name
			}
			elem, err := marshal(val.Field(i))
			if err != nil {
				return nil, err
			}
			dict[tag] = elem
		}
		return dict, nil
	default:
		return nil, errors.New("unsupported type")
	}
}

// Helper function to encode BValue to []byte
func (s *BencodeSerializer) encode(bValue BValue) ([]byte, error) {
	switch v := bValue.(type) {
	case BString:
		return []byte(fmt.Sprintf("%d:%s", len(v), string(v))), nil
	case BInt:
		return []byte(fmt.Sprintf("i%de", v)), nil
	case BList:
		var buf bytes.Buffer
		buf.WriteByte('l')
		for _, elem := range v {
			encoded, err := s.encode(elem)
			if err != nil {
				return nil, err
			}
			buf.Write(encoded)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	case BDict:
		var buf bytes.Buffer
		buf.WriteByte('d')
		for key, value := range v {
			encodedKey, err := s.encode(BString(key))
			if err != nil {
				return nil, err
			}
			buf.Write(encodedKey)
			encodedValue, err := s.encode(value)
			if err != nil {
				return nil, err
			}
			buf.Write(encodedValue)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	default:
		return nil, errors.New("unsupported BValue type")
	}
}

// Helper function to decode []byte to BValue
func (s *BencodeSerializer) decode() (BValue, error) {
	if s.pos >= len(s.data) {
		return nil, errors.New("no more data")
	}

	switch s.data[s.pos] {
	case 'i':
		s.pos++
		end := bytes.IndexByte(s.data[s.pos:], 'e')
		if end == -1 {
			return nil, errors.New("invalid integer encoding")
		}
		end += s.pos
		val, err := strconv.ParseInt(string(s.data[s.pos:end]), 10, 64)
		if err != nil {
			return nil, err
		}
		s.pos = end + 1
		return BInt(val), nil
	case 'l':
		s.pos++
		var list BList
		for s.pos < len(s.data) && s.data[s.pos] != 'e' {
			elem, err := s.decode()
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}
		if s.pos >= len(s.data) {
			return nil, errors.New("list not terminated with 'e'")
		}
		s.pos++
		return list, nil
	case 'd':
		s.pos++
		dict := make(BDict)
		for s.pos < len(s.data) && s.data[s.pos] != 'e' {
			key, err := s.decode()
			if err != nil {
				return nil, err
			}
			if keyStr, ok := key.(BString); ok {
				value, err := s.decode()
				if err != nil {
					return nil, err
				}
				dict[string(keyStr)] = value
			} else {
				return nil, errors.New("invalid dictionary key")
			}
		}
		if s.pos >= len(s.data) {
			return nil, errors.New("dictionary not terminated with 'e'")
		}
		s.pos++
		return dict, nil
	default:
		colon := bytes.IndexByte(s.data[s.pos:], ':')
		if colon == -1 {
			return nil, errors.New("invalid string encoding")
		}
		colon += s.pos
		length, err := strconv.Atoi(string(s.data[s.pos:colon]))
		if err != nil {
			return nil, err
		}
		start := colon + 1
		end := start + length
		if end > len(s.data) {
			return nil, errors.New("string length exceeds data length")
		}
		s.pos = end
		return BString(s.data[start:end]), nil
	}
}

// Helper function to unmarshal BValue to Go value
func unmarshal(bValue BValue, val reflect.Value) error {
	switch v := bValue.(type) {
	case BString:
		if val.Kind() == reflect.String {
			val.SetString(string(v))
			return nil
		}
	case BInt:
		if val.Kind() >= reflect.Int && val.Kind() <= reflect.Int64 {
			val.SetInt(int64(v))
			return nil
		}
	case BList:
		if val.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(val.Type(), len(v), len(v))
			for i, elem := range v {
				if err := unmarshal(elem, slice.Index(i)); err != nil {
					return err
				}
			}
			val.Set(slice)
			return nil
		}
	case BDict:
		if val.Kind() == reflect.Map {
			if val.Type().Key().Kind() != reflect.String {
				return errors.New("map keys must be strings")
			}
			mapType := reflect.MakeMap(val.Type())
			for key, elem := range v {
				mapKey := reflect.ValueOf(key)
				mapVal := reflect.New(val.Type().Elem()).Elem()
				if err := unmarshal(elem, mapVal); err != nil {
					return err
				}
				mapType.SetMapIndex(mapKey, mapVal)
			}
			val.Set(mapType)
			return nil
		} else if val.Kind() == reflect.Struct {
			for key, elem := range v {
				field := val.FieldByNameFunc(func(fieldName string) bool {
					field, _ := val.Type().FieldByName(fieldName)
					tag := field.Tag.Get("bencode")
					if tag == "" {
						tag = fieldName
					}
					return tag == key
				})
				if field.IsValid() && field.CanSet() {

					fieldType := field.Type()

					if fieldType.Kind() == reflect.Ptr {

						field.Set(reflect.New(fieldType.Elem()))

						field = field.Elem()

					}
					if err := unmarshal(elem, field); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
	return errors.New("type mismatch or unsupported type")
}
