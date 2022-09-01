package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var execGlobalOutput string = fmt.Sprintf("[+] PID %d\n", os.Getpid())
var entryptType string = ""

func toBase64(str string) string {
	str_bytes := []byte(str)
	str_b64 := base64.StdEncoding.EncodeToString(str_bytes)
	return str_b64
}

func toHex(str string) string {
	str_bytes := []byte(str)
	str_hex := hex.EncodeToString(str_bytes)
	return str_hex
}

func toBase64Hex(str string) string {
	str_b64hex := toHex(toBase64(str))
	return str_b64hex
}

func getOutputContinually(name string, args ...string) {
	cmd := exec.Command(name, args...)
	closed := make(chan struct{})
	defer close(closed)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	defer stdoutPipe.Close()

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			data := string(scanner.Bytes())
			if err != nil {
				fmt.Println("transfer error with bytes:", scanner.Bytes())
				continue
			}
			execGlobalOutput += string(data + "\n")
			//fmt.Printf("%s\n", string(data))
		}
	}()
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func balabala(writer http.ResponseWriter, request *http.Request) {
	var data string

	switch entryptType {
	case "base64":
		data = toBase64(execGlobalOutput)
	case "hex":
		data = toHex(execGlobalOutput)
	case "base64hex":
		data = toBase64Hex(execGlobalOutput)
	default:
		data = execGlobalOutput
	}
	writer.Write([]byte(data))
}

func httpserver(ip string, port string) {
	http.HandleFunc("/", balabala)
	fmt.Printf("[*] Start HTTP server on %s:%s\n", ip, port)
	err := http.ListenAndServe(ip+":"+port, nil)
	if err != nil {
		fmt.Println("http server error...")
		return
	}
}

func main() {
	var (
		ip      string
		port    string
		command string
		encrypt string
		maxtime int
	)
	flag.StringVar(&ip, "ip", "0.0.0.0", "IP")
	flag.StringVar(&port, "port", "443", "Port")
	flag.StringVar(&command, "cmd", "whoami", "Command to exec")
	flag.StringVar(&encrypt, "encrypt", "", "Encrypt type, support: base64, hex, base64hex")
	flag.IntVar(&maxtime, "time", 20, "Set alive max time(seconds), program will exit after this seconds. Use 0 to make program always alive")
	flag.Parse()

	commandArgs := strings.Fields(command)
	entryptType = encrypt
	go httpserver(ip, port)
	go getOutputContinually(commandArgs[0], commandArgs[1:]...)
	if maxtime == 0 {
		for {
			time.Sleep(5 * time.Second)
		}
	} else {
		time.Sleep(time.Duration(maxtime) * time.Second)
	}

}
