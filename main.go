package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Payload struct {
	Content string
	Name    string
	Proxied bool
	Type    string
	Comment string
	Id      string
	Ttl     int
}

type CloudflareParams struct {
	ZoneId        string `json:"zoneId"`
	DnsRecordId   string `json:"dnsRecordId"`
	ApiEmail      string `json:"apiEmail"`
	ApiKey        string `json:"apiKey"`
	DnsRecordName string `json:"dnsRecordName"`
	Comment       string `json:"comment"`
	Proxied       bool   `json:"proxied"`
	Ttl           int    `json:"ttl"`
}

type Config struct {
	IntervalSeconds  int              `json:"intervalSeconds"`
	CloudflareParams CloudflareParams `json:"cloudflareParams"`
}

func main() {
	lastPublicIp := getip()

	var intervalSeconds, cfTtl int
	var cfProxied, generateConfig, forceFirst bool
	var cfZoneId, cfDnsRecordId, cfApiEmail, cfApiKey, cfDnsRecordName, cfComment, configFilePath string
	flag.IntVar(&intervalSeconds, "interval", 600, "Interval in seconds to check for IP changes")
	flag.StringVar(&configFilePath, "config", "", "Path to the config file")
	flag.BoolVar(&generateConfig, "generate", false, "Generate a config file with the provided parameters")
	flag.BoolVar(&forceFirst, "force", false, "Force the first update to Cloudflare")
	flag.StringVar(&cfZoneId, "zone", "", "Cloudflare Zone ID")
	flag.StringVar(&cfDnsRecordId, "record", "", "Cloudflare DNS Record ID")
	flag.StringVar(&cfApiEmail, "email", "", "Cloudflare API Email")
	flag.StringVar(&cfApiKey, "key", "", "Cloudflare API Key")
	flag.StringVar(&cfDnsRecordName, "name", "", "Cloudflare DNS Record Name")
	flag.StringVar(&cfComment, "comment", "", "Comment about the DNS Record")
	flag.BoolVar(&cfProxied, "proxied", false, "Whether the DNS Record is proxied by Cloudflare or not")
	flag.IntVar(&cfTtl, "ttl", 3600, "TTL of the DNS Record")
	flag.Parse()

	if generateConfig {
		config := Config{
			IntervalSeconds: intervalSeconds,
			CloudflareParams: CloudflareParams{
				ZoneId:        cfZoneId,
				DnsRecordId:   cfDnsRecordId,
				ApiEmail:      cfApiEmail,
				ApiKey:        cfApiKey,
				DnsRecordName: cfDnsRecordName,
				Comment:       cfComment,
				Proxied:       cfProxied,
				Ttl:           cfTtl,
			},
		}

		configJson, err := json.MarshalIndent(config, "", "    ")
		if err != nil {
			fmt.Println(err)
			return
		}

		err = os.WriteFile("config.json", configJson, 0644)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Config file generated successfully")
	}

	if configFilePath != "" {
		config, err := readConfig(configFilePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		intervalSeconds = config.IntervalSeconds
		cfZoneId = config.CloudflareParams.ZoneId
		cfDnsRecordId = config.CloudflareParams.DnsRecordId
		cfApiEmail = config.CloudflareParams.ApiEmail
		cfApiKey = config.CloudflareParams.ApiKey
		cfDnsRecordName = config.CloudflareParams.DnsRecordName
		cfComment = config.CloudflareParams.Comment
		cfProxied = config.CloudflareParams.Proxied
		cfTtl = config.CloudflareParams.Ttl
	}

	for {
		currentPublicIp := getip()
		if lastPublicIp != currentPublicIp || forceFirst {
			if !forceFirst {
				fmt.Println("IP changed: ", lastPublicIp, " -> ", currentPublicIp)
			}

			updateCloudflare(currentPublicIp, cfZoneId, cfDnsRecordId, cfApiEmail, cfApiKey, cfDnsRecordName, cfComment, cfProxied)
			lastPublicIp = currentPublicIp

			forceFirst = false
		}

		time.Sleep(time.Duration(intervalSeconds) * time.Second)
	}
}

func getip() string {
	req, err := http.Get("https://checkip.amazonaws.com/")
	if err != nil {
		return err.Error()
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)

	if err != nil {
		return err.Error()
	}

	return string(body)
}

func readConfig(configFilePath string) (Config, error) {
	var config Config
	file, err := os.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func updateCloudflare(newIp, cfZoneId, cfDnsRecordId, cfApiEmail, cfApiKey, cfDnsRecordName, cfComment string, cfProxied bool) {
	req, err := http.NewRequest("PATCH", fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", cfZoneId, cfDnsRecordId), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Auth-Email", cfApiEmail)
	req.Header.Add("X-Auth-Key", cfApiKey)

	payload := &Payload{
		Type:    "A",
		Name:    cfDnsRecordName,
		Content: newIp,
		Comment: cfComment,
		Ttl:     3600,
		Proxied: cfProxied,
		Id:      cfDnsRecordId,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Body = io.NopCloser(bytes.NewReader(b))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}
