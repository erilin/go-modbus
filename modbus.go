package modbus

import "log"

type ReaderWriter interface {
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	Flush() error
}

type Modbus struct {
	rw ReaderWriter
}

//SendFunc3 Address (addr), Start(s), Number of registers(r), Values (v)
func (mb *Modbus) SendFunc3(addr byte, s uint16, r uint16, v []int16) ([]Register, error) {
	mb.rw.Flush()

	//Message is 1 addr + 1 fcn + 2 start + 2 reg + 1 count + 2 * reg vals + 2 CRC
	msg := make([]byte, 9+2*r)
	//Function 16 response is fixed at 8 bytes
	//rsp := [8]byte{}

	//Add bytecount to message:
	msg[6] = byte(r * 2)
	//Put write values into message prior to sending:
	for i := uint16(0); i < r; i++ {
		msg[7+2*i] = byte(v[i] >> 8)
		msg[8+2*i] = byte(v[i])
	}

	buildMessage(addr, F03, 0, r, msg)

	err := mb.rw.Flush()
	if err != nil {
		log.Printf("ReaderWrite.Flush: %s", err.Error())
	}
	_, err = mb.rw.Write(msg)
	if err != nil {
		log.Printf("ReaderWrite.Write: %s", err.Error())
	}

	resp := []byte{}
	_, err = mb.rw.Read(resp)

	return []Register{}, nil
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

const (
	F03 int = 3
	F16 int = 16
)
