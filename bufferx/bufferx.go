package bufferx

import (
	"encoding/binary"
	"errors"
	"reflect"
	"sync"
)

type BufferX struct {
	buf []byte
	pos int
	err error
}

type Reader struct {
	data []byte
	pos  int
	err  error
}

func (b *BufferX) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *BufferX) WL(target interface{}) {
	if b.err != nil {
		return
	}

	value := reflect.ValueOf(target)
	if !value.IsValid() {
		b.err = errors.New("unsupported WL target")
		return
	}

	if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.Uint8 {
		b.buf = append(b.buf, value.Bytes()...)
		return
	}

	switch value.Kind() {
	case reflect.Uint8:
		b.buf = append(b.buf, uint8(value.Uint()))
	case reflect.Uint16:
		var tmp [2]byte
		binary.LittleEndian.PutUint16(tmp[:], uint16(value.Uint()))
		b.buf = append(b.buf, tmp[:]...)
	case reflect.Uint32:
		var tmp [4]byte
		binary.LittleEndian.PutUint32(tmp[:], uint32(value.Uint()))
		b.buf = append(b.buf, tmp[:]...)
	case reflect.Uint64:
		var tmp [8]byte
		binary.LittleEndian.PutUint64(tmp[:], value.Uint())
		b.buf = append(b.buf, tmp[:]...)
	default:
		b.err = binary.Write(b, binary.LittleEndian, target)
	}
}

func (b *BufferX) WB(target interface{}) {
	if b.err != nil {
		return
	}

	value := reflect.ValueOf(target)
	if !value.IsValid() {
		b.err = errors.New("unsupported WB target")
		return
	}

	if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.Uint8 {
		b.buf = append(b.buf, value.Bytes()...)
		return
	}

	switch value.Kind() {
	case reflect.Uint8:
		b.buf = append(b.buf, uint8(value.Uint()))
	case reflect.Uint16:
		var tmp [2]byte
		binary.BigEndian.PutUint16(tmp[:], uint16(value.Uint()))
		b.buf = append(b.buf, tmp[:]...)
	case reflect.Uint32:
		var tmp [4]byte
		binary.BigEndian.PutUint32(tmp[:], uint32(value.Uint()))
		b.buf = append(b.buf, tmp[:]...)
	case reflect.Uint64:
		var tmp [8]byte
		binary.BigEndian.PutUint64(tmp[:], value.Uint())
		b.buf = append(b.buf, tmp[:]...)
	default:
		b.err = binary.Write(b, binary.BigEndian, target)
	}
}

func (b *BufferX) RL(target interface{}) {
	if b.err != nil {
		return
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		b.err = errors.New("RL target must be a non-nil pointer")
		return
	}

	elem := value.Elem()
	switch elem.Kind() {
	case reflect.Uint8:
		elem.SetUint(uint64(b.readUint8()))
	case reflect.Uint16:
		elem.SetUint(uint64(b.readUint16LE()))
	case reflect.Uint32:
		elem.SetUint(uint64(b.readUint32LE()))
	case reflect.Uint64:
		elem.SetUint(b.readUint64LE())
	case reflect.Slice:
		if elem.Type().Elem().Kind() != reflect.Uint8 {
			b.err = errors.New("RL only supports []byte slices")
			return
		}
		length := elem.Len()
		slice := b.readBytes(length)
		if slice != nil {
			elem.SetBytes(slice)
		}
	default:
		b.err = binary.Read(b, binary.LittleEndian, target)
	}
}

func (b *BufferX) RB(target interface{}) {
	if b.err != nil {
		return
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		b.err = errors.New("RB target must be a non-nil pointer")
		return
	}

	elem := value.Elem()
	switch elem.Kind() {
	case reflect.Uint8:
		elem.SetUint(uint64(b.readUint8()))
	case reflect.Uint16:
		elem.SetUint(uint64(b.readUint16BE()))
	case reflect.Uint32:
		elem.SetUint(uint64(b.readUint32BE()))
	case reflect.Uint64:
		elem.SetUint(uint64(b.readUint64BE()))
	case reflect.Slice:
		if elem.Type().Elem().Kind() != reflect.Uint8 {
			b.err = errors.New("RB only supports []byte slices")
			return
		}
		length := elem.Len()
		slice := b.readBytes(length)
		if slice != nil {
			elem.SetBytes(slice)
		}
	default:
		b.err = binary.Read(b, binary.BigEndian, target)
	}
}

