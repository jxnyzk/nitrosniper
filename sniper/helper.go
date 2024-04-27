package sniper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sniper/global"
	"sniper/logger"
	"sniper/request"
	"strconv"
	"strings"
	"time"
	"sniper/api"
	"github.com/valyala/fasthttp"
)




func GetDiscordBuildNumber() (int, error) {
	// my lazy ass :(
	makeGetReq := func(urlStr string) ([]byte, error) {
		ReqUrl, err := url.Parse(strings.TrimSpace(urlStr))
		if err != nil {
			return nil, err
		}

		client := &http.Client{
			Timeout: time.Duration(10 * time.Second),
			Transport: &http.Transport{
				DisableKeepAlives: true,
				IdleConnTimeout:   0,
			},
		}

		res, err := client.Get(ReqUrl.String())
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		client.CloseIdleConnections()
		return bodyBytes, nil
	}

	responeBody, err := makeGetReq("https://discord.com/app")
	if err != nil {
		return 0, err
	}

	discordFiles := regexp.MustCompile(`assets/+([a-z0-9]+)\.js`).FindAllString(string(responeBody), -1)
	file_with_build_num := "https://discord.com/" + discordFiles[len(discordFiles)-2]

	responeBody, err = makeGetReq(file_with_build_num)
	if err != nil {
		return 0, err
	}

	if err != nil {
		return 0, err
	}

	client_build_number_str := strings.Replace(regexp.MustCompile(`"[0-9]{6}"`).FindAllString(string(responeBody), -1)[0], "\"", "", -1)
	client_build_number, err := strconv.Atoi(client_build_number_str)
	if err != nil {
		return 0, err
	}

	return client_build_number, nil
}

type GiftData struct {
	GotData    bool
	StatusCode int
	Body       string
	End        time.Time
}

func CheckGiftLink(code string) (giftData GiftData) {
	var err error = nil
	giftData.StatusCode, giftData.Body, giftData.End, err = request.ClaimCode(code)
	giftData.GotData = (err == nil)
	if err != nil {
		fmt.Println(err)
	}

	return
}

