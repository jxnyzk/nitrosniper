package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sniper/logger"
	"sniper/global"
)

var (
	res    Res
	c      = &http.Client{}
)

func Auth(key string) bool {
	hwid, err := GetCpuID()
	cipher := NewAESCipher()
	if err != nil {
		fmt.Println("Error getting hwid")
		return false
	}
	js := fmt.Sprintf(`{"license": "%s", "hwid": "%s", "program": "Tempo"}`, key, hwid)
	enc, err := cipher.Encrypt(js)
	if err != nil {
		//logger.Error().Str("ID", "43671").Msg("Error! Report to Kian")
		logger.Error("Error! Report to Kian", logger.FieldString("ID", "43671"))
		return false
	}

	req, err := http.NewRequest("POST", "https://auth.spellman.vip:443/hwid", strings.NewReader(enc))
	if err != nil {
		//Logger.Error().Str("ID", "54245").Msg("Error! Report to Kian")
		logger.Error("Error! Report to Kian", logger.FieldString("ID", "54245"))
		return false
	}

	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		//Logger.Error().Str("ID", "54233").Msg("Error! Report to Kian")
		logger.Error("Error! Report to Kian", logger.FieldString("ID", "54233"))
		return false
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//Logger.Error().Str("ID", "54533").Msg("Error! Report to Kian")
		logger.Error("Error! Report to Kian", logger.FieldString("ID", "54533"))
		return false
	}

	dec, err := cipher.Decrypt(string(body))
	if err != nil {
		//Logger.Error().Str("ID", "34786").Msg("Error! Report to Kian")
		logger.Error("Error! Report to Kian", logger.FieldString("ID", "34786"))
		return false
	}
	
	if err := json.Unmarshal([]byte(dec), &res); err != nil {
		//Logger.Error().Str("ID", "51234").Msg("Error! Report to Kian")
		logger.Error("Error! Report to Kian", logger.FieldString("ID", "51234"))
		return false
	}

	if !res.Suc {
		
		return false
	}
	
	global.User = res.User
	return true
}
