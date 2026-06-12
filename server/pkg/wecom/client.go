package wecom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiBase = "https://qyapi.weixin.qq.com"

type Client struct {
	CorpID      string
	AgentID     string
	Secret      string
	token       string
	tokenUntil  time.Time
	tokenLocker sync.Mutex
}

func New(corpID, agentID, secret string) *Client {
	return &Client{CorpID: strings.TrimSpace(corpID), AgentID: strings.TrimSpace(agentID), Secret: strings.TrimSpace(secret)}
}

func (c *Client) Enabled() bool {
	return c != nil && c.CorpID != "" && c.AgentID != "" && c.Secret != ""
}

func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if !c.Enabled() {
		return "", fmt.Errorf("wecom client is not configured")
	}

	c.tokenLocker.Lock()
	defer c.tokenLocker.Unlock()

	if c.token != "" && time.Now().Before(c.tokenUntil) {
		return c.token, nil
	}

	endpoint := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", apiBase, url.QueryEscape(c.CorpID), url.QueryEscape(c.Secret))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var payload struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("wecom gettoken http status %d", response.StatusCode)
	}
	if payload.ErrCode != 0 || strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("wecom gettoken failed: %s", payload.ErrMsg)
	}

	expiresIn := payload.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200
	}
	c.token = payload.AccessToken
	c.tokenUntil = time.Now().Add(time.Duration(expiresIn-120) * time.Second)
	return c.token, nil
}

func (c *Client) SendTextCard(ctx context.Context, toUser, title, description, linkURL string) error {
	if !c.Enabled() {
		return fmt.Errorf("wecom client is not configured")
	}
	toUser = strings.TrimSpace(toUser)
	if toUser == "" {
		return fmt.Errorf("wecom toUser is empty")
	}

	agentID, err := strconv.Atoi(strings.TrimSpace(c.AgentID))
	if err != nil || agentID <= 0 {
		return fmt.Errorf("invalid wecom agent id")
	}

	token, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]any{
		"touser":  toUser,
		"msgtype": "textcard",
		"agentid": agentID,
		"textcard": map[string]any{
			"title":       title,
			"description": description,
			"url":         linkURL,
			"btntxt":      "查看订单",
		},
		"enable_duplicate_check": 0,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", apiBase, url.QueryEscape(token))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("wecom send message http status %d", response.StatusCode)
	}
	if payload.ErrCode != 0 {
		return fmt.Errorf("wecom send textcard failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}
	return nil
}

func (c *Client) SendText(ctx context.Context, toUser, content string) error {
	if !c.Enabled() {
		return fmt.Errorf("wecom client is not configured")
	}

	toUser = strings.TrimSpace(toUser)
	if toUser == "" {
		return fmt.Errorf("wecom toUser is empty")
	}

	agentID, err := strconv.Atoi(strings.TrimSpace(c.AgentID))
	if err != nil || agentID <= 0 {
		return fmt.Errorf("invalid wecom agent id")
	}

	token, err := c.AccessToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]any{
		"touser":  toUser,
		"msgtype": "text",
		"agentid": agentID,
		"text": map[string]any{
			"content": content,
		},
		"enable_duplicate_check": 0,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", apiBase, url.QueryEscape(token))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var payload struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("wecom send text http status %d", response.StatusCode)
	}

	if payload.ErrCode != 0 {
		return fmt.Errorf("wecom send text failed: %d %s", payload.ErrCode, payload.ErrMsg)
	}

	return nil
}
