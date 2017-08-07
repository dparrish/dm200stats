package collector

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`var ([^=]+)="?([^;"]+)"?;`)

type Stats struct {
	SyncSpeedDown   float64
	SyncSpeedUp     float64
	AttenuationDown float64
	AttenuationUp   float64
	NoiseDown       float64
	NoiseUp         float64
	BytesDown       float64
	BytesUp         float64
}

// Collect collects stats from the provided IP, doing basic auth with the provided username and password.
func Collect(ctx context.Context, ip, username, password string) (*Stats, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/RST_statistic.htm", ip), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.SetBasicAuth(username, password)

	cl := &http.Client{}
	r, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Pull all vars out of the blob of HTML & Javascript.
	vars := make(map[string]string)
	for _, match := range re.FindAllStringSubmatch(string(body), -1) {
		vars[match[1]] = strings.TrimSpace(match[2])
	}

	// Ensure all the required variables are there.
	for _, v := range []string{"conn_down", "conn_up", "line_down", "line_up", "noise_down", "noise_up", "lan_txpkts", "lan_rxpkts"} {
		if _, ok := vars[v]; !ok {
			return nil, fmt.Errorf("missing var %q", v)
		}
	}

	return &Stats{
		SyncSpeedDown:   parseSpeed(vars["conn_down"]),
		SyncSpeedUp:     parseSpeed(vars["conn_up"]),
		AttenuationDown: parseDb(vars["line_down"]),
		AttenuationUp:   parseDb(vars["line_up"]),
		NoiseDown:       parseDb(vars["noise_down"]),
		NoiseUp:         parseDb(vars["noise_up"]),
		BytesDown:       parseFloat(vars["lan_txpkts"]),
		BytesUp:         parseFloat(vars["lan_rxpkts"]),
	}, nil
}

func parseSpeed(speed string) float64 {
	parts := strings.Split(speed, " ")
	if len(parts) != 2 {
		return 0
	}
	v, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	switch strings.ToLower(parts[1]) {
	case "bps":
		break
	case "kbps":
		v *= 1024
	case "mbps":
		v *= 1024 * 1024
	case "gbps":
		v *= 1024 * 1024 * 1024
	}
	return v
}

func parseDb(db string) float64 {
	parts := strings.Split(db, " ")
	if len(parts) != 2 {
		return 0
	}
	v, _ := strconv.ParseFloat(parts[0], 64)
	return v
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
