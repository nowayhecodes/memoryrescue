package memoryrescue

import "io"

// Provides a byte buffer which can minimize memory allocations.
//
// Also, it can be used with functions appending data to the given []byte slice
//
type Buffer struct {

	// Byte buffer to use in apped-like workloads
	buff []byte
}

// Returns the Buffer size
func (bf *Buffer) Len() int {
	return len(bf.buff)
}

// Appends all the data from reader to buffer
func (bf *Buffer) ReadFrom(reader io.Reader) (int64, error) {
	p := bf.buff
	start := int64(len(p))
	max := int64(cap(p))
	n := start

	if max == 0 {
		max = 64
		p = make([]byte, max)
	} else {
		p = p[:max]
	}

	for {
		if n == max {
			max *= 2
			newByte := make([]byte, max)
			copy(newByte, p)
			p = newByte
		}

		nn, err := reader.Read(p[n:])
		n += int64(nn)

		if err != nil {
			bf.buff = p[:n]
			n -= start

			if err == io.EOF {
				return n, nil
			}
			return n, err
		}
	}
}

// Write to buffer
func (bf *Buffer) WriteTo(writer io.Writer) (int64, error) {
	n, err := writer.Write(bf.buff)
	return int64(n), err
}

func (bf *Buffer) Write(p []byte) (int, error) {
	bf.buff = append(bf.buff, p...)
	return len(p), nil
}

// Appends bytes to buffer
func (bf *Buffer) WriteByte(bt byte) error {
	bf.buff = append(bf.buff, bt)
	return nil
}

// Appends string to buffer
func (bf *Buffer) WriteString(str string) (int, error) {
	bf.buff = append(bf.buff, str...)
	return len(str), nil
}

// Sets buffer to bytes slice
func (bf *Buffer) Set(bytes []byte) {
	bf.buff = append(bf.buff[:0], bytes...)
}

func (bf *Buffer) SetString(str string) {
	bf.buff = append(bf.buff[:0], str...)
}

func (bf *Buffer) String() string {
	return string(bf.buff)
}

// Returns all bytes in Buffer
func (bf *Buffer) Bytes() []byte {
	return bf.buff
}

func (bf *Buffer) Reset() {
	bf.buff = bf.buff[:0]
}
