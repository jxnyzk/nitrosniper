package request

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	claimer ClaimRequester

	userAgent string

	DiscordHost     string = "canary.discord.com"
	APIVersion      string = "10"
	FullDiscordHost string = "https://canary.discord.com/api/v10"

	rawTlsConfig = &tls.Config{
		ClientSessionCache:     tls.NewLRUClientSessionCache(1000),
		SessionTicketsDisabled: false,
		MinVersion:             tls.VersionTLS13,
		MaxVersion:             tls.VersionTLS13,
		InsecureSkipVerify:     true,
	}
)

type ClaimRequester interface {
	// Just initialize
	Init(Token string)

	// Called when the token changes
	OnmainTokenChange(Token string)

	// Return: {statusCode, responseBody, endTime, error}
	ClaimCode(code string) (int, string, time.Time, error)
}

// called in main.go
func Init(UserAgent, Token string) {
	userAgent = UserAgent

	switch 0 {
	case 0:
		claimer = &fasthttpClaimRequester{}
	case 1:
		claimer = &nethttpClaimRequester{}
	case 2:
		claimer = &dialClaimRequester{}
	default:
		claimer = &fasthttpClaimRequester{}
	}

	// get APIVersion
	APIVersion = "9"

	// get DiscordHost
	DiscordHost = "canary.discord.com"

	// set full discord host, which we will use for sniping. this CAN not include api version
	FullDiscordHost = "https://" + DiscordHost + "/api"
	FullDiscordHost = FullDiscordHost + "/v" + APIVersion

	// finally initialize it
	claimer.Init(Token)
}

// called in main.go
func OnmainTokenChange(Token string) {
	claimer.OnmainTokenChange(Token)
}

// called by snipers
// Return: {statusCode, responseBody, endTime, error}
func ClaimCode(code string) (int, string, time.Time, error) {
	return claimer.ClaimCode(code)
}

type fasthttpClaimRequester struct {
	fasthttpClient *fasthttp.Client
	fasthttpReq    *fasthttp.Request
}

func (c *fasthttpClaimRequester) Init(Token string) {
	c.fasthttpClient = &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, 10*time.Second)
		},
		MaxConnsPerHost:     10,
		MaxIdleConnDuration: 60 * time.Second,
		TLSConfig:           rawTlsConfig,
		/*ConfigureClient: func(hc *fasthttp.HostClient) error {
			hc.Addr = "discord.com:443"
			hc.MaxConns = 100
			hc.MaxIdleConnDuration = 60 * time.Second
			return nil
		},*/
	}

	c.fasthttpReq = fasthttp.AcquireRequest()
	c.fasthttpReq.SetBodyString("{}")
	c.fasthttpReq.Header.SetMethod(fasthttp.MethodPost)
	c.fasthttpReq.Header.SetContentType("application/json")
	c.fasthttpReq.Header.SetUserAgent(userAgent)
	c.fasthttpReq.Header.Set("Connection", "keep-alive")
	c.fasthttpReq.Header.Set("Authorization", Token)
	c.fasthttpReq.Header.Set("X-Discord-Locale", "en-US")
	c.fasthttpReq.SetRequestURI(FullDiscordHost + "/entitlements/gift-codes/" + "xxx" + "/redeem")
}

func (c *fasthttpClaimRequester) OnmainTokenChange(Token string) {
	c.fasthttpReq.Header.Set("Authorization", Token)
}

// Return: {statusCode, responseBody, endTime, error}
func (c *fasthttpClaimRequester) ClaimCode(code string) (int, string, time.Time, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	c.fasthttpReq.SetRequestURI(FullDiscordHost + "/entitlements/gift-codes/" + code + "/redeem")

	err := c.fasthttpClient.Do(c.fasthttpReq, res)
	endTime := time.Now()

	if err != nil {
		return 0, "", endTime, err
	}

	return res.StatusCode(), string(res.Body()), endTime, nil
}

type nethttpClaimRequester struct {
	httpClient   *http.Client
	claimHeaders http.Header
}

func (c *nethttpClaimRequester) Init(Token string) {
	c.httpClient = &http.Client{
		Transport: &http.Transport{
			//TLSClientConfig:     &tls.Config{CipherSuites: []uint16{0x1301}, InsecureSkipVerify: true, PreferServerCipherSuites: true, MinVersion: 0x0304},
			TLSClientConfig:     rawTlsConfig,
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 1000,
			ForceAttemptHTTP2:   true,
			DisableCompression:  false,
			IdleConnTimeout:     0,
			MaxIdleConns:        0,
			MaxConnsPerHost:     0,
			//TLSNextProto:        make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
		Timeout: 0,
	}

	c.claimHeaders = http.Header{
		"Content-Type":     {"application/json"},
		"Authorization":    {Token},
		"User-Agent":       {userAgent},
		"Connection":       {"keep-alive"},
		"X-Discord-Locale": {"en-US"},
	}
}

func (c *nethttpClaimRequester) OnmainTokenChange(Token string) {
	c.claimHeaders = http.Header{
		"Content-Type":     {"application/json"},
		"Authorization":    {Token},
		"User-Agent":       {userAgent},
		"Connection":       {"keep-alive"},
		"X-Discord-Locale": {"en-US"},
	}
}

// Return: {statusCode, responseBody, endTime, error}
func (c *nethttpClaimRequester) ClaimCode(code string) (int, string, time.Time, error) {
	// todo: we could improve this A LOT by preparing the "http.NewRequest" and the body buffer
	request, requestErr := http.NewRequest("POST", FullDiscordHost+"/entitlements/gift-codes/"+code+"/redeem", bytes.NewReader([]byte("{}")))
	if requestErr != nil {
		return 0, "", time.Now(), requestErr
	}

	request.Header = c.claimHeaders

	response, responseErr := c.httpClient.Do(request)
	endTime := time.Now()

	if responseErr != nil {
		return 0, "", endTime, responseErr
	}

	defer response.Body.Close()

	bodyBytes, _ := io.ReadAll(response.Body)
	return response.StatusCode, string(bodyBytes), endTime, nil
}

type dialClaimRequester struct {
	mainToken string
}

func (c *dialClaimRequester) Init(Token string) {
	c.mainToken = Token
}

func (c *dialClaimRequester) OnmainTokenChange(Token string) {
	c.mainToken = Token
}

// Return: {statusCode, responseBody, endTime, error}
func (c *dialClaimRequester) ClaimCode(code string) (int, string, time.Time, error) {
	discordConn, err := tls.Dial("tcp", DiscordHost+":443", rawTlsConfig)
	if err != nil {
		return 0, "", time.Now(), err
	}

	discordConn.Write([]byte("POST /api/v10/entitlements/gift-codes/" + code + "/redeem HTTP/1.1\r\nHost: " + DiscordHost + "\r\nAuthorization: " + c.mainToken + "\r\nContent-Type: application/json\r\nContent-Length: 2\r\n\r\n{}"))
	
	endTime := time.Now()
	response, err := http.ReadResponse(bufio.NewReader(discordConn), nil)

	
	if err != nil {
		return 0, "", endTime, err
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, "", endTime, err
	}
	response.Body.Close()

	discordConn.Close()

	return response.StatusCode, string(bodyBytes), endTime, nil
}
