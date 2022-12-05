package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

func main() {

	siteUrl := "https://api.myanimelist.net/v2/users/LavaCorDeRosa/animelist?limit=100&offset=300"

	req, _ := http.NewRequest("GET", siteUrl, nil)

	req.Header.Add("X-MAL-CLIENT-ID", "325ae8cd00c25128a5791e1422c9e0eb")

	p, _ := url.Parse("http://45.233.196.210")

	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(p),
		},
	}

	res, err := client.Do(req)
	if err != nil {
		logrus.Fatalln(err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))

}
