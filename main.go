// Decodes NMEA sentences and stores them in redis. Binary decoding taken from aislib.

package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {

	// get config
	rds, fwd, fwdadr, lsn := config()

	// set up udp listener
	server, err := net.ResolveUDPAddr("udp", lsn)
	checkError(err)

	conn, err := net.ListenUDP("udp", server)
	checkError(err)
	defer conn.Close()

	// connect to redis ttl and historical db
	db0 := redis.NewClient(&redis.Options{
		Addr:     rds,
		Password: "",
		DB:       0,
	})

	_, err = db0.Ping().Result()
	checkError(err)

	db1 := redis.NewClient(&redis.Options{
		Addr:     rds,
		Password: "",
		DB:       1,
	})

	_, err = db1.Ping().Result()
	checkError(err)

	// set up udp forwarder
	var con *net.UDPConn
	if fwd == true {
		serverAddr, err := net.ResolveUDPAddr("udp", fwdadr)
		checkError(err)
		con, err = net.DialUDP("udp", nil, serverAddr)
		checkError(err)
		defer con.Close()
	}

	// start web server
	go api(db0)

	buffer := make([]byte, 1024)

	for {
		// read packet from buffer
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// if forwarder configured then forward
		if fwd == true {
			con.Write(buffer[0:n])
		}

		raw := string(buffer[0:n])

		// drop packet if not nmea header
		prefix := strings.HasPrefix(raw, "!AIVDM")
		if prefix != true {
			continue
		}

		// drop message if incorrect checksum
		c := aisChecksum(raw)
		if c != true {
			continue
		}

		count, seq, multi, channel, payload, padding := extract(raw)

		switch aisType(payload) {
		case 1, 2, 3:
			// decode and map
			a := decodeClassA(payload, count, seq)
			b := map[string]interface{}{
				"Type":     a.Type,
				"Repeat":   a.Repeat,
				"MMSI":     a.MMSI,
				"Speed":    a.Speed,
				"Accuracy": a.Accuracy,
				"Lon":      a.Lon,
				"Lat":      a.Lat,
				"Course":   a.Course,
				"Heading":  a.Heading,
				"Second":   a.Second,
				"RAIM":     a.RAIM,
				"Radio":    a.Radio,
				"Status":   a.Status,
				"Turn":     a.Turn,
				"Maneuver": a.Maneuver,
			}

			// write to db0
			_, err := db0.HMSet(strconv.FormatUint(uint64(a.MMSI), 10), b).Result()
			if err != nil {
				fmt.Println(err)
				continue
			}

			// set to expire
			db0.Expire(strconv.FormatUint(uint64(a.MMSI), 10), 12*time.Hour)

			// write to db1
			_, err = db1.HMSet(strconv.FormatUint(uint64(a.MMSI), 10), b).Result()
			if err != nil {
				fmt.Println(err)
				continue
			}

		default:
			continue
		}

		// dont know if i need these
		_ = multi
		_ = channel
		_ = padding

	}
}