func (b *BufferX) Error() error {
	return b.err
}

func (b *BufferX) Reset() {
	b.buf = b.buf[:0]
	b.pos = 0
	b.err = nil
}

func (b *BufferX) Len() int {
	return len(b.buf) - b.pos
}

func (b *BufferX) Bytes() []byte {
	return b.buf
}

func (b *BufferX) Read(p []byte) (int, error) {
	if b.pos >= len(b.buf) {
		return 0, errors.New("EOF")
	}

	n := copy(p, b.buf[b.pos:])
	b.pos += n
	return n, nil
}

func (b *BufferX) readBytes(length int) []byte {
	if b.err != nil {
		return nil
	}
	if length < 0 || b.pos+length > len(b.buf) {
		b.err = errors.New("read beyond buffer bounds")
		return nil
	}

	slice := b.buf[b.pos : b.pos+length]
	b.pos += length
	return slice
}

func (b *BufferX) readUint8() uint8 {
	return uint8(b.readUint64(1))
}

func (b *BufferX) readUint16LE() uint16 {
	return uint16(b.readUint64(2))
}

func (b *BufferX) readUint32LE() uint32 {
	return uint32(b.readUint64(4))
}

func (b *BufferX) readUint64LE() uint64 {
	return b.readUint64(8)
}

func (b *BufferX) readUint16BE() uint16 {
	if b.err != nil {
		return 0
	}
	if b.pos+2 > len(b.buf) {
		b.err = errors.New("read beyond buffer bounds")
		return 0
	}
	value := binary.BigEndian.Uint16(b.buf[b.pos : b.pos+2])
	b.pos += 2
	return value
}

func (b *BufferX) readUint32BE() uint32 {
	if b.err != nil {
		return 0
	}
	if b.pos+4 > len(b.buf) {
		b.err = errors.New("read beyond buffer bounds")
		return 0
	}
	value := binary.BigEndian.Uint32(b.buf[b.pos : b.pos+4])
	b.pos += 4
	return value
}

func (b *BufferX) readUint64BE() uint64 {
	if b.err != nil {
		return 0
	}
	if b.pos+8 > len(b.buf) {
		b.err = errors.New("read beyond buffer bounds")
		return 0
	}
	value := binary.BigEndian.Uint64(b.buf[b.pos : b.pos+8])
	b.pos += 8
	return value
}

func (b *BufferX) readUint64(size int) uint64 {
	if b.err != nil {
		return 0
	}
	if b.pos+size > len(b.buf) {
		b.err = errors.New("read beyond buffer bounds")
		return 0
	}

	var value uint64
	switch size {
	case 1:
		value = uint64(b.buf[b.pos])
	case 2:
		value = uint64(binary.LittleEndian.Uint16(b.buf[b.pos : b.pos+2]))
	case 4:
		value = uint64(binary.LittleEndian.Uint32(b.buf[b.pos : b.pos+4]))
	case 8:
		value = binary.LittleEndian.Uint64(b.buf[b.pos : b.pos+8])
	default:
		b.err = errors.New("unsupported integer size")
		return 0
	}

	b.pos += size
	return value
}

func New(data []byte) *BufferX {
	if data == nil {
		return &BufferX{buf: make([]byte, 0, 128)}
	}
	return &BufferX{buf: data}
}

