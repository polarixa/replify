package randn

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/polarixa/replify/pkg/sysx"
)

// readMachineIDEnv attempts to read the machine ID from the 'XID_MACHINE_ID' environment variable.
// It returns a 3-byte slice if the variable is a valid numeric value within the 0 to 16,777,215 range,
// otherwise it returns nil.
func readMachineIDEnv() []byte {
	if !sysx.Hasenv("XID_MACHINE_ID") {
		return nil
	}
	num := sysx.GetenvInt("XID_MACHINE_ID", 0)
	if num < 0 || num > 0xFFFFFF {
		return nil
	}
	return []byte{byte(num >> 16), byte(num >> 8), byte(num)}
}

// readMachineID returns the machine ID extracted from the 'XID_MACHINE_ID' environment variable
// or derived from the platform's host ID. If both methods fail, it falls back to a random sequence.
func readMachineID() []byte {
	if id := readMachineIDEnv(); len(id) == 3 {
		return id
	}

	id := make([]byte, 3)
	mID, err := readPlatformMachineID()
	if err != nil || len(mID) == 0 {
		mID, err = os.Hostname()
	}
	if err == nil && len(mID) != 0 {
		hw := sha256.New()
		hw.Write([]byte(mID))
		copy(id, hw.Sum(nil))
	} else {
		if _, randErr := rand.Reader.Read(id); randErr != nil {
			panic(fmt.Errorf("randn: cannot get hostname nor generate a random number: %v; %v", err, randErr))
		}
	}
	return id
}

// encode converts a 12-byte slice to a 20-byte base32 hex lowercased representation
// using an unrolled version of the standard library base32 algorithm. It is optimized
// for performance and requires dst to be at least 20 bytes and id to be at least 12 bytes.
func encode(dst, id []byte) {
	_ = dst[19]
	_ = id[11]

	dst[19] = encoding[(id[11]<<4)&0x1F]
	dst[18] = encoding[(id[11]>>1)&0x1F]
	dst[17] = encoding[(id[11]>>6)|(id[10]<<2)&0x1F]
	dst[16] = encoding[id[10]>>3]
	dst[15] = encoding[id[9]&0x1F]
	dst[14] = encoding[(id[9]>>5)|(id[8]<<3)&0x1F]
	dst[13] = encoding[(id[8]>>2)&0x1F]
	dst[12] = encoding[id[8]>>7|(id[7]<<1)&0x1F]
	dst[11] = encoding[(id[7]>>4)|(id[6]<<4)&0x1F]
	dst[10] = encoding[(id[6]>>1)&0x1F]
	dst[9] = encoding[(id[6]>>6)|(id[5]<<2)&0x1F]
	dst[8] = encoding[id[5]>>3]
	dst[7] = encoding[id[4]&0x1F]
	dst[6] = encoding[id[4]>>5|(id[3]<<3)&0x1F]
	dst[5] = encoding[(id[3]>>2)&0x1F]
	dst[4] = encoding[id[3]>>7|(id[2]<<1)&0x1F]
	dst[3] = encoding[(id[2]>>4)|(id[1]<<4)&0x1F]
	dst[2] = encoding[(id[1]>>1)&0x1F]
	dst[1] = encoding[(id[1]>>6)|(id[0]<<2)&0x1F]
	dst[0] = encoding[id[0]>>3]
}

// decode decodes a 20-byte base32 hex lowercased representation into a XID.
// It returns true if the source is valid and decoding was successful, false otherwise.
// Performance is optimized using an unrolled version of the standard base32 algorithm.
func decode(id *XID, src []byte) bool {
	_ = src[19]
	_ = id[11]

	id[11] = decodeBytes[src[17]]<<6 | decodeBytes[src[18]]<<1 | decodeBytes[src[19]]>>4
	if encoding[(id[11]<<4)&0x1F] != src[19] {
		return false
	}
	id[10] = decodeBytes[src[16]]<<3 | decodeBytes[src[17]]>>2
	id[9] = decodeBytes[src[14]]<<5 | decodeBytes[src[15]]
	id[8] = decodeBytes[src[12]]<<7 | decodeBytes[src[13]]<<2 | decodeBytes[src[14]]>>3
	id[7] = decodeBytes[src[11]]<<4 | decodeBytes[src[12]]>>1
	id[6] = decodeBytes[src[9]]<<6 | decodeBytes[src[10]]<<1 | decodeBytes[src[11]]>>4
	id[5] = decodeBytes[src[8]]<<3 | decodeBytes[src[9]]>>2
	id[4] = decodeBytes[src[6]]<<5 | decodeBytes[src[7]]
	id[3] = decodeBytes[src[4]]<<7 | decodeBytes[src[5]]<<2 | decodeBytes[src[6]]>>3
	id[2] = decodeBytes[src[3]]<<4 | decodeBytes[src[4]]>>1
	id[1] = decodeBytes[src[1]]<<6 | decodeBytes[src[2]]<<1 | decodeBytes[src[3]]>>4
	id[0] = decodeBytes[src[0]]<<3 | decodeBytes[src[1]]>>2
	return true
}
