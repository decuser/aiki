package substrate

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
)

// Canvas transport protocol.
//
// Framing:
//   u32 little endian payload length
//   payload bytes
//
// Payload:
//   u8  protoVersion (currently 1)
//   u8  opcode
//   ... opcode specific fields
//
// The Aiki surface API remains string op based. This only changes transport.

const canvasProtoVersion = 1

// Opcodes.
const (
	canvasOpCmd    = 1
	canvasOpClose  = 2
	canvasOpSetBG  = 3
	canvasOpSetFG  = 4
	canvasOpBatch  = 5
	canvasOpTurtle = 6
)

// CanvasWireBatch groups multiple commands into a single frame.
// This is transport internal and transparent to Aiki level code.
type CanvasWireBatch struct{ Cmds []any }

// CanvasWireCmd carries a generic canvas command.
// Op is an ASCII op name like "dot" or "line".
type CanvasWireCmd struct {
	Op      string
	Args    []int32
	RGBA    color.RGBA
	HasRGBA bool
	Pen     float32
}

type CanvasWireClose struct{}
type CanvasWireSetBG struct{ RGBA color.RGBA }
type CanvasWireSetFG struct{ RGBA color.RGBA }
type CanvasWireTurtle struct {
	X, Y, Heading float32
	Visible       bool
	RGBA          color.RGBA
}

func CanvasWriteFrame(w io.Writer, v any) error {
	enc := canvasNewEncoder()
	payload, err := enc.encodePayload(v)
	if err != nil {
		return err
	}
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(payload)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err = w.Write(payload)
	return err
}

func CanvasReadCommand(r io.Reader) (any, error) {
	dec := canvasNewDecoder(r)
	return dec.Read()
}

// Encoder and decoder are reusable to reduce allocations.

type canvasEncoder struct {
	buf []byte
}

func canvasNewEncoder() *canvasEncoder {
	return &canvasEncoder{buf: make([]byte, 0, 1024)}
}

func (e *canvasEncoder) encodePayload(v any) ([]byte, error) {
	e.buf = e.buf[:0]
	e.buf = append(e.buf, byte(canvasProtoVersion))
	return e.encodeOneWithVersion(v)
}

func (e *canvasEncoder) encodeOneWithVersion(v any) ([]byte, error) {
	// version already written, now write opcode and fields.
	return e.encodeOneNoVersion(v)
}

func (e *canvasEncoder) encodeOneNoVersion(v any) ([]byte, error) {
	switch x := v.(type) {
	case CanvasWireClose:
		e.buf = append(e.buf, byte(canvasOpClose))
		return e.buf, nil
	case CanvasWireSetBG:
		e.buf = append(e.buf, byte(canvasOpSetBG), x.RGBA.R, x.RGBA.G, x.RGBA.B, x.RGBA.A)
		return e.buf, nil
	case CanvasWireSetFG:
		e.buf = append(e.buf, byte(canvasOpSetFG), x.RGBA.R, x.RGBA.G, x.RGBA.B, x.RGBA.A)
		return e.buf, nil
	case CanvasWireTurtle:
		e.buf = append(e.buf, byte(canvasOpTurtle))
		e.buf = appendF32LE(e.buf, x.X)
		e.buf = appendF32LE(e.buf, x.Y)
		e.buf = appendF32LE(e.buf, x.Heading)
		vis := byte(0)
		if x.Visible {
			vis = 1
		}
		e.buf = append(e.buf, vis, x.RGBA.R, x.RGBA.G, x.RGBA.B, x.RGBA.A)
		return e.buf, nil
	case CanvasWireCmd:
		e.buf = append(e.buf, byte(canvasOpCmd))
		if len(x.Op) > 65535 {
			return nil, errors.New("canvas: op too long")
		}
		e.buf = appendU16LE(e.buf, uint16(len(x.Op)))
		e.buf = append(e.buf, []byte(x.Op)...)

		if len(x.Args) > 65535 {
			return nil, errors.New("canvas: too many args")
		}
		e.buf = appendU16LE(e.buf, uint16(len(x.Args)))
		for _, a := range x.Args {
			e.buf = appendI32LE(e.buf, a)
		}
		if x.HasRGBA {
			e.buf = append(e.buf, 1, x.RGBA.R, x.RGBA.G, x.RGBA.B, x.RGBA.A)
		} else {
			e.buf = append(e.buf, 0)
		}
		e.buf = appendF32LE(e.buf, x.Pen)
		return e.buf, nil
	case CanvasWireBatch:
		// Frame payload: ver, batch opcode, u16 count, then [u32 len][subpayload]...
		e.buf = append(e.buf, byte(canvasOpBatch))
		if len(x.Cmds) > 65535 {
			return nil, errors.New("canvas: batch too large")
		}
		e.buf = appendU16LE(e.buf, uint16(len(x.Cmds)))
		for _, c := range x.Cmds {
			start := len(e.buf)
			e.buf = appendU32LE(e.buf, 0) // placeholder
			subStart := len(e.buf)
			// subpayload starts with opcode, not version.
			switch c.(type) {
			case CanvasWireBatch:
				return nil, errors.New("canvas: nested batch not allowed")
			}
			saved := e.buf
			// Temporarily encode into same buffer, without version.
			e.buf = saved
			if _, err := e.encodeOneNoVersion(c); err != nil {
				return nil, err
			}
			subLen := uint32(len(e.buf) - subStart)
			binary.LittleEndian.PutUint32(e.buf[start:start+4], subLen)
		}
		return e.buf, nil
	default:
		return nil, fmt.Errorf("canvas: unknown command type %T", v)
	}
}

