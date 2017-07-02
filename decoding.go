package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type PositionReport struct {
	Type     uint8
	Repeat   uint8
	MMSI     uint32
	Speed    float32 // speed over ground - SOG (sc U3)
	Accuracy bool    // position accuracy
	Lon      float64 // (sc I4)
	Lat      float64 // (sc I4)
	Course   float32 //course over ground - COG (sc U1)
	Heading  uint16  // true heading - HDG
	Second   uint8   // timestamp
	RAIM     bool    // RAIM flag
	Radio    uint32  // Radio status
}

type ClassAPositionReport struct {
	PositionReport
	Status   uint8   // navigation status (enumerated type)
	Turn     float32 // rate of turn - ROT (sc - Special Calc I3)
	Maneuver uint8   // maneuver indicator (enumerated)
}

type Plot struct {
	MMSI string
	Lon  string
	Lat  string
}

var NavigationStatusCodes = [...]string{
	"Under way using engine", "At anchor", "Not under command", "Restricted maneuverability",
	"Constrained by her draught", "Moored", "Aground", "Engaged in Fishing", "Under way sailing",
	"Reserved for future amendment of Navigational Status for HSC",
	"Reserved for future amendment of Navigational Status for WIG", "Reserved for future use",
	"Reserved for future use", "Reserved for future use", "AIS-SART is active", "Not defined",
}

func aisChar(c byte) byte {
	c -= 48
	if c > 40 {
		c -= 8
	}
	return c
}

func aisChecksum(raw string) bool {
	if len(raw) < 5 {
		return false
	}

	r := strings.Replace(raw, "\r\n", "", -2)

	x, err := hex.DecodeString(r[len(r)-2:])
	if err != nil {
		return false
	}

	y := []byte(r[1 : len(r)-3])
	z := y[0]
	for i := 1; i < len(y); i++ {
		z ^= y[i]
	}

	if x[0] != z {
		return false
	}
	return true
}

func aisType(payload string) uint8 {
	data := []byte(payload[:1])
	return aisChar(data[0])
}

func bitsToInt(first, last int, payload []byte) uint32 {
	size := uint(last - first) // Bit fields start at 0
	processed, remain := uint(0), uint(0)
	result, temp := uint32(0), uint32(0)

	from := first / 6
	forTimes := last/6 - from

	if len(payload)*6 < last+1 {
		return 0
	}
	for i := 0; i <= forTimes; i++ {
		temp = uint32(payload[from+i]) - 48
		if temp > 40 {
			temp -= 8
		}

		if i == 0 {
			remain = uint(first % 6)
			processed = 5 - remain
			temp = temp << (31 - processed) >> (31 - size)
		} else if i < forTimes {
			processed = processed + 6
			temp = temp << (size - processed)
		} else {
			remain = uint(last%6) + 1
			temp = temp >> (6 - remain)
		}
		result = result | temp
	}
	return result
}

func cbnCoordinates(first int, data []byte) (float64, float64) {
	lon := float64((int32(bitsToInt(first, first+27, data)) << 4)) / 16
	lat := float64((int32(bitsToInt(first+28, first+54, data)) << 5)) / 32

	return CoordinatesMin2Deg(lon, lat)
}

func cbnSpeed(first int, data []byte) float32 {
	speed := float32(bitsToInt(first, first+9, data))
	if speed < 1022 {
		speed /= 10
	}
	return speed
}

func cbnBool(bit int, data []byte) bool {
	if bitsToInt(bit, bit, data) == 1 {
		return true
	}
	return false
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func CoordinatesMin2Deg(minLon, minLat float64) (float64, float64) {
	lonSign := 1.0
	latSign := 1.0

	if math.Signbit(minLon) {
		minLon = -minLon
		lonSign = -1
	}
	if math.Signbit(minLat) {
		minLat = -minLat
		latSign = -1
	}

	degrees := float64(int(minLon / 600000))
	minutes := float64(minLon-600000*degrees) / 10000
	lon := degrees + minutes/60

	degrees = float64(int(minLat / 600000))
	minutes = float64(minLat-600000*degrees) / 10000
	lat := degrees + minutes/60

	return lonSign * lon, latSign * lat
}

func decodeClassA(payload string, count int, seq int) ClassAPositionReport {
	var r ClassAPositionReport
	if count == seq {
		data := []byte(payload)
		r.Repeat = uint8(bitsToInt(6, 7, data))
		r.MMSI = bitsToInt(8, 37, data)
		r.Status = uint8(bitsToInt(38, 41, data))
		r.Turn = float32(int8(bitsToInt(42, 49, data)))
		if r.Turn != 0 && r.Turn <= 126 && r.Turn >= -126 {
			sign := float32(1)
			if math.Signbit(float64(r.Turn)) {
				sign = -1
			}
			r.Turn = sign * (r.Turn / 4.733) * (r.Turn / 4.733)

		}

		r.Speed = cbnSpeed(50, data)
		r.Accuracy = cbnBool(60, data)
		r.Lon, r.Lat = cbnCoordinates(61, data)
		r.Course = float32(bitsToInt(116, 127, data)) / 10
		r.Heading = uint16(bitsToInt(128, 136, data))
		r.Second = uint8(bitsToInt(137, 142, data))
		r.Maneuver = uint8(bitsToInt(143, 144, data))
		r.RAIM = cbnBool(148, data)
		r.Radio = bitsToInt(149, 167, data)
	}
	return r
}

func extract(raw string) (int, int, int, string, string, int) {
	x := strings.Split(raw, ",")
	c, _ := strconv.Atoi(x[1])
	s, _ := strconv.Atoi(x[2])
	m, _ := strconv.Atoi(x[3])
	n := strings.Split(x[6], "*")
	o, _ := strconv.Atoi(n[0])
	return c, s, m, x[4], x[5], o
}
