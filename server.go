package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
	"text/template"
)

func current(db0 *redis.Client) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		a, _, _ := db0.Scan(0, "", 10000).Result()
		var data []string
		for _, v := range a {
			b, _ := db0.HMGet(v, "longitude", "latitude").Result()
			c := fmt.Sprintf("L.marker([%s,%s], {title: '%s'}).addTo(mymap).bindPopup('<a href=\"https://www.marinetraffic.com/en/ais/details/ships/mmsi:%s\" target=\"_blank\">%s</a>');", b[1], b[0], v, v, v)
			data = append(data, c)
		}
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, data)
	}
	return http.HandlerFunc(fn)
}

func api(db0 *redis.Client) {
	mux := http.NewServeMux()
	mux.Handle("/v1/current", current(db0))
	http.ListenAndServe("127.0.0.1:8080", mux)
}
