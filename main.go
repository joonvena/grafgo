package main

import (
	"log"
	"net/http"
	"os/exec"
	"strings"
	"text/template"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

type Stats struct {
	MemoryInfo  []Memory
	NetworkInfo []Network
	CPUInfo     []CPU
	GraphData   []GraphData
	Uptime      string
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

type CPU struct {
	Model string
}

type GraphData struct {
	CPULoad []CPULoad
}

type CPULoad struct {
	Timestamp string
	Load      float64
}

var graphdata []GraphData

// Helper function to convert Kb -> Mb
func convertToMB(kb uint64) uint64 {
	return kb / 1000
}

// Helper function to get time to format that Chart.js accepts
func getTime() string {
	timeNow := time.Now()
	formatTime := timeNow.Format("15-04-05")
	// Remove dashes from time
	time := strings.Replace(string(formatTime), "-", "", 2)
	return time
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

func (s *Stats) getUptime() error {
	cmd := exec.Command("uptime", "-p")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	uptime := strings.Replace(string(out), "up ", "", 1)
	s.Uptime = uptime
	return nil
}

func (s *Stats) getCPUInfo() error {
	stat, err := linuxproc.ReadCPUInfo("/proc/cpuinfo")
	if err != nil {
		log.Println(err)
	}
	for _, cpu := range stat.Processors {
		cpu := CPU{Model: cpu.ModelName}
		s.CPUInfo = append(s.CPUInfo, cpu)
	}
	load, err := linuxproc.ReadLoadAvg("/proc/loadavg")
	if err != nil {
		log.Println(err)
	}
	time := getTime()
	loadValue := CPULoad{Timestamp: time, Load: load.Last1Min}
	graphvalue := GraphData{CPULoad: []CPULoad{loadValue}}
	graphdata = append(graphdata, graphvalue)
	s.GraphData = append(s.GraphData, graphdata...)
	return nil
}

// TODO
func (s *Stats) getDiskInfo() error {
	stat, err := linuxproc.ReadDisk("/dev/sda")
	if err != nil {
		log.Println(err)
	}
	log.Println(stat.Free)
	return nil
}

// Get data every 10 seconds.
func getData(s *Stats) {
	for {
		*s = Stats{}
		s.getMemoryInfo()
		s.getNetworkInfo()
		s.getUptime()
		s.getCPUInfo()
		s.getDiskInfo()
		time.Sleep(10 * time.Second)
	}
}

func main() {
	stats := Stats{}
	go getData(&stats)
	tmpl := template.Must(template.ParseFiles("templates/template.html"))
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, &stats)
	})
	http.ListenAndServe(":8080", nil)
}
