package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

const ntpServer = "pool.ntp.org"

func main() {
	currentTime, err := getNetworkTime(ntpServer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(currentTime.Format(time.RFC3339))
}

// getNetworkTime запрашивает текущее время у NTP-сервера.
func getNetworkTime(server string) (time.Time, error) {
	t, err := ntp.Time(server)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get time from NTP server %q: %w", server, err)
	}

	return t, nil
}
