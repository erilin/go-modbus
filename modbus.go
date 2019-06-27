package modbus

import (
	"fmt"
	"log"
)

//ReaderWriter interface for read or write to dupline master
type ReaderWriter interface {
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	Flush() error
}

//Modbus struct for interacting
type Modbus struct {
	rw ReaderWriter
}

//NewModbus creates Modbus
func NewModbus(rw ReaderWriter) Modbus {
	return Modbus{rw: rw}
}

//SendFunc3 Address (addr), Start(s), Number of registers(r), Values (v)
func (mb *Modbus) SendFunc3(addr byte, s uint16, r uint16) ([]Register, error) {
	//Function 3 request is always 8 bytes:
	msg := make([]byte, 8)
	//Function 3 response buffer:
	rsp := make([]byte, 5+2*r)

	buildMessage(addr, F03, 0, r, msg)

	err := mb.rw.Flush()
	if err != nil {
		log.Printf("ReaderWrite. Flush: %s", err.Error())
		return []Register{}, nil
	}
	_, err = mb.rw.Write(msg)
	if err != nil {
		log.Printf("ReaderWrite. Write: %s", err.Error())
		return []Register{}, nil
	}

	//Get response
	n, err := mb.rw.Read(rsp)
	if err != nil {
		log.Printf("ReaderWrite. Read: %s", err.Error())
		return []Register{}, nil
	}

	//Evaluate message:
	err = checkCRC(rsp)
	if err != nil {
		log.Printf("CheckCRC: %s", err.Error())
		return []Register{}, nil
	}

	//Return requested register values:
	c := (n - 5) / 2
	regs := make([]Register, c)

	for i := 0; i < c; i++ {
		reg := Register{
			HiByte: rsp[2*i+3],
			LoByte: rsp[2*i+4],
		}
		regs[i] = reg
	}

	return regs, nil
}

func buildMessage(addr byte, fn int, s uint16, r uint16, msg []byte) {
	msg[0] = addr
	msg[1] = byte(fn)
	msg[2] = byte(s >> 8)
	msg[3] = byte(s)
	msg[4] = byte(r >> 8)
	msg[5] = byte(r)

	crc := crc(msg)

	msg[len(msg)-2] = crc[0]
	msg[len(msg)-1] = crc[1]
}

//crc func expects a modbus message of any length as well as a 2 byte CRC array in which to return the CRC values
func crc(msg []byte) [2]byte {
	full := uint16(0xFFFF)

	for i := 0; i < len(msg)-2; i++ {
		full = full ^ uint16(msg[i])
		for j := 0; j < 8; j++ {
			lsb := full & 0x0001
			full = (full >> 1) & 0x7FFF
			if lsb == 1 {
				full = full ^ 0xA001
			}
		}
	}

	crc := [2]byte{}
	crc[1] = byte((full >> 8) & 0xFF)
	crc[0] = byte(full & 0xFF)
	return crc
}

func checkCRC(r []byte) error {
	crc := crc(r)
	l := len(r)
	vld := crc[0] == r[l-2] && crc[1] == r[l-1]

	if vld {
		return nil
	}

	return fmt.Errorf("CRC error. Invalid CRC in response")
}

const (
	//F03 Modbus 3 func
	F03 int = 3
	//F16 Modbus 16 func
	F16 int = 16
)