func NewWithCapacity(capacity int) *BufferX {
	return &BufferX{buf: make([]byte, 0, capacity)}
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

func (r *Reader) Error() error {
	return r.err
}

func (r *Reader) Len() int {
	return len(r.data) - r.pos
}

func (r *Reader) ReadBytes(length int) []byte {
	if r.err != nil {
		return nil
	}
	if length < 0 || r.pos+length > len(r.data) {
		r.err = errors.New("read beyond buffer bounds")
		return nil
	}

	slice := r.data[r.pos : r.pos+length]
	r.pos += length
	return slice
}

func (r *Reader) ReadUint8() uint8 {
	return r.readUint8()
}

func (r *Reader) ReadUint16() uint16 {
	return r.readUint16LE()
}

func (r *Reader) ReadUint32() uint32 {
	return r.readUint32LE()
}

func (r *Reader) ReadUint64() uint64 {
	return r.readUint64LE()
}

func (r *Reader) readUint8() uint8 {
	return uint8(r.readUint64(1))
}

func (r *Reader) readUint16LE() uint16 {
	return uint16(r.readUint64(2))
}

func (r *Reader) readUint32LE() uint32 {
	return uint32(r.readUint64(4))
}

func (r *Reader) readUint64LE() uint64 {
	return r.readUint64(8)
}

func (r *Reader) readUint16BE() uint16 {
	return binary.BigEndian.Uint16(r.ReadBytes(2))
}

func (r *Reader) readUint32BE() uint32 {
	return binary.BigEndian.Uint32(r.ReadBytes(4))
}

func (r *Reader) readUint64BE() uint64 {
	return binary.BigEndian.Uint64(r.ReadBytes(8))
}

func (r *Reader) readUint64(size int) uint64 {
	if r.err != nil {
		return 0
	}
	if r.pos+size > len(r.data) {
		r.err = errors.New("read beyond buffer bounds")
		return 0
	}

	var value uint64
	switch size {
	case 1:
		value = uint64(r.data[r.pos])
	case 2:
		value = uint64(binary.LittleEndian.Uint16(r.data[r.pos : r.pos+2]))
	case 4:
		value = uint64(binary.LittleEndian.Uint32(r.data[r.pos : r.pos+4]))
	case 8:
		value = binary.LittleEndian.Uint64(r.data[r.pos : r.pos+8])
	default:
		r.err = errors.New("unsupported integer size")
		return 0
	}

	r.pos += size
	return value
}

func (r *Reader) RL(target interface{}) {
	if r.err != nil {
		return
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		r.err = errors.New("RL target must be a non-nil pointer")
		return
	}

	elem := value.Elem()
	switch elem.Kind() {
	case reflect.Uint8:
		elem.SetUint(uint64(r.readUint8()))
	case reflect.Uint16:
		elem.SetUint(uint64(r.readUint16LE()))
	case reflect.Uint32:
		elem.SetUint(uint64(r.readUint32LE()))
	case reflect.Uint64:
		elem.SetUint(r.readUint64LE())
	case reflect.Slice:
		if elem.Type().Elem().Kind() != reflect.Uint8 {
			r.err = errors.New("RL only supports []byte slices")
			return
		}
		// Zero-copy: return slice directly without copying
		slice := r.ReadBytes(elem.Len())
		if slice != nil {
			elem.SetBytes(slice)
		}
	default:
		r.err = errors.New("unsupported RL target type")
	}
}

func (r *Reader) RB(target interface{}) {
	if r.err != nil {
		return
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		r.err = errors.New("RB target must be a non-nil pointer")
		return
	}

	elem := value.Elem()
	switch elem.Kind() {
	case reflect.Uint8:
		elem.SetUint(uint64(r.readUint8()))
	case reflect.Uint16:
		elem.SetUint(uint64(r.readUint16BE()))
	case reflect.Uint32:
		elem.SetUint(uint64(r.readUint32BE()))
	case reflect.Uint64:
		elem.SetUint(uint64(r.readUint64BE()))
	case reflect.Slice:
		if elem.Type().Elem().Kind() != reflect.Uint8 {
			r.err = errors.New("RB only supports []byte slices")
			return
		}
		// Zero-copy: return slice directly without copying
		slice := r.ReadBytes(elem.Len())
		if slice != nil {
			elem.SetBytes(slice)
		}
	default:
		r.err = errors.New("unsupported RB target type")
	}
}

// Buffer pool for reuse
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &BufferX{buf: make([]byte, 0, 128)}
	},
}

func Get() *BufferX {
	buf := bufferPool.Get().(*BufferX)
	buf.Reset()
	return buf
}

func Put(buf *BufferX) {
	buf.Reset()
	bufferPool.Put(buf)
}
