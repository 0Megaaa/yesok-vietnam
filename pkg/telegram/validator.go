package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"yesok-vietnam/config"
)

// InitData 代表 Telegram Web App 传来的用户信息
type InitData struct {
	QueryID        string    `json:"query_id"`
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Language       string    `json:"language_code"`
	AllowsWriteAcc bool      `json:"allows_write_to_pm"`
	AuthDate       time.Time `json:"auth_date"`
	Hash           string    `json:"hash"`
}

type Validator struct {
	botToken string
}

func NewValidator() *Validator {
	return &Validator{
		botToken: config.Global.TG.BotToken,
	}
}

// Validate 校验 Telegram 传来的 initData 是否真实且未过期
func (v *Validator) Validate(initDataStr string, maxAge time.Duration) (*InitData, error) {
	// 1. 解析 URL 编码的数据
	vals, err := url.ParseQuery(initDataStr)
	if err != nil {
		return nil, err
	}

	// 2. 提取并移除 hash
	hash := vals.Get("hash")
	if hash == "" {
		return nil, errors.New("missing hash")
	}
	vals.Del("hash")

	// 3. 按字母顺序排序键名并构建 data-check-string
	var keys []string
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var dataCheckArr []string
	for _, k := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, vals.Get(k)))
	}
	dataCheckStr := strings.Join(dataCheckArr, "\n")

	// 4. 计算 HMAC-SHA-256 签名
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(v.botToken))

	mac := hmac.New(sha256.New, secretKey.Sum(nil))
	mac.Write([]byte(dataCheckStr))
	calculatedHash := hex.EncodeToString(mac.Sum(nil))

	// 5. 校验签名是否一致
	if calculatedHash != hash {
		return nil, errors.New("invalid signature")
	}

	// 6. 校验时间戳是否过期
	authDateInt, _ := strconv.ParseInt(vals.Get("auth_date"), 10, 64)
	authDate := time.Unix(authDateInt, 0)
	if time.Since(authDate) > maxAge {
		return nil, errors.New("initData expired")
	}

	// 7. 解析 User 的 JSON 数据
	var data InitData
	data.Hash = hash
	data.AuthDate = authDate
	data.QueryID = vals.Get("query_id")

	userJSON := vals.Get("user")
	if userJSON != "" {
		var u struct {
			ID           int64  `json:"id"`
			Username     string `json:"username"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			LanguageCode string `json:"language_code"`
		}
		// 解析嵌套的 JSON
		if err := json.Unmarshal([]byte(userJSON), &u); err == nil {
			data.UserID = u.ID
			data.Username = u.Username
			data.FirstName = u.FirstName
			data.LastName = u.LastName
			data.Language = u.LanguageCode
		}
	}

	return &data, nil
}
