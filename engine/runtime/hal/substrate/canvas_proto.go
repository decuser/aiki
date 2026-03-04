package substrate

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
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
	canvasOpCmd   = 1
	canvasOpClose = 2
	canvasOpSetBG = 3
	canvasOpSetFG = 4
)

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

func CanvasWriteFrame(w io.Writer, v any) error {
	payload, err := canvasEncodePayload(v)
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
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.LittleEndian.Uint32(hdr[:])
	if n == 0 {
		return nil, errors.New("canvas: empty frame")
	}
	if n > 1<<20 {
		return nil, fmt.Errorf("canvas: frame too large: %d", n)
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	cmd, err := canvasDecodePayload(payload)
	if err != nil {
		return nil, err
	}
	if os.Getenv("AIKI_CANVAS_OBSERVER") == "1" {
		canvasObserveDecoded(cmd)
	}
	return cmd, nil
}

func canvasEncodePayload(v any) ([]byte, error) {
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	_ = bw.WriteByte(byte(canvasProtoVersion))

	switch x := v.(type) {
	case CanvasWireClose:
		_ = bw.WriteByte(byte(canvasOpClose))
	case CanvasWireSetBG:
		_ = bw.WriteByte(byte(canvasOpSetBG))
		_ = bw.WriteByte(x.RGBA.R)
		_ = bw.WriteByte(x.RGBA.G)
		_ = bw.WriteByte(x.RGBA.B)
		_ = bw.WriteByte(x.RGBA.A)
	case CanvasWireSetFG:
		_ = bw.WriteByte(byte(canvasOpSetFG))
		_ = bw.WriteByte(x.RGBA.R)
		_ = bw.WriteByte(x.RGBA.G)
		_ = bw.WriteByte(x.RGBA.B)
		_ = bw.WriteByte(x.RGBA.A)
	case CanvasWireCmd:
		_ = bw.WriteByte(byte(canvasOpCmd))

		if len(x.Op) > 65535 {
			return nil, errors.New("canvas: op too long")
		}
		if err := binary.Write(bw, binary.LittleEndian, uint16(len(x.Op))); err != nil {
			return nil, err
		}
		if _, err := bw.WriteString(x.Op); err != nil {
			return nil, err
		}

		if len(x.Args) > 65535 {
			return nil, errors.New("canvas: too many args")
		}
		if err := binary.Write(bw, binary.LittleEndian, uint16(len(x.Args))); err != nil {
			return nil, err
		}
		for _, a := range x.Args {
			if err := binary.Write(bw, binary.LittleEndian, a); err != nil {
				return nil, err
			}
		}

		has := byte(0)
		if x.HasRGBA {
			has = 1
		}
		_ = bw.WriteByte(has)
		if x.HasRGBA {
			_ = bw.WriteByte(x.RGBA.R)
			_ = bw.WriteByte(x.RGBA.G)
			_ = bw.WriteByte(x.RGBA.B)
			_ = bw.WriteByte(x.RGBA.A)
		}
		if err := binary.Write(bw, binary.LittleEndian, x.Pen); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("canvas: unknown command type %T", v)
	}

	if err := bw.Flush(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func canvasDecodePayload(payload []byte) (any, error) {
	br := bytes.NewReader(payload)
	ver, err := br.ReadByte()
	if err != nil {
		return nil, err
	}
	if ver != canvasProtoVersion {
		return nil, fmt.Errorf("canvas: proto version %d", ver)
	}
	op, err := br.ReadByte()
	if err != nil {
		return nil, err
	}

	switch op {
	case canvasOpClose:
		return CanvasWireClose{}, nil
	case canvasOpSetBG:
		var rgba [4]byte
		if _, err := io.ReadFull(br, rgba[:]); err != nil {
			return nil, err
		}
		return CanvasWireSetBG{RGBA: color.RGBA{rgba[0], rgba[1], rgba[2], rgba[3]}}, nil
	case canvasOpSetFG:
		var rgba [4]byte
		if _, err := io.ReadFull(br, rgba[:]); err != nil {
			return nil, err
		}
		return CanvasWireSetFG{RGBA: color.RGBA{rgba[0], rgba[1], rgba[2], rgba[3]}}, nil
	case canvasOpCmd:
		var oplen uint16
		if err := binary.Read(br, binary.LittleEndian, &oplen); err != nil {
			return nil, err
		}
		opb := make([]byte, int(oplen))
		if _, err := io.ReadFull(br, opb); err != nil {
			return nil, err
		}
		var argc uint16
		if err := binary.Read(br, binary.LittleEndian, &argc); err != nil {
			return nil, err
		}
		args := make([]int32, int(argc))
		for i := 0; i < int(argc); i++ {
			if err := binary.Read(br, binary.LittleEndian, &args[i]); err != nil {
				return nil, err
			}
		}
		has, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		cmd := CanvasWireCmd{Op: string(opb), Args: args}
		if has == 1 {
			var rgba [4]byte
			if _, err := io.ReadFull(br, rgba[:]); err != nil {
				return nil, err
			}
			cmd.HasRGBA = true
			cmd.RGBA = color.RGBA{rgba[0], rgba[1], rgba[2], rgba[3]}
		}
		if err := binary.Read(br, binary.LittleEndian, &cmd.Pen); err != nil {
			return nil, err
		}
		return cmd, nil
	default:
		return nil, fmt.Errorf("canvas: unknown opcode %d", op)
	}
}

func canvasObserveDecoded(cmd any) {
	b, err := json.Marshal(cmd)
	if err != nil {
		return
	}
	_, _ = os.Stderr.Write(append(b, '\n'))
}
