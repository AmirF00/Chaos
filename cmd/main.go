package main

import (
	"encoding/json"
	"fmt"
	"github.com/SharokhAtaie/gitPush"
	sUtils "github.com/projectdiscovery/utils/slice"
	"io/ioutil"
	"net/http"
)

type Program struct {
	Programs []struct {
		Name    string   `json:"name"`
		URL     string   `json:"url"`
		Bounty  bool     `json:"bounty"`
		Swag    bool     `json:"swag"`
		Domains []string `json:"domains"`
	} `json:"programs"`
}

func main() {
	// Get data from Chaos URL
	body := GetRequest("https://raw.githubusercontent.com/projectdiscovery/public-bugbounty-programs/main/chaos-bugbounty-list.json")
	var newChaos Program
	err := json.Unmarshal(body, &newChaos)
	CheckErr(err)

	data, err := ioutil.ReadFile("/home/hunter/Chaos/chaos-bugbounty-list.json")
	CheckErr(err)

	var oldChaos Program
	err = json.Unmarshal(data, &oldChaos)
	CheckErr(err)

	data2, err := json.Marshal(newChaos)
	CheckErr(err)

	compare := Compare(oldChaos, newChaos)

	if len(compare) > 0 {
		err = ioutil.WriteFile("../chaos-bugbounty-list.json", data2, 0644)
		CheckErr(err)

		push.Run("../.", "update data")
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

func Compare(old, new Program) []string {
	var oldData []string
	var newData []string

	for _, v := range old.Programs {
		for _, a := range v.Domains {
			oldData = append(oldData, a)
		}
	}

	for _, v := range new.Programs {
		for _, a := range v.Domains {
			newData = append(newData, a)
		}
	}
	test, _ := sUtils.Diff(newData, oldData)
	return test
}

