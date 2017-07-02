package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"net/http"
)

func current(db0 *redis.Client) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		a, _, _ := db0.Scan(0, "", 10000).Result()
		h, _ := ioutil.ReadFile("header.html")
		f, _ := ioutil.ReadFile("footer.html")
		fmt.Fprint(w, string(h))
		for _, v := range a {
			b, _ := db0.HMGet(v, "Lon", "Lat").Result()
			fmt.Fprint(w, "L.marker([", b[1], ",", b[0], "], {title: '", v, "'}).addTo(mymap).bindPopup('<a href=\"https://www.marinetraffic.com/en/ais/details/ships/mmsi:", v, "\" target=\"_blank\">", v, "</a>');", "\n")
		}
		fmt.Fprint(w, string(f))
		//              w.Write([]byte(v))
	}
	return http.HandlerFunc(fn)
}

func api(db0 *redis.Client) {
	mux := http.NewServeMux()
	c := current(db0)
	mux.Handle("/v1/current", c)
	http.ListenAndServe("127.0.0.1:8080", mux)
}