type CanvasDecoder struct {
	r   *bufio.Reader
	buf []byte
}

func NewCanvasDecoder(r io.Reader) *CanvasDecoder {
	return canvasNewDecoder(r)
}

func canvasNewDecoder(r io.Reader) *CanvasDecoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReaderSize(r, 64*1024)
	}
	return &CanvasDecoder{r: br, buf: make([]byte, 0, 64*1024)}
}

func (d *CanvasDecoder) Read() (any, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(d.r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.LittleEndian.Uint32(hdr[:])
	if n == 0 {
		return nil, errors.New("canvas: empty frame")
	}
	if n > 1<<20 {
		return nil, fmt.Errorf("canvas: frame too large: %d", n)
	}
	if cap(d.buf) < int(n) {
		d.buf = make([]byte, n)
	} else {
		d.buf = d.buf[:n]
	}
	if _, err := io.ReadFull(d.r, d.buf); err != nil {
		return nil, err
	}
	cmd, err := canvasDecodePayloadFast(d.buf)
	if err != nil {
		return nil, err
	}
	if os.Getenv("AIKI_CANVAS_OBSERVER") == "1" {
		canvasObserveDecoded(cmd)
	}
	return cmd, nil
}

func canvasDecodePayloadFast(payload []byte) (any, error) {
	if len(payload) < 2 {
		return nil, errors.New("canvas: short payload")
	}
	ver := payload[0]
	if ver != canvasProtoVersion {
		return nil, fmt.Errorf("canvas: proto version %d", ver)
	}
	op := payload[1]
	i := 2
	switch op {
	case canvasOpClose:
		return CanvasWireClose{}, nil
	case canvasOpSetBG:
		if len(payload) < i+4 {
			return nil, io.ErrUnexpectedEOF
		}
		rgba := color.RGBA{payload[i], payload[i+1], payload[i+2], payload[i+3]}
		return CanvasWireSetBG{RGBA: rgba}, nil
	case canvasOpSetFG:
		if len(payload) < i+4 {
			return nil, io.ErrUnexpectedEOF
		}
		rgba := color.RGBA{payload[i], payload[i+1], payload[i+2], payload[i+3]}
		return CanvasWireSetFG{RGBA: rgba}, nil
	case canvasOpTurtle:
		// x, y, heading (3 floats = 12 bytes) + visible (1 byte) + rgba (4 bytes) = 17 bytes
		if len(payload) < i+17 {
			return nil, io.ErrUnexpectedEOF
		}
		var err error
		var x, y, h float32
		x, i, err = readF32LE(payload, i)
		if err != nil {
			return nil, err
		}
		y, i, err = readF32LE(payload, i)
		if err != nil {
			return nil, err
		}
		h, i, err = readF32LE(payload, i)
		if err != nil {
			return nil, err
		}
		vis := payload[i] == 1
		i++
		rgba := color.RGBA{payload[i], payload[i+1], payload[i+2], payload[i+3]}
		return CanvasWireTurtle{X: x, Y: y, Heading: h, Visible: vis, RGBA: rgba}, nil
	case canvasOpCmd:
		var err error
		var oplen uint16
		oplen, i, err = readU16LE(payload, i)
		if err != nil {
			return nil, err
		}
		if len(payload) < i+int(oplen) {
			return nil, io.ErrUnexpectedEOF
		}
		op := string(payload[i : i+int(oplen)])
		i += int(oplen)
		var argc uint16
		argc, i, err = readU16LE(payload, i)
		if err != nil {
			return nil, err
		}
		args := make([]int32, int(argc))
		for j := 0; j < int(argc); j++ {
			var a int32
			a, i, err = readI32LE(payload, i)
			if err != nil {
				return nil, err
			}
			args[j] = a
		}
		if len(payload) < i+1 {
			return nil, io.ErrUnexpectedEOF
		}
		has := payload[i]
		i++
		cmd := CanvasWireCmd{Op: op, Args: args}
		if has == 1 {
			if len(payload) < i+4 {
				return nil, io.ErrUnexpectedEOF
			}
			cmd.HasRGBA = true
			cmd.RGBA = color.RGBA{payload[i], payload[i+1], payload[i+2], payload[i+3]}
			i += 4
		}
		var pen float32
		pen, i, err = readF32LE(payload, i)
		if err != nil {
			return nil, err
		}
		cmd.Pen = pen
		return cmd, nil
	case canvasOpBatch:
		var err error
		var ncmd uint16
		ncmd, i, err = readU16LE(payload, i)
		if err != nil {
			return nil, err
		}
		cmds := make([]any, 0, int(ncmd))
		for k := 0; k < int(ncmd); k++ {
			var sublen uint32
			sublen, i, err = readU32LE(payload, i)
			if err != nil {
				return nil, err
			}
			if len(payload) < i+int(sublen) {
				return nil, io.ErrUnexpectedEOF
			}
			sub := payload[i : i+int(sublen)]
			i += int(sublen)
			one, err := canvasDecodeSubpayload(sub)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, one)
		}
		return CanvasWireBatch{Cmds: cmds}, nil
	default:
		return nil, fmt.Errorf("canvas: unknown opcode %d", op)
	}
}

