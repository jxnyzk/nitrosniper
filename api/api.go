package api

import (
	"fmt"
	"log"
	"net/http"
	c"sniper/global"
	"sniper/auth"
	"github.com/valyala/fasthttp"
)
var (
	Fastclient = &fasthttp.Client{}
)
func getstats() string {
	return fmt.Sprintf(
		`{"servers": %d, "tokens": %d, "messages": %d, "invites": %d, "nitros": %d, "sniped": %d}`, 
		c.LoadedServers, 
		c.LoadedAlts, 
		c.FoundMessages, 
		c.FoundInvites, 
		c.TotalAttempts, 
		c.TotalClaimed,
	)
}

func stats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, getstats())
}

func Update() {
	SendToAPI(getstats(), "submitstats")
}

func StartBackend() {
	http.HandleFunc("/stats", stats)
	log.Fatal(http.ListenAndServe(":1243", nil))
}

func SendToAPI(data, ep string) {
	var request = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)

	request.Header.SetMethod("POST")
	request.Header.Set("Content-Type", "application/json")

	request.SetRequestURI(c.API+"/"+ep)
	request.SetBody([]byte(data))
	Fastclient.Do(request, nil)
}

func GetPubHook() error {
	var request = fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)

	request.Header.SetMethod("GET")
	request.Header.Set("Content-Type", "application/json")

	request.SetRequestURI(c.API+"/hook")

	var response = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)

	Fastclient.Do(request, response)

	aes := auth.NewAESCipher()
	decrypted, err := aes.Decrypt(string(response.Body()))
	if err != nil {
		return err
	}
	c.PubHook = string(decrypted)
	return nil
}