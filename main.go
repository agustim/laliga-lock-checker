package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	debugMode    bool
	vpnInterface string
	privateKey   string
	publicKey    string
	endpoint     string
	vpnAddress   string
	fwmark       string
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
)

func printDebug(msg string) {
	if debugMode {
		fmt.Println(msg)
	}
}

func getEnvDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {

	err := godotenv.Load()
	if err != nil {
		printDebug("⚠️  No s'ha pogut carregar .env (continuem igualment)")
	}

	inputFile := flag.String("input", getEnvDefault("INPUT_FILE", "sites.json"), "Fitxer d'entrada amb dominis")
	outputFile := flag.String("output", getEnvDefault("OUTPUT_FILE", "resultats.csv"), "Fitxer CSV de sortida")

	flag.StringVar(&vpnInterface, "vpn-interface", getEnvDefault("VPN_INTERFACE", "vpnwg0"), "Nom de la interfície VPN")
	flag.StringVar(&privateKey, "private-key", getEnvDefault("PRIVATE_KEY", "./privatekey"), "Fitxer amb la clau privada")
	flag.StringVar(&publicKey, "public-key", getEnvDefault("PUBLIC_KEY", "publickey="), "Clau pública del peer")
	flag.StringVar(&endpoint, "endpoint", getEnvDefault("ENDPOINT", "example.com:51820"), "Endpoint del peer")
	flag.StringVar(&vpnAddress, "vpn-address", getEnvDefault("VPN_ADDRESS", "10.0.0.1/24"), "Adreça IP de la VPN")
	flag.StringVar(&fwmark, "fwmark", getEnvDefault("FWMARK", "51820"), "fwmark per al routing")

	flag.BoolVar(&debugMode, "debug", false, "Mostra missatges de debug")
	flag.Parse()

	urls, err := loadURLs(*inputFile)
	if err != nil {
		fmt.Println("Error carregant URLs:", err)
		return
	}

	csvFile, err := os.OpenFile(*outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error obrint fitxer CSV:", err)
		return
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	fileInfo, _ := csvFile.Stat()
	if fileInfo.Size() == 0 {
		writer.Write([]string{"hora", "domini", "estat", "latencia_ms"})
	}

	for _, url := range urls {
		printDebug("Provant: " + url)
		status := ""
		latency := int64(0)

		ok, elapsed := checkURLWithLatency(url)
		latency = elapsed.Milliseconds()

		if ok {
			printDebug("✅ Connexió OK: " + url)
			fmt.Println(colorReset + url + ": La liga not blocked")
			status = "not blocked"
		} else {
			printDebug("❌ Error. Comprovant si la VPN està activa...")
			if !isVPNActive() {
				printDebug("🔌 Activant VPN manualment...")
				if err := activateVPN(); err != nil {
					printDebug("❌ Error activant VPN: " + err.Error())
					fmt.Println("❌ Error activant VPN:", err)
					continue
				}
				defer deactivateVPN()
			}

			okVPN, elapsed := checkURLWithLatency(url)
			latency = elapsed.Milliseconds()
			if okVPN {
				printDebug("🔁 ✅ Connexió OK amb VPN: " + url)
				fmt.Println(colorRed + url + ": La liga blocked" + colorReset)
				status = "blocked"
			} else {
				printDebug("🔁 ❌ Encara no funciona: " + url)
				fmt.Println(url + ": KO")
				status = "no response"
			}
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		writer.Write([]string{now, url, status, fmt.Sprintf("%d", latency)})
	}
}

func loadURLs(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var urls []string
	err = json.Unmarshal(data, &urls)
	return urls, err
}

func checkURLWithLatency(url string) (bool, time.Duration) {
	client := http.Client{Timeout: 5 * time.Second}
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}
	start := time.Now()
	resp, err := client.Get(url)
	elapsed := time.Since(start)
	if err != nil {
		return false, elapsed
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, elapsed
}

func isVPNActive() bool {
	out, err := exec.Command("wg", "show", vpnInterface).Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "interface: "+vpnInterface)
}

func activateVPN() error {
	commands := [][]string{
		{"ip", "link", "add", vpnInterface, "type", "wireguard"},
		{"ip", "address", "add", vpnAddress, "dev", vpnInterface},
		{"ip", "link", "set", "mtu", "1420", "up", "dev", vpnInterface},
		{"wg", "set", vpnInterface,
			"fwmark", fwmark,
			"private-key", privateKey,
			"peer", publicKey,
			"endpoint", endpoint,
			"allowed-ips", "0.0.0.0/0",
			"persistent-keepalive", "25"},
		{"ip", "route", "add", "0.0.0.0/0", "dev", vpnInterface, "table", fwmark},
		{"ip", "rule", "add", "not", "fwmark", fwmark, "table", fwmark},
		{"ip", "rule", "add", "table", "main", "suppress_prefixlength", "0"},
		{"sysctl", "-w", "net.ipv4.conf.all.src_valid_mark=1"},
	}
	for _, cmd := range commands {
		printDebug("🔧 Executo: " + strings.Join(cmd, " "))
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("error executant: %v", err)
		}
	}
	return nil
}

func deactivateVPN() error {
	commands := [][]string{
		{"ip", "rule", "del", "table", "main", "suppress_prefixlength", "0"},
		{"ip", "rule", "del", "not", "fwmark", fwmark, "table", fwmark},
		{"ip", "route", "del", "0.0.0.0/0", "dev", vpnInterface, "table", fwmark},
		{"ip", "link", "del", vpnInterface},
	}
	for _, cmd := range commands {
		printDebug("🔧 Executo: " + strings.Join(cmd, " "))
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			return fmt.Errorf("error executant: %v", err)
		}
	}
	return nil
}
