package sniper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sniper/discows"
	"sniper/files"
	"sniper/global"
	"sniper/logger"
	"strings"
	"sync/atomic"
	"time"
)

var (
	//	reGiftLink   = regexp.MustCompile("(?i)(cord.gift/|cord.com/gifts/|promos.discord.gg/|cord.com/billing/promotions/)([a-zA-Z0-9]+)")
	reInviteLink = regexp.MustCompile("(discord.gg/|discord.com/invites/)([0-9a-zA-Z]+)")
)

func checkIfDupeCode(code string) bool {
	for _, _code := range global.DetectedNitros {
		if code == _code {
			return true
		}
	}

	return false
}

func getNitroGift(content string) (has bool, giftId string) {
	var lowContent string = strings.ToLower(content)
	has = strings.Contains(lowContent, "cord.gift") || strings.Contains(lowContent, "promos.discord.gg")
	if has {
		var gift []string = strings.Split(content, "/")
		giftId = strings.Split(strings.Split(gift[len(gift)-1], "\n")[0], " ")[0]
		giftId, _ = strings.CutSuffix(giftId, "#")
	}

	return has, giftId
}

type Sniper struct {
	client *discows.Client
	opened bool
	Token  string
	Loaded bool
	Guilds int
}

func (sniper *Sniper) Init() (err error) {
	sniper.opened = false
	sniper.client = discows.NewClient(sniper.Token,
		global.DiscordBuildNumber,
		sniper.onClose,
		sniper.onReady,
		sniper.onMessageCreate)

	err = sniper.client.Open()
	if err != nil {
		return
	}

	sniper.opened = true
	return
}

func (sniper *Sniper) Close() {
	if !sniper.opened {
		return
	}

	sniper.client.Close()
	sniper.opened = false
}

func (sniper *Sniper) onClose(code int, text string) error {
	if !sniper.opened {
		return nil
	}

	if code == discows.CloseEventCodeAuthenticationFailed.Code {
		atomic.AddUint64(&global.TotalAlts, ^uint64(0))
		atomic.AddUint64(&global.DeadAlts, uint64(1))

		if sniper.Loaded {
			atomic.AddUint64(&global.LoadedAlts, ^uint64(0))
			atomic.AddUint64(&global.LoadedServers, ^uint64(sniper.Guilds-1))
		}

		if tokenFull := global.GetTokenFull(sniper.Token); len(tokenFull) > 0 {
			files.AppendFile("data/dead_alts.txt", tokenFull)
		} else {
			files.AppendFile("data/dead_alts.txt", sniper.Token)
		}

		global.RemoveAltToken(sniper.Token)
	}

	sniper.Close()
	return nil
}

func (sniper *Sniper) onReady(e *discows.ReadyMessage) {
	if !sniper.opened {
		return
	}

	// it RE-LOADED then
	if sniper.Loaded {
		atomic.AddUint64(&global.LoadedAlts, ^uint64(0))
		atomic.AddUint64(&global.LoadedServers, ^uint64(sniper.Guilds-1))
	}

	sniper.Loaded = true
	sniper.Guilds = len(e.Guilds)

	atomic.AddUint64(&global.LoadedAlts, 1)
	atomic.AddUint64(&global.LoadedServers, uint64(len(e.Guilds)))

	//logger.Success("Token ready", logger.FieldString("loaded", fmt.Sprintf("%d/%d", loadedAlts, totalAlts)), logger.FieldInt("guilds", sniper.Guilds), logger.FieldString("username", sniper.Username), logger.FieldString("token", sniper.Token))
}

func (sniper *Sniper) checkIfInviteLink(messageContent string) {
	// if !global.Config.Sniper.SaveInvites {
	// 	return
	// }

	if !reInviteLink.MatchString(messageContent) {
		return
	}

	code := reInviteLink.FindStringSubmatch(messageContent)

	if len(code) < 2 {
		return
	}

	if len(code[2]) < 1 {
		return
	}

	if global.Config.ScrapeInvites {
		global.Invites = append(global.Invites, code[2])
	}

	atomic.AddUint64(&global.FoundInvites, 1)
}

func (sniper *Sniper) checkIfPromocode(giftCode, giftResponse string) {

	if !strings.Contains(giftResponse, `"Payment source required to redeem gift."`) {
		return
	}

	if global.Config.ScrapePomoCodes {
		global.Promocodes = append(global.Promocodes, giftCode)
	}
	atomic.AddUint64(&global.FoundPromocodes, 1)
}

type nitroClaimedStruct struct {
	SubscriptionPlan struct {
		Name string `json:"name"`
	} `json:"subscription_plan"`
	StoreListing struct {
		Sku struct {
			Name string `json:"name"`
		} `json:"sku"`
	} `json:"store_listing"`
}

func (sniper *Sniper) onGiftClaim(startTime time.Time, giftId string, nitroType string, delay string) {
	global.DetectedNitros = append(global.DetectedNitros, giftId)
}

func (sniper *Sniper) onGiftMiss(startTime time.Time, giftId string, delay string) {
	global.DetectedNitros = append(global.DetectedNitros, giftId)
}

