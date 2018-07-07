// Decodes NMEA sentences, stores in redis, and poorly maps

package main

import (
	"fmt"
	ais "github.com/eholzbach/aislib"
	"github.com/go-redis/redis"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func updateRedis(db0 *redis.Client, db1 *redis.Client, m uint32, a map[string]interface{}) {
	mmsi := strconv.FormatUint(uint64(m), 10)

	// write to db0
	_, err := db0.HMSet(mmsi, a).Result()
	if err != nil {
		fmt.Println(err)
		return
	}

	// set to expire
	db0.Expire(mmsi, 12*time.Hour)

	// write to db1
	_, err = db1.HMSet(mmsi, a).Result()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {

	// get config
	rds, fwd, fwdadr, lsn := Config()

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
	var con []*net.UDPConn
	if fwd == true {
		for _, v := range fwdadr {
			serverAddr, err := net.ResolveUDPAddr("udp", v)
			checkError(err)
			c, err := net.DialUDP("udp", nil, serverAddr)
			checkError(err)
			defer c.Close()
			con = append(con, c)
		}
	}

	// start web server
	go api(db0)

	buffer := make([]byte, 1024)

	// start aislib router
	send := make(chan string, 1024*8)
	receive := make(chan ais.Message, 1024*8)
	failed := make(chan ais.FailedSentence, 1024*8)
	go ais.Router(send, receive, failed)

	for {
		// read packet from buffer
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// if forwarder configured then forward
		if fwd == true {
			for _, v := range con {
				v.Write(buffer[0:n])
			}
		}

		var message ais.Message
		var problematic ais.FailedSentence
		raw := string(buffer[0:n])
		raw = strings.Replace(raw, "\r\n", "", -2)

		send <- raw
		select {
		case message = <-receive:
			switch message.Type {
			case 1, 2, 3:
				t, _ := ais.DecodeClassAPositionReport(message.Payload)
				a := decodeA(t)
				updateRedis(db0, db1, t.MMSI, a)
			case 4:
				t, _ := ais.DecodeBaseStationReport(message.Payload)
				a := decodeBase(t)
				updateRedis(db0, db1, t.MMSI, a)
			case 5:
				t, _ := ais.DecodeStaticVoyageData(message.Payload)
				a := decodeV(t)
				updateRedis(db0, db1, t.MMSI, a)
			case 8:
				t, _ := ais.DecodeBinaryBroadcast(message.Payload)
				a := decodeBinary(t)
				updateRedis(db0, db1, t.MMSI, a)
			case 18:
				t, _ := ais.DecodeClassBPositionReport(message.Payload)
				a := decodeB(t)
				updateRedis(db0, db1, t.MMSI, a)
			default:
			}
		case problematic = <-failed:
			fmt.Println(problematic)
		}
	}
}
