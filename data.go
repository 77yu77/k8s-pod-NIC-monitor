package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
}

type NICMessage struct {
	receive_rate         int
	receive_packet_rate  int
	receive_errors_rate  int
	receive_drop_rate    int
	transmit_rate        int
	transmit_packet_rate int
	transmit_errors_rate int
	transmit_drop_rate   int
}

const (
	vNICSymbol                    = "Link encap"
	receiveRateName               = "ReceiveRate"
	receivePacketRateName         = "ReceivePacketRate"
	receiveErrorsRateName         = "ReceiveErrorsRate"
	receiveDropRateName           = "ReceiveDropRate"
	transmitRateName              = "transmitRate"
	transmitPacketRateName        = "transmitPacketRate"
	transmitErrorsRateName        = "transmitErrorsRate"
	transmitDropRateName          = "transmitDropRate"
	receiveRateDescription        = "The receive data rate(bytes per second) of pod NIC"
	receivePacketRateDescription  = "The receive packets rate(packet num per second) of pod NIC"
	receiveErrorsRateDescription  = "The receive errors rate(errors num per second) of pod NIC"
	receiveDropRateDescription    = "The receive drop packets rate(drop packets num per second) of pod NIC"
	transmitRateDescription       = "The transmit data rate(bytes per second) of pod NIC"
	transmitPacketRateDescription = "The transmit packet rate(packet num per second) of pod NIC"
	transmitErrorsRateDescription = "The transmit errors rate(errors num per second) of pod NIC"
	transmitDropRateDescription   = "The transmit drop packets rate(drop packets num per second) of pod NIC"
)

func GeneratePromData(name string, NIC string, data int) string {
	result := fmt.Sprintf("%s{NICName=\"%s\"} %d\n", name, NIC, data)
	return result
}

func GenerateDescribeMessage(name string, description string) string {
	result := fmt.Sprintf("# HELP %s %s\n# TYPE %s counter\n", name, description, name)
	return result
}

func (m Metrics) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// get NIC names of pod
	cmd := exec.Command("sh", "-c", "ifconfig")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("run ifconfig cmd error:%s", err.Error())
		return
	}
	outLines := strings.Split(string(stdout), "\n")
	NIC_names := []string{}
	for _, line := range outLines {
		if len(line) != 0 && strings.Contains(line, vNICSymbol) {
			words := strings.Fields(line)
			NIC_names = append(NIC_names, words[0])
		}
	}

	// get all NIC message
	NIC_messages := make(map[string]NICMessage, 0)
	for _, NIC_name := range NIC_names {
		NIC_messages[NIC_name] = GetNICMessage(NIC_name)
	}

	// send NIC message
	// receive data rate
	fmt.Fprint(w, GenerateDescribeMessage(receiveRateName, receiveRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(receiveRateName, NIC_name, NIC_messages[NIC_name].receive_rate))
	}

	// receive packet rate
	fmt.Fprint(w, GenerateDescribeMessage(receivePacketRateName, receivePacketRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(receivePacketRateName, NIC_name, NIC_messages[NIC_name].receive_packet_rate))
	}

	// receive error rate
	fmt.Fprint(w, GenerateDescribeMessage(receiveErrorsRateName, receiveErrorsRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(receiveErrorsRateName, NIC_name, NIC_messages[NIC_name].receive_errors_rate))
	}

	// receive drop packet rate
	fmt.Fprint(w, GenerateDescribeMessage(receiveDropRateName, receiveDropRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(receiveDropRateName, NIC_name, NIC_messages[NIC_name].receive_drop_rate))
	}

	// transmit data rate
	fmt.Fprint(w, GenerateDescribeMessage(transmitRateName, transmitRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(transmitRateName, NIC_name, NIC_messages[NIC_name].transmit_rate))
	}

	// transmit packet rate
	fmt.Fprint(w, GenerateDescribeMessage(transmitPacketRateName, transmitPacketRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(transmitPacketRateName, NIC_name, NIC_messages[NIC_name].transmit_packet_rate))
	}

	// transmit errors rate
	fmt.Fprint(w, GenerateDescribeMessage(transmitErrorsRateName, transmitErrorsRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(transmitErrorsRateName, NIC_name, NIC_messages[NIC_name].transmit_errors_rate))
	}

	// transmit drop packet rate
	fmt.Fprint(w, GenerateDescribeMessage(transmitDropRateName, transmitDropRateDescription))
	for _, NIC_name := range NIC_names {
		fmt.Fprint(w, GeneratePromData(transmitDropRateName, NIC_name, NIC_messages[NIC_name].transmit_drop_rate))
	}

}

func GetNICMessage(NIC string) NICMessage {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s", NIC))
	stdout, _ := cmd.CombinedOutput()
	message_pre := strings.Fields(string(stdout))
	receive_bytes_pre, _ := strconv.Atoi(message_pre[1])
	receive_packet_pre, _ := strconv.Atoi(message_pre[2])
	receive_errors_pre, _ := strconv.Atoi(message_pre[3])
	receive_drop_pre, _ := strconv.Atoi(message_pre[4])
	transmit_bytes_pre, _ := strconv.Atoi(message_pre[9])
	transmit_packet_pre, _ := strconv.Atoi(message_pre[10])
	transmit_errors_pre, _ := strconv.Atoi(message_pre[11])
	transmit_drop_pre, _ := strconv.Atoi(message_pre[12])
	time.Sleep(time.Second)

	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat /proc/net/dev | grep %s", NIC))
	stdout, _ = cmd.CombinedOutput()
	message := strings.Fields(string(stdout))
	receive_bytes, _ := strconv.Atoi(message[1])
	receive_packet, _ := strconv.Atoi(message[2])
	receive_errors, _ := strconv.Atoi(message[3])
	receive_drop, _ := strconv.Atoi(message[4])
	transmit_bytes, _ := strconv.Atoi(message[9])
	transmit_packet, _ := strconv.Atoi(message[10])
	transmit_errors, _ := strconv.Atoi(message[11])
	transmit_drop, _ := strconv.Atoi(message[12])
	NIC_message := NICMessage{
		receive_rate:         receive_bytes - receive_bytes_pre,
		receive_packet_rate:  receive_packet - receive_packet_pre,
		receive_errors_rate:  receive_errors - receive_errors_pre,
		receive_drop_rate:    receive_drop - receive_drop_pre,
		transmit_rate:        transmit_bytes - transmit_bytes_pre,
		transmit_packet_rate: transmit_packet - transmit_packet_pre,
		transmit_errors_rate: transmit_errors - transmit_errors_pre,
		transmit_drop_rate:   transmit_drop - transmit_drop_pre,
	}

	return NIC_message
}

func main() {
	metric := Metrics{}
	http.Handle("/metrics", metric)
	http.ListenAndServe(":2112", nil)
}

// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pod_NIC_exporter data.go
