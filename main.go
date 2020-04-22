package main

import (
	"log"
	"net/http"
	"text/template"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

type Stats struct {
	MemoryInfo  []Memory
	NetworkInfo []Network
}

type Memory struct {
	MemTotal uint64
	MemFree  uint64
}

type Network struct {
	Iface         string
	ReceivedBytes uint64
	SentBytes     uint64
}

// Helper function to convert Kb -> Mb
func convertToMB(kb uint64) uint64 {
	return kb / 1000
}

func (s *Stats) getMemoryInfo() error {
	stat, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return err
	}
	memoryInfo := Memory{MemTotal: convertToMB(stat.MemTotal), MemFree: convertToMB(stat.MemFree + stat.MemAvailable)}
	s.MemoryInfo = append(s.MemoryInfo, memoryInfo)
	return nil
}

func (s *Stats) getNetworkInfo() error {
	stat, err := linuxproc.ReadNetworkStat("/proc/net/dev")
	if err != nil {
		return err
	}

	for _, network := range stat {
		iface := Network{Iface: network.Iface, ReceivedBytes: convertToMB(network.RxBytes), SentBytes: convertToMB(network.TxBytes)}
		s.NetworkInfo = append(s.NetworkInfo, iface)
	}
	return nil
}

// Get data every 10 seconds.
func getData(s *Stats) {
	for {
		*s = Stats{}
		s.getMemoryInfo()
		s.getNetworkInfo()
		time.Sleep(10 * time.Second)
	}
}

func testOut(s *Stats) {
	for {
		log.Println(s)
	}
}

func main() {
	stats := Stats{}
	go getData(&stats)
	tmpl := template.Must(template.ParseFiles("templates/template.html"))
	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, &stats)
	})
	http.ListenAndServe(":8080", nil)
}
