package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	fUtils "github.com/projectdiscovery/utils/file"
	sUtils "github.com/projectdiscovery/utils/slice"
	"io/ioutil"
	"net/http"
	"time"
)

type webhookPayload struct {
	Contents string `json:"content"`
}

type Program struct {
	Programs []struct {
		Name    string   `json:"name"`
		URL     string   `json:"url"`
		Bounty  bool     `json:"bounty"`
		Swag    bool     `json:"swag"`
		Domains []string `json:"domains"`
	} `json:"programs"`
}

type Output struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
}

var (
	authHeader    = "gsl46xLLEgchaDQuxUwokwaKc5dge10or8nq2Qk5V21z1lFuZGMWDA5dIH0iFgE9"
	maxSubdomains = 999999
)

func main() {
	oldDomains := GetDomains()

	oldSubs := fUtils.FileExists("../oldSubdomains.json")
	subdomains, _ := GetSubdomain(oldDomains)
	if !oldSubs {
		SaveStructToFile(subdomains)
		return
	}
	output := CompareSubdomain(subdomains.Subdomains)
	SendMessage(output)
	SaveStructToFile(subdomains)
}

// GetDomains collect Domains (bounty:true) from input collection
func GetDomains() []string {
	var domains []string
	data := GetRequest("https://github.com/AmirF00/Chaos/raw/main/chaos-bugbounty-list.json")

	var oldChaos Program
	err := json.Unmarshal(data, &oldChaos)
	CheckErr(err)

	for _, v := range oldChaos.Programs {
		if v.Bounty == true {
			for _, a := range v.Domains {
				domains = append(domains, a)
			}
		}
	}
	return domains
}

func CompareSubdomain(newSubDomain []string) map[string][]string {
	data, err := ioutil.ReadFile("../oldSubdomains.json")
	CheckErr(err)

	var output Output
	err = json.Unmarshal(data, &output)
	CheckErr(err)
	result := make(map[string][]string)
	comparedSubDomains, _ := sUtils.Diff(newSubDomain, output.Subdomains)
	if len(comparedSubDomains) > 0 {
		result[output.Domain] = comparedSubDomains
	}
	return result
}

// GetSubdomain get a http request to project discovery endpoit to get all subdomains then insert to database
func GetSubdomain(domains []string) (Output, error) {
	var result Output
	skipDomains := []string{"swisscom.ch", "telenet.be", "telenor.se", "bredbandsbolaget.se", "tumblr.com", "aol.com", "zendesk.com"}
	for _, domain := range domains {
		select {
		case <-context.Background().Done():
			return result, context.Background().Err()
		default:
		}

		time.Sleep(100 * time.Millisecond)
		if contains(skipDomains, domain) {
			continue
		}

		req, err5 := http.NewRequestWithContext(context.Background(), "GET", "https://dns.projectdiscovery.io/dns/"+domain+"/subdomains", nil)
		if err5 != nil {
			return result, err5
		}
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err4 := client.Do(req)
		if err4 != nil {
			return result, err4
		}

		defer resp.Body.Close()

		data, err3 := ioutil.ReadAll(resp.Body)
		if err3 != nil {
			return result, err3
		}

		var output Output
		if err2 := json.Unmarshal(data, &output); err2 != nil {
			return result, err2
		}

		result = output
	}
	return result, nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func SendMessage(input map[string][]string) {
	url := "https://discord.com/api/webhooks/1095101052535193743/XaS9ZBncEG-n14qQAKMF74Y8386NgUNFNazm-3qdTeLHqxC8kmAYzj5PX8n_HyYatnRR"
	for domain, subdomains := range input {
		content := fmt.Sprintf("**%v**\n\n```NewAssets*:\n", domain)
		for _, asset := range subdomains {
			content += asset + "." + domain + "\n"
		}
		content += "```"

		postData, err := json.Marshal(webhookPayload{Contents: content})
		if err != nil {
			fmt.Println(err)
		}

		_, err = http.Post(url, "application/json", bytes.NewBuffer(postData))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func CheckErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func GetRequest(url string) []byte {
	resp, err := http.Get(url)
	CheckErr(err)
	defer resp.Body.Close()

	data, err2 := ioutil.ReadAll(resp.Body)
	CheckErr(err2)
	return data
}

func SaveStructToFile(subdomains Output) {
	jsonSubs, _ := json.Marshal(subdomains)
	err := ioutil.WriteFile("../oldSubdomains.json", jsonSubs, 0644)
	CheckErr(err)
}