func (sniper *Sniper) onMessageCreate(e *discows.DiscordMessage) {
	//logger.Info("message", logger.FieldAny("content", e.Message.Content))

	go func() {
		// if e.Author.ID == s.State.User.ID && e.Content == "!test_claim" {
		// 	go sniper.onGiftClaim("giftID.n-am", "Nitro Monthly", "0.069s")
		// }

		if containsNitro, giftId := getNitroGift(e.Content); containsNitro {
			if len(giftId) >= 16 {
				if !checkIfDupeCode(giftId) {
					var spamIdentifier string = e.GuildID
					if len(spamIdentifier) < 2 {
						spamIdentifier = e.Author.ID
					}

					if !global.SpamDetectorPtr.IsSpam(spamIdentifier) {
						var startTime = time.Now()
						giftData := CheckGiftLink(giftId)
						if !giftData.GotData {
							guildID := "Unknown"
							if len(e.GuildID) > 2 {
								guildID = e.GuildID
							}

							logger.Error("Failed to get gift data (request failed)", logger.FieldString("code", giftId), logger.FieldString("author", e.Author.String()), logger.FieldString("guild_id", guildID))
							return
						}

						var sniperUsername string = "Unknown"
						var authorName string = e.Author.String()
						var guildId string
						var guildName string = "Unknown"

						if sniper.client != nil {
							sniperUsername = sniper.client.Cache.User.String()
						}

						if len(e.GuildID) > 2 {
							guildId = e.GuildID
							if sniper.client != nil {
								if tempGuildName := sniper.client.Cache.GetGuildName(guildId); len(tempGuildName) > 0 {
									guildName = tempGuildName
								}
							}
						} else {
							guildId = "DMs"
							guildName = "DMs"
						}

						timeDiff := giftData.End.Sub(startTime)
						delayFormatted := fmt.Sprintf("%f", timeDiff.Seconds()) + "s"

						switch giftData.StatusCode {
						case 0:
							logger.Error("Error sniping", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("response", giftData.Body))

						case 200:
							var nitroType string = "Unknown"

							var claimResponse nitroClaimedStruct
							if json.Unmarshal([]byte(giftData.Body), &claimResponse) == nil {
								if len(claimResponse.SubscriptionPlan.Name) >= 3 {
									nitroType = claimResponse.SubscriptionPlan.Name
								} else if len(claimResponse.StoreListing.Sku.Name) >= 3 {
									nitroType = claimResponse.StoreListing.Sku.Name
								}
							}

							go WebhookSuccess(giftId, timeDiff, sniperUsername, nitroType, authorName, guildId, guildName)
							go sniper.onGiftClaim(startTime, giftId, nitroType, delayFormatted)

							atomic.AddUint64(&global.TotalClaimed, 1)

							//logger.Success("Claimed Nitro!", logger.FieldString("type", nitroType), logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("author", authorName), logger.FieldString("guild_id", guildId), logger.FieldString("mainToken", global.HideTokenLog(global.SnipingToken)))

						case 400:
							go WebhookFail(giftId, timeDiff, sniperUsername, authorName, guildId, guildName, giftData.Body)
							go sniper.onGiftMiss(startTime, giftId, delayFormatted)

							// only doing that here, on miss
							go sniper.checkIfPromocode(giftId, giftData.Body)

							//logger.Fail("Missed gift", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("author", authorName), logger.FieldString("guild_id", guildId))

							atomic.AddUint64(&global.TotalMissed, 1)

						case 401:
							go WebhookFail(giftId, timeDiff, sniperUsername, authorName, guildId, guildName, giftData.Body)
							go sniper.onGiftMiss(startTime, giftId, delayFormatted)

							//logger.Fail("Unauthorized mainToken", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("author", authorName), logger.FieldString("guild_id", guildId), logger.FieldString("mainToken", global.HideTokenLog(global.SnipingToken)))

							atomic.AddUint64(&global.TotalMissed, 1)

						case 403:
							go WebhookFail(giftId, timeDiff, sniperUsername, authorName, guildId, guildName, giftData.Body)
							go sniper.onGiftMiss(startTime, giftId, delayFormatted)

							//logger.Fail("Account is locked", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("author", authorName), logger.FieldString("guild_id", guildId), logger.FieldString("mainToken", global.HideTokenLog(global.SnipingToken)))

							atomic.AddUint64(&global.TotalMissed, 1)

						case 404:
							go sniper.onGiftMiss(startTime, giftId, delayFormatted)

							//logger.Fail("Unknown gift code", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("author", authorName), logger.FieldString("guild_id", guildId))

							atomic.AddUint64(&global.TotalInvalid, 1)

						case 429:
							go WebhookFail(giftId, timeDiff, sniperUsername, authorName, guildId, guildName, giftData.Body)
							go sniper.onGiftMiss(startTime, giftId, delayFormatted)

							//logger.Fail("Rate limit", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("author", authorName), logger.FieldString("guild_id", guildId))

							atomic.AddUint64(&global.TotalMissed, 1)

						default:
							logger.Error("Unknown snipe status", logger.FieldString("code", giftId), logger.FieldString("delay", delayFormatted), logger.FieldString("sniper", sniperUsername), logger.FieldString("response", giftData.Body))
						}

						// increase attempts
						atomic.AddUint64(&global.TotalAttempts, 1)
					}
				}

				//return
			}
		}

		sniper.checkIfInviteLink(e.Content)

		atomic.AddUint64(&global.FoundMessages, 1)
	}()
}