func canvasDecodeSubpayload(sub []byte) (any, error) {
	if len(sub) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	op := sub[0]
	i := 1
	switch op {
	case canvasOpClose:
		return CanvasWireClose{}, nil
	case canvasOpSetBG:
		if len(sub) < i+4 {
			return nil, io.ErrUnexpectedEOF
		}
		rgba := color.RGBA{sub[i], sub[i+1], sub[i+2], sub[i+3]}
		return CanvasWireSetBG{RGBA: rgba}, nil
	case canvasOpSetFG:
		if len(sub) < i+4 {
			return nil, io.ErrUnexpectedEOF
		}
		rgba := color.RGBA{sub[i], sub[i+1], sub[i+2], sub[i+3]}
		return CanvasWireSetFG{RGBA: rgba}, nil
	case canvasOpTurtle:
		if len(sub) < i+17 {
			return nil, io.ErrUnexpectedEOF
		}
		var err error
		var x, y, h float32
		x, i, err = readF32LE(sub, i)
		if err != nil {
			return nil, err
		}
		y, i, err = readF32LE(sub, i)
		if err != nil {
			return nil, err
		}
		h, i, err = readF32LE(sub, i)
		if err != nil {
			return nil, err
		}
		vis := sub[i] == 1
		i++
		rgba := color.RGBA{sub[i], sub[i+1], sub[i+2], sub[i+3]}
		return CanvasWireTurtle{X: x, Y: y, Heading: h, Visible: vis, RGBA: rgba}, nil
	case canvasOpCmd:
		var err error
		var oplen uint16
		oplen, i, err = readU16LE(sub, i)
		if err != nil {
			return nil, err
		}
		if len(sub) < i+int(oplen) {
			return nil, io.ErrUnexpectedEOF
		}
		opname := string(sub[i : i+int(oplen)])
		i += int(oplen)
		var argc uint16
		argc, i, err = readU16LE(sub, i)
		if err != nil {
			return nil, err
		}
		args := make([]int32, int(argc))
		for j := 0; j < int(argc); j++ {
			var a int32
			a, i, err = readI32LE(sub, i)
			if err != nil {
				return nil, err
			}
			args[j] = a
		}
		if len(sub) < i+1 {
			return nil, io.ErrUnexpectedEOF
		}
		has := sub[i]
		i++
		cmd := CanvasWireCmd{Op: opname, Args: args}
		if has == 1 {
			if len(sub) < i+4 {
				return nil, io.ErrUnexpectedEOF
			}
			cmd.HasRGBA = true
			cmd.RGBA = color.RGBA{sub[i], sub[i+1], sub[i+2], sub[i+3]}
			i += 4
		}
		var pen float32
		pen, i, err = readF32LE(sub, i)
		if err != nil {
			return nil, err
		}
		cmd.Pen = pen
		return cmd, nil
	default:
		return nil, fmt.Errorf("canvas: unknown opcode %d", op)
	}
}

func appendU16LE(dst []byte, v uint16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	return append(dst, b[:]...)
}

func appendU32LE(dst []byte, v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return append(dst, b[:]...)
}

func appendI32LE(dst []byte, v int32) []byte {
	return appendU32LE(dst, uint32(v))
}

func appendF32LE(dst []byte, v float32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(v))
	return append(dst, b[:]...)
}

func readU16LE(p []byte, i int) (uint16, int, error) {
	if len(p) < i+2 {
		return 0, i, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint16(p[i : i+2]), i + 2, nil
}

func readU32LE(p []byte, i int) (uint32, int, error) {
	if len(p) < i+4 {
		return 0, i, io.ErrUnexpectedEOF
	}
	return binary.LittleEndian.Uint32(p[i : i+4]), i + 4, nil
}

func readI32LE(p []byte, i int) (int32, int, error) {
	u, ni, err := readU32LE(p, i)
	return int32(u), ni, err
}

func readF32LE(p []byte, i int) (float32, int, error) {
	u, ni, err := readU32LE(p, i)
	return math.Float32frombits(u), ni, err
}

func canvasObserveDecoded(cmd any) {
	b, err := json.Marshal(cmd)
	if err != nil {
		return
	}
	_, _ = os.Stderr.Write(append(b, '\n'))
}
