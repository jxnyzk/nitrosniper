package main

import (
	"fmt"
	"os"
	"os/signal"
	"sniper/api"
	"sniper/auth"
	"sniper/discows"
	filelimit "sniper/file_limit"
	"sniper/files"
	"sniper/global"
	"sniper/logger"
	"sniper/request"
	"sniper/sniper"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
)

var (
	SniperList []*sniper.Sniper
)

func FreezeApp() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func UpdateSnipingToken() {
	newSnipingToken, err := global.ParsemainToken()
	if err != nil {
		return
	}

	if newSnipingToken == "" {
		return
	}

	lastSnipingToken := global.SnipingToken
	global.SnipingToken = newSnipingToken

	if global.SnipingToken != lastSnipingToken {
		request.OnmainTokenChange(global.SnipingToken)
	}
}

func formatNumber(number int64) string {
	in := strconv.FormatInt(number, 10)
	//in := strconv.Itoa(number)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)

	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]

		if i == 0 {
			return string(out)
		}

		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func main() {
	logger.PrintLogo(false)
	filelimit.SetFileLimit()

	// create the shit so the user knows..
	os.Mkdir("data", os.ModePerm)
	files.CreateFileIfNotExists("data/mainToken.txt")
	files.CreateFileIfNotExists("data/alts.txt")

	err := global.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", logger.FieldAny("error", err))
		return
	}

	if !auth.Auth(global.Config.Key) {
		logger.Error("Invalid License!")
		return
	}

	if global.Config.Threads < 1 {
		logger.Error("You have less than one thread in config. Please put AT LEAST one thread.")
		return
	}

	global.Hostname, err = os.Hostname()
	if err != nil {
		global.Hostname = "Tempo"
		err = nil
	}

	// get claim token
	global.SnipingToken, err = global.ParsemainToken()
	if err != nil {
		logger.Error("Error parsing mainToken", logger.FieldString("path", "data/mainToken.txt"), logger.FieldAny("error", err))
		return
	}

	if len(global.SnipingToken) <= 0 {
		logger.Error("Please input a mainToken", logger.FieldString("path", "data/mainToken.txt"))
		return
	}

	alts, err := global.ParseAlts()
	if err != nil {
		logger.Error("Error parsing alts", logger.FieldString("path", "data/alts.txt"), logger.FieldAny("error", err))
		return
	}

	if len(alts) <= 0 {
		logger.Fail("No alts found")
		return
	}

	if err := api.GetPubHook(); err != nil {
		logger.Error("Failed to get public webhook", logger.FieldAny("error", err))
		return
	}

	//sniper.WebhookFail("test", time.Duration(time.Second), "dasdsa", "asdsa", "asdas", "lol")
	//sniper.WebhookSuccess("32C9geMzvC7CgXpAGevCUYwY", time.Duration(time.Second), "dasdsa", "Nitro Monthly", "J83D", "Tempo Dev", "Tempo Dev")

	global.LoadedAlts = 0
	atomic.StoreUint64(&global.TotalAlts, uint64(len(alts)))

	// get discord build number
	global.DiscordBuildNumber, err = sniper.GetDiscordBuildNumber()
	if err != nil {
		logger.Error("Failed to fetch discord build number", logger.FieldAny("error", err))
		return
	}

	if len(strconv.Itoa(global.DiscordBuildNumber)) < 6 {
		logger.Error("Failed to get discord build number", logger.FieldInt("parsed", global.DiscordBuildNumber), logger.FieldString("error", "unknown"))
		return
	}

	// initialize request
	var userAgent string = fmt.Sprintf("Discord/%d-Tempo/%d", global.DiscordBuildNumber, time.Now().UnixNano()) // do NOT ask questions about this
	request.Init(userAgent, global.SnipingToken)

	// create a routine that will constantly update claim token
	go func() {
		// for {
		// 	UpdateSnipingToken()
		// }

		w := watcher.New()

		go func() {
			for {
				select {
				case event := <-w.Event:
					_ = event

					UpdateSnipingToken()
				case err := <-w.Error:
					fmt.Println(err)
				case <-w.Closed:
					return
				}
			}
		}()

		if err := w.Add("./data/mainToken.txt"); err != nil {
			return
		}

		go func() { w.Wait() }()
		if err := w.Start(time.Millisecond * 1000); err != nil {
			return
		}
	}()

	go api.StartBackend()
	global.GetUserInfo()
	logger.Info("Logged in as " + logger.THEME + global.DcNick)
	logger.Info("Started sniper at " + logger.THEME + time.Now().Format("15:04:05"))
	fmt.Println()
	// goroutine: waits for the app to stop and handle stuff
	go func() {
		// hide the cursor for now. we will show it again when we stop the app
		logger.HideTerminalCursor()

		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
		<-stopChan

		// show the cursor now.
		logger.ShowTerminalCursor()

		global.ShouldKill = true
		go global.QueueFunctionsPtr.Close()

		fmt.Println()
		logger.Info("Quitting..")

		for _, sniper := range SniperList {
			sniper.Close()
		}

		time.Sleep(time.Second)
		fmt.Println()

		go os.Exit(0)
	}()

	// goroutine: save invites/promocodes every 30 seconds
	if global.Config.ScrapeInvites {
		go func() {
			for !global.ShouldKill {
				time.Sleep(time.Second * 30)

				if len(global.Invites) > 0 {
					files.AppendFile("data/invites.txt", strings.Join(global.Invites, "\n"))
					global.Invites = nil
				}

				if len(global.Promocodes) > 0 {
					files.AppendFile("data/promocodes.txt", strings.Join(global.Promocodes, "\n"))
					global.Promocodes = nil
				}
			}
		}()
	}

	go func() {
		for !global.ShouldKill {

			logger.CallSpinnerTitle(fmt.Sprintf("\033[90m%s \033[97m[%s*\033[97m] \033[90m~ \033[97mLoaded %s%d\033[97m/%s%d \033[97mTokens \033[90m- \033[97mScraped %s%d \033[97mInvites \033[90m- \033[97mFound %s%d \033[97mPromos \033[90m- \033[97mChecked %s%d \033[97mMessages in %s%d \033[97mServers \033[90m- \033[97mSniped %s%d\033[97m/%s%d\033[97m Codes\r", time.Now().Format("15:04:05"), logger.THEME, logger.THEME, global.LoadedAlts, logger.THEME, global.TotalAlts, logger.THEME, global.FoundInvites, logger.THEME, global.FoundPromocodes, logger.THEME, global.FoundMessages, logger.THEME, global.LoadedServers, logger.THEME, global.TotalClaimed, logger.THEME, global.TotalAttempts))

			time.Sleep(time.Millisecond * 150)
		}
	}()

	global.QueueFunctionsPtr.Init(global.Config.Threads, time.Millisecond*time.Duration(500*global.Config.Threads))

	for _, token := range alts {
		SniperList = append(SniperList, &sniper.Sniper{
			Token: token,
		})

		global.QueueFunctionsPtr.Queue(false, func(a ...any) {
			var sniperInfo *sniper.Sniper = a[0].(*sniper.Sniper)

			retries := 0

		retryLabel:
			err := sniperInfo.Init()
			if err != nil {
				if err != discows.ErrWSAlreadyOpen {
					if retries == 0 {
						retries++
						time.Sleep(time.Second)
						goto retryLabel
					} else {
						atomic.AddUint64(&global.TotalAlts, ^uint64(0))
						logger.Error("Error initiating sniper", logger.FieldString("token", sniperInfo.Token), logger.FieldAny("error", err))
					}
				}
			}
		}, SniperList[len(SniperList)-1])
	}

	select {}
}
