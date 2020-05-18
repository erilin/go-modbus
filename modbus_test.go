package modbus

import "testing"

func TestSendFunc3(t *testing.T) {
	/*
	   Request
	   This command is requesting the content of analog output holding registers # 0 to 127 from the slave device with address 1.
	   01 03 0000 16 7687

	   01: The Slave Address (01 hex = address1 )
	   03: The Function Code 3 (read Analog Output Holding Registers)
	   0000: The Data Address of the first register requested.
	   0016: The total number of registers requested. (read 3 registers 40108 to 40110)
	   7687: The CRC (cyclic redundancy check) for error checking.

	   Response
	   01 03 06 AE41 5652 4340 49AD

	   01: The Slave Address (11 hex = address17 )
	   03: The Function Code 3 (read Analog Output Holding Registers)
	   06: The number of data bytes to follow (3 registers x 2 bytes each = 6 bytes)
	   AE41: The contents of register 40108
	   5652: The contents of register 40109
	   4340: The contents of register 40110
	   49AD: The CRC (cyclic redundancy check).
	*/

	rw := fakeReadWriter{}
	mb := NewModbus(&rw)

	mb.SendFunc3(byte(1), 0, 8)

	if rw.written[0] != 0x01 {
		t.Errorf("Address should be 0x01. Actual 0x%x", rw.written[0])
	}
	if rw.written[1] != 0x03 {
		t.Errorf("Modbus func should be 0x03 (3). Actual 0x%x", rw.written[1])
	}
	if rw.written[2]+rw.written[3] != 0x0000 {
		t.Errorf("Address of first register should be 0x0000 (0). Actual 0x%x", rw.written[2]+rw.written[3])
	}
	if rw.written[4]+rw.written[5] != 0x0008 {
		t.Errorf("Number of requested registers should be 0x0008 (8). Actual 0x%x", rw.written[4]+rw.written[5])
	}
}

type fakeReadWriter struct {
	written []byte
	read    []byte
}

func (rw *fakeReadWriter) Read(buf []byte) (int, error) {
	rw.read = buf
	return len(buf), nil
}

func (rw *fakeReadWriter) Write(buf []byte) (int, error) {
	rw.written = buf
	return len(buf), nil
}
