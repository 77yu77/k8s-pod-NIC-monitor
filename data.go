package main

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	Name string
}

type LinkPair struct {
	Pod1       string
	Pod2       string
	Rbandwidth string
	Tbandwidth string
	Laterncy   string
	RBFR       string
	TBFR       string
}

const (
	vNICSymbol = "10.233"
	LinkSymbol = "link"
	Test       = "eth0"
	SrcSymbol  = "src"
)

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	name := "LinkStatus"
	linkTypes := []string{}
	linkTypes = append(linkTypes, "networkName", "star1", "star2", "Rbandwidth", "Tbandwidth", "laterncy", "RBFR", "TBFR")
	metrics := GetTopology()
	result := fmt.Sprintf("# HELP %s data structure\n# TYPE %s counter\n", m.Name, m.Name)
	fmt.Fprint(w, result)
	for _, metric := range metrics {
		data := []string{}
		data = append(data, m.Name, metric.Pod1, metric.Pod2, metric.Rbandwidth, metric.Tbandwidth, metric.Laterncy, metric.RBFR, metric.TBFR)
		if s, err := GeneratePromData(name, linkTypes, data); err == nil {
			fmt.Fprint(w, s)
		}
	}

}

func GetNICTraffic(NIC string) (string, string, string, string) {
	// cmd := exec.Command("sh", "-c", fmt.Sprintf("/flow/data.sh %s", NIC))
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $2}'", NIC))
	stdout, _ := cmd.CombinedOutput()
	receive_bytes_pre, _ := strconv.Atoi(strings.Fields(string(stdout))[0])

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $4}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	receive_errors_bytes_pre, _ := strconv.Atoi(strings.Fields(string(stdout))[0])

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $10}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	transmit_bytes_pre, _ := strconv.Atoi(strings.Fields(string(stdout))[0])

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $12}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	transmit_errors_bytes_pre, _ := strconv.Atoi(strings.Fields(string(stdout))[0])
	time.Sleep(time.Second)

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $2}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	receive_bytes, _ := strconv.Atoi(strings.Fields(string(stdout))[0])

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $4}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	receive_errors_bytes, _ := strconv.Atoi(strings.Fields(string(stdout))[0])

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $10}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	transmit_bytes, _ := strconv.Atoi(strings.Fields(string(stdout))[0])

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s | sed 's/:/ /g' | awk '{print $12}'", NIC))
	stdout, _ = cmd.CombinedOutput()
	transmit_errors_bytes, _ := strconv.Atoi(strings.Fields(string(stdout))[0])
	print(receive_bytes, " ", receive_bytes_pre, " ", transmit_bytes, " ", transmit_bytes_pre, "\n")
	receive_errors_percent := 0
	transmit_errors_percent := 0
	receive_rate := (receive_bytes - receive_bytes_pre) / 1000
	transmit_rate := (transmit_bytes - transmit_bytes_pre) / 1000
	if receive_rate != 0 {
		receive_errors_percent = (receive_errors_bytes - receive_errors_bytes_pre) / receive_rate
	}
	if transmit_rate != 0 {
		transmit_errors_percent = (transmit_errors_bytes - transmit_errors_bytes_pre) / transmit_rate
	}
	r_rate := strconv.Itoa(receive_rate) + "KB"
	t_rate := strconv.Itoa(transmit_rate) + "KB"
	r_e_rate := strconv.Itoa(receive_errors_percent) + "%"
	t_e_rate := strconv.Itoa(transmit_errors_percent) + "%"
	return r_rate, t_rate, r_e_rate, t_e_rate
}

func GetLinkLaterncy(target string) string {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ping -c 1 -i 0.2 %s", target))
	stdout, _ := cmd.CombinedOutput()
	outStr := string(stdout)
	print("\n")
	print(outStr)
	outLines := strings.Split(outStr, "\n")
	var laterncy string
	for _, line := range outLines {
		if strings.Contains(line, "round-trip") {
			words := strings.Split(line, "/")
			laterncy = words[3] + "ms"
		}
	}
	if laterncy == "" {
		return "-"
	}
	return laterncy
}

func GetTopology() []LinkPair {
	cmd := exec.Command("sh", "-c", "echo $PODNAME")
	stdout, _ := cmd.CombinedOutput()
	podName := string(stdout)
	print(podName)
	cmd = exec.Command("sh", "-c", "ip route")
	stdout, _ = cmd.CombinedOutput()
	outStr := string(stdout)
	outLines := strings.Split(outStr, "\n")
	linkPairs := make([]LinkPair, 0)
	for _, line := range outLines {
		if len(line) != 0 {
			if strings.Contains(line, LinkSymbol) && strings.Contains(line, SrcSymbol) {
				words := strings.Fields(line)
				if words[2] != Test {
					r_rate, t_rate, r_e_rate, t_e_rate := GetNICTraffic(words[2])
					print(r_rate, " ", t_rate, " ", r_e_rate, " ", t_e_rate, "\n")
					linkPairs = append(linkPairs, LinkPair{Pod1: podName, Pod2: words[2], Rbandwidth: r_rate, Tbandwidth: t_rate, Laterncy: GetLinkLaterncy(words[6]), RBFR: r_e_rate, TBFR: t_e_rate})
				}
			}
		}
	}
	return linkPairs
}

//	func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//		name := "StarStatus"
//		types := []string{}
//		types = append(types, "name", "timestamp", "CPU", "memory", "disk", "status")
//		data := []string{}
//		data = append(data, "node1", fmt.Sprint(time.Now().Unix()), "0.01%", "100MB", "200MB", "up")
//		if s, err := GeneratePromData(name, types, data); err == nil {
//			fmt.Fprint(w, s)
//		}
//	}
func main() {
	metric := Metrics{Name: "Network1"}
	http.Handle("/metrics", metric)
	http.ListenAndServe(":2112", nil)
}

func GeneratePromData(name string, types []string, datas []string) (string, error) {
	var result string
	if len(types) != len(datas) {
		return result, errors.New("lens of types is difference from lens of datas for prometheus type")
	}
	result += name + "{"
	for index := range types {
		result += fmt.Sprintf("%s=\"%s\",", types[index], datas[index])
	}
	result = result[:len(result)-1] + "} 1\n"
	return result, nil
}

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pod_NIC_exporter data.go
