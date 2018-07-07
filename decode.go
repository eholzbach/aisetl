package main

import (
	"fmt"
	ais "github.com/eholzbach/aislib"
	"strconv"
)

func type5size2String(min, max, size int) string {
	s := ""
	switch size {
	case min:
		s = "Not available"
	case max:
		s = ">" + strconv.Itoa(max) + " meters"
	default:
		s = strconv.Itoa(size) + " meters"
	}
	return s
}

func decodeBase(m ais.BaseStationReport) map[string]interface{} {
	accuracy := "High accuracy (<10m)"
	if m.Accuracy == false {
		accuracy = "Low accuracy (>10m)"
	}

	raim := "not in use"
	if m.RAIM == true {
		raim = "in use"
	}

	device, country := ais.DecodeMMSI(m.MMSI)

	message := map[string]interface{}{
		"type":      "4",
		"repeat":    m.Repeat,
		"mmsi":      m.MMSI,
		"device":    device,
		"country":   country,
		"time":      m.Time.String(),
		"accuracy":  accuracy,
		"longitude": m.Lon,
		"latitude":  m.Lat,
		"epfd":      ais.EpfdFixTypes[m.EPFD],
		"raim":      raim,
	}

	return message
}

func decodeA(m ais.ClassAPositionReport) map[string]interface{} {
	turn := ""
	switch {
	case m.Turn == 0:
		turn = "not turning"
	case m.Turn == 127:
		turn = "right at more than 5deg/30s"
	case m.Turn == -127:
		turn = "left at more than 5deg/30s"
	case m.Turn == -128:
		turn = "no turn information"
	case m.Turn > 0 && m.Turn < 127:
		turn = "right at " + strconv.FormatFloat(float64(m.Turn), 'f', 3, 32)
	case m.Turn < 0 && m.Turn > -127:
		turn = "left at " + strconv.FormatFloat(float64(-m.Turn), 'f', 3, 32)
	}

	speed := ""
	switch {
	case m.Speed <= 102:
		speed = strconv.FormatFloat(float64(m.Speed), 'f', 1, 32) + " knots"
	case m.Speed == 1022:
		speed = ">102.2 knots"
	case m.Speed == 1023:
		speed = "information not available"
	}

	accuracy := "High accuracy (<10m)"
	if m.Accuracy == false {
		accuracy = "Low accuracy (>10m)"
	}

	course := ""
	switch {
	case m.Course < 360:
		course = fmt.Sprintf("%.1f", m.Course)
	case m.Course == 360:
		course = "not available"
	case m.Course > 360:
		course = "please report this to developer"
	}

	heading := ""
	switch {
	case m.Heading <= 359:
		heading = fmt.Sprintf("%d", m.Heading)
	case m.Heading == 511:
		heading = "not available"
	case m.Heading != 511 && m.Heading >= 360:
		heading = "please report this to developer"
	}

	maneuver := ""
	switch {
	case m.Maneuver == 0:
		maneuver = "not available"
	case m.Maneuver == 1:
		maneuver = "no special maneuver"
	case m.Maneuver == 2:
		maneuver = "special maneuver"
	}
	raim := "not in use"
	if m.RAIM == true {
		raim = "in use"
	}

	device, country := ais.DecodeMMSI(m.MMSI)

	message := map[string]interface{}{
		"type":      m.Type,
		"repeat":    m.Repeat,
		"mmsi":      m.MMSI,
		"device":    device,
		"country":   country,
		"status":    ais.NavigationStatusCodes[m.Status],
		"turn":      turn,
		"speed":     speed,
		"accuracy":  accuracy,
		"longitude": m.Lon,
		"latitude":  m.Lat,
		"course":    course,
		"heading":   heading,
		"manuever":  maneuver,
		"raim":      raim,
	}

	return message
}

func decodeB(m ais.ClassBPositionReport) map[string]interface{} {
	speed := ""
	switch {
	case m.Speed <= 102:
		speed = strconv.FormatFloat(float64(m.Speed), 'f', 1, 32) + " knots"
	case m.Speed == 1022:
		speed = ">102.2 knots"
	case m.Speed == 1023:
		speed = "information not available"
	}

	accuracy := "High accuracy (<10m)"
	if m.Accuracy == false {
		accuracy = "Low accuracy (>10m)"
	}

	course := ""
	switch {
	case m.Course < 360:
		course = fmt.Sprintf("%.1f", m.Course)
	case m.Course == 360:
		course = "not available"
	case m.Course > 360:
		course = "please report this to developer"
	}

	heading := ""
	switch {
	case m.Heading <= 359:
		heading = fmt.Sprintf("%d°", m.Heading)
	case m.Heading == 511:
		heading = "not available"
	case m.Heading != 511 && m.Heading >= 360:
		heading = "please report this to developer"
	}

	device, country := ais.DecodeMMSI(m.MMSI)

	message := map[string]interface{}{
		"type":      "18",
		"repeat":    m.Repeat,
		"mmsi":      m.MMSI,
		"device":    device,
		"country":   country,
		"speed":     speed,
		"accuracy":  accuracy,
		"longitude": m.Lon,
		"latitude":  m.Lat,
		"course":    course,
		"heading":   heading,
		"csu":       m.CSUnit,
		"display":   m.Display,
		"dsc":       m.DSC,
		"band":      m.Band,
		"message22": m.Msg22,
		"assigned":  m.Assigned,
		"raim":      m.RAIM,
	}

	return message
}

func decodeV(m ais.StaticVoyageData) map[string]interface{} {
	imo := ""
	if m.IMO == 0 {
		imo = "Inland Vessel"
	} else {
		imo = strconv.Itoa(int(m.IMO))
	}

	draught := ""
	if m.Draught == 0 {
		draught = "Not available"
	} else {
		draught = strconv.Itoa(10*int(m.IMO)) + " meters"
	}

	device, country := ais.DecodeMMSI(m.MMSI)

	message := map[string]interface{}{
		"type":        "5",
		"repeat":      m.Repeat,
		"mmsi":        m.MMSI,
		"device":      device,
		"country":     country,
		"version":     m.AisVersion,
		"imo":         imo,
		"callsign":    m.Callsign,
		"vesselname":  m.VesselName,
		"shiptype":    ais.ShipType[int(m.ShipType)],
		"dimtobow":    type5size2String(0, 511, int(m.ToBow)),
		"dimtostern":  type5size2String(0, 511, int(m.ToStern)),
		"dimtoport":   type5size2String(0, 511, int(m.ToPort)),
		"dimtostrbrd": type5size2String(0, 511, int(m.ToStarboard)),
		"epfd":        ais.EpfdFixTypes[m.EPFD],
		"eta":         m.ETA.String(),
		"draught":     draught,
		"destination": m.Destination,
	}

	return message
}

func decodeBinary(m ais.BinaryBroadcast) map[string]interface{} {

	device, country := ais.DecodeMMSI(m.MMSI)
	dacfid := fmt.Sprintf("%d-%d %s", m.DAC, m.FID, ais.BinaryBroadcastType[int(m.DAC)][int(m.FID)])

	message := map[string]interface{}{
		"type":    "8",
		"repeat":  m.Repeat,
		"mmsi":    m.MMSI,
		"device":  device,
		"country": country,
		"dacfid":  dacfid,
	}

	return message
}
