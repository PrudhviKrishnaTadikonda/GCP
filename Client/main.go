package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const signerUrl = "https://image-censor.ue.r.appspot.com/sign"

func getSignedURL(target string, values url.Values) (string, error) {
	resp, err := http.PostForm(target, values)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func main() {
	u, err := getSignedURL(signerUrl, url.Values{"content_type": {"image/png"}, "ext": {"png"}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Signed URL here: %q\n", u)

	b, err := ioutil.ReadFile("heartwithscissor.png")
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("PUT", u, bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "image/png")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)
}
