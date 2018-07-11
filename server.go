package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
	"text/template"
)

func getType(t interface{}) string {
	var a string
	switch t {
	case "Ship":
		a = "shipIcon"
	case "Coastal Station":
		a = "baseIcon"
	default:
		a = "otherIcon"
	}
	return a
}

func renderPage(db0 *redis.Client) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		a, _, _ := db0.Scan(0, "", 1000).Result()
		var data []string
		for _, v := range a {
			b, _ := db0.HMGet(v, "longitude", "latitude", "device", "country").Result()
			if b[0] != nil && b[1] != nil {
				icon := getType(b[2])
				c := fmt.Sprintf("L.marker([%s,%s], {icon: %s}, {title: '%s'}).addTo(mymap).bindPopup('%s<br><br>mmsi: %s<br>country: %s').on('mouseover', function (e) { this.openPopup()}).on('mouseout', function (e) { this.closePopup()}).on('click', function (e) { window.open(\"https://www.marinetraffic.com/en/ais/details/ships/mmsi:%s\")});", b[1], b[0], icon, v, b[2], v, b[3], v)
				data = append(data, c)
			}
		}
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, data)
	}
	return http.HandlerFunc(fn)
}

func webServer(db0 *redis.Client) {
	mux := http.NewServeMux()
	mux.Handle("/", renderPage(db0))
	http.ListenAndServe("127.0.0.1:8080", mux)
}