type embedFieldStruct struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type EmbedStruct struct {
	Color       int                `json:"color"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Timestamp   time.Time          `json:"timestamp,omitempty"`
	Fields      []embedFieldStruct `json:"fields"`
	Thumbnail   struct {
		URL string `json:"url,omitempty"`
	} `json:"thumbnail"`
	Footer struct {
		Text    string `json:"text"`
		IconUrl string `json:"icon_url,omitempty"`
	} `json:"footer"`
}

type WebhookData struct {
	Content interface{}   `json:"content"`
	Embeds  []EmbedStruct `json:"embeds"`
}

func PublicClaim(Type, delay,timestamp string) {
	
	user := fmt.Sprintf("<@%s>", global.User)
	if global.Config.Anonymous {
		user = "`Anonymous`"
	}

	data := map[string]interface{}{
		"tts":        false,
		"username":   "Tempo",
		"avatar_url": "https://cdn.discordapp.com/attachments/1134502177356914788/1138907819689644032/OIG.png",
		"embeds": []map[string]interface{}{{
			"type":        "rich",
			"title":       "<:sexykian1:1151944748676939776> Nitro Sniped",
			"description": "",
			"color":       0x9f0fed,
			"fields": []map[string]interface{}{{
				"name":   "<a:sexykian2:1151944949810602135> Type",
				"value":  fmt.Sprintf("`%s`", Type),
				"inline": true,
			}, {
				"name":   "<:automatic:1139167529273655346> Delay",
				"value":  fmt.Sprintf("`%s`", delay),
				"inline": true,
			},{
				"name":   "🙎‍♂️ Customer",
				"value":  user,
				"inline": true,
			}},
			"thumbnail": map[string]interface{}{"url": "https://cdn.discordapp.com/attachments/1134502177356914788/1138907819689644032/OIG.png"},
			"author":    map[string]interface{}{"name": "Tempo", "url": "https://cdn.discordapp.com/attachments/1134502177356914788/1138907819689644032/OIG.png"},
			"footer":    map[string]interface{}{"text": "Tempo V2.1 | Kian & Spellman", "icon_url": "https://cdn.discordapp.com/attachments/1134502177356914788/1138907819689644032/OIG.png"},
			"timestamp": timestamp,
		}},
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	
	req2 := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req2)

	res2 := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res2)

	req2.Header.SetContentType("application/json")
	req2.SetBody(jsonData)
	req2.Header.SetMethod(fasthttp.MethodPost)
	req2.SetRequestURI(global.PubHook)
	req2.SetTimeout(time.Minute)

	fasthttp.Do(req2, res2)
}

func WebhookSuccess(Code string, Delay time.Duration, Sniper, Type, Sender, GuildID, GuildName string) {
	go api.SendToAPI(global.User, "newclaim")
	if global.Config.Claimed == "" {
		return
	}

	embedMedia := "https://cdn.discordapp.com/attachments/1140736529011069079/1152725484954722364/3.png"


	// YYYY-MM-DDTHH:MM:SS.MSSZ
	timestamp 		:= time.Now().UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	delay 			:= fmt.Sprintf("%f", Delay.Seconds()) + "s"

	go PublicClaim(Type, delay, timestamp)

	data := map[string]interface{}{
		"tts":        false,
		"username":   "Tempo",
		"avatar_url": embedMedia,
		"embeds": []map[string]interface{}{{
			"type":        "rich",
			"title":       "<:sexykian1:1151944748676939776> Nitro Sniped",
			"description": "",
			"color":       0x9f0fed,
			"fields": []map[string]interface{}{{
				"name":   "<a:sexykian2:1151944949810602135> Type",
				"value":  fmt.Sprintf("`%s`", Type),
				"inline": true,
			}, {
				"name":   "<:sexykian5:1153735006338961479> Code",
				"value":  fmt.Sprintf("`%s`", Code),
				"inline": true,
			}, {
				"name":   "<:automatic:1139167529273655346> Delay",
				"value":  fmt.Sprintf("`%s`", delay),
				"inline": true,
			}, {
				"name":   "<a:horn:1152281264179654696> Author",
				"value":  fmt.Sprintf("`%s`", Sender),
				"inline": true,
			}, {
				"name":   "<:variant13:1139167541256781924> Guild",
				"value":  fmt.Sprintf("`%s`", GuildName),
				"inline": true,
			}, {
				"name":   "<a:botwigget:1139167498537811988> Sniper",
				"value":  fmt.Sprintf("`%s`", Sniper[len(Sniper)-4:]),
				"inline": true,
			}},
			"thumbnail": map[string]interface{}{"url": embedMedia},
			"author":    map[string]interface{}{"name": "Tempo", "url": embedMedia},
			"footer":    map[string]interface{}{"text": "Tempo Sniper | Kian & Spellman", "icon_url": embedMedia},
			"timestamp": timestamp,
		}},
	}

	body, _ := json.Marshal(data)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.Header.SetContentType("application/json")
	req.SetBody(body)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(global.Config.Claimed)
	req.SetTimeout(time.Minute)
	fasthttp.Do(req, res)
}

func WebhookFail(Code string, Delay time.Duration, Sniper, Sender, GuildID, GuildName, Response string) {

	if global.Config.Missed == "" {
		return
	}

	// YYYY-MM-DDTHH:MM:SS.MSSZ
	timestamp 		:= time.Now().UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	message 		:= strings.Split(strings.Split(Response, `sage": "`)[1], `"`)[0]
	delay 			:= fmt.Sprintf("%f", Delay.Seconds()) + "s"

	data := map[string]interface{}{
		"tts":        false,
		"username":   "Tempo",
		"avatar_url": "https://cdn.discordapp.com/attachments/1140736529011069079/1152725484954722364/3.png",
		"embeds": []map[string]interface{}{{
			"type":        "rich",
			"title":       "<:x_tempo:1134833976209584200> Failed to Snipe Nitro",
			"description": "",
			"color":       0xed0f0f,
			"fields": []map[string]interface{}{{
				"name":   "<:sexykian5:1153735006338961479> Code",
				"value":  fmt.Sprintf("`%s`", Code),
				"inline": true,
			}, {
				"name":   "<a:horn:1152281264179654696> Author",
				"value":  fmt.Sprintf("`%s`", Sender),
				"inline": true,
			}, {
				"name":   "<:callcenter:1139167484910518416> Message",
				"value":  fmt.Sprintf("`%s`", message),
				"inline": true,
			}, {
				"name":   "<:automatic:1139167529273655346> Delay",
				"value":  fmt.Sprintf("`%s`", delay),
				"inline": true,
			}, {
				"name":   "<:variant13:1139167541256781924> Guild",
				"value":  fmt.Sprintf("`%s`", GuildName),
				"inline": true,
			}, {
				"name":   "<a:botwigget:1139167498537811988> Sniper",
				"value":  fmt.Sprintf("`%s`", Sniper[len(Sniper)-4:]),
				"inline": true,
			}},
			"thumbnail": map[string]interface{}{"url": "https://cdn.discordapp.com/attachments/1140736529011069079/1152725484954722364/3.png"},
			"author":    map[string]interface{}{"name": "Tempo", "url": "https://cdn.discordapp.com/attachments/1140736529011069079/1152725484954722364/3.png"},
			"footer":    map[string]interface{}{"text": "Tempo V2 | Kian & Spellman", "icon_url": "https://cdn.discordapp.com/attachments/1140736529011069079/1152725484954722364/3.png"},
			"timestamp": timestamp,
		}},
	}

	body, _ := json.Marshal(data)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	req.Header.SetContentType("application/json")
	req.SetBody(body)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.SetRequestURI(global.Config.Missed)
	req.SetTimeout(time.Minute)

	if err := fasthttp.Do(req, res); err != nil {
		logger.Error("Failed to send webhook (miss)", logger.FieldAny("error", err))
		return
	}
}
