package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"badgerbadgerbadgerbadger.dev/goplay/internal/util"
)

type SlashCommand struct {
	Token               string `schema:"token"`
	Command             string `schema:"command"`
	Text                string `schema:"text"`
	ResponseURL         string `schema:"response_url"`
	TriggerID           string `schema:"trigger_id"`
	UserId              string `schema:"user_id"`
	UserName            string `schema:"user_name"`
	TeamID              string `schema:"team_id"`
	ChannelID           string `schema:"channel_id"`
	ChannelName         string `schema:"channel_name"`
	APIAppID            string `schema:"api_app_id"`
	IsEnterpriseInstall bool   `schema:"is_enterprise_install"`
	TeamDomain          string `schema:"team_domain"`
}

type ResponseTpe string

type CommandReply struct {
	ResponseType    ResponseTpe `json:"response_type"`
	Text            string      `json:"text"`
	ReplaceOriginal bool        `json:"replace_original"`
	DeleteOriginal  bool        `json:"delete_original"`
}

const (
	ephemeral ResponseTpe = "ephemeral"
	inChannel ResponseTpe = "in_channel"
)

type Slack struct {
	config SlackConfig
}

func NewSlack(config SlackConfig) Slack {
	return Slack{
		config: config,
	}
}

type OauthResponse struct {
	Ok         bool       `json:"ok"`
	Error      *string    `json:"error"`
	AuthedUser AuthedUser `json:"authed_user"`
}

type AuthedUser struct {
	ID          string `json:"id"`
	Scope       string `json:"scope"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (s *Slack) Authenticate(code string) error {
	authReq, err := http.NewRequest("GET", "https://slack.com/api/oauth.v2.access", nil)
	util.Must(err, "failed to create new request")

	authReqQuery := authReq.URL.Query()

	authReqQuery.Set("code", code)
	authReqQuery.Set("client_id", s.config.Oauth.ClientID)
	authReqQuery.Set("client_secret", s.config.Oauth.ClientSecret)

	authReq.URL.RawQuery = authReqQuery.Encode()

	res, err := (&http.Client{}).Do(authReq)
	defer res.Body.Close()

	if err != nil {
		return errors.Wrap(err, "failed to call slack oauth")
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("oauth responded with non-ok status")
	}

	oauthResp := OauthResponse{}
	err = json.NewDecoder(res.Body).
		Decode(&oauthResp)
	if err != nil {
		return errors.Wrapf(err, "failed to read oauth response body")
	}

	if !oauthResp.Ok {
		return errors.New(*oauthResp.Error)
	}

	err = StoreAuthedUser(oauthResp.AuthedUser.ID, oauthResp.AuthedUser)
	if err != nil {
		return errors.Wrap(err, "failed to save authed user")
	}

	log.Infof("%s\n", oauthResp)

	return nil
}

func (s *Slack) Rant(sc SlashCommand) error {

	upperCased := strings.ToUpper(sc.Text)
	tripleExclaimed := strings.ReplaceAll(upperCased, "!", "!!!")
	questionExclaimed := strings.ReplaceAll(tripleExclaimed, "?", "?!")

	rant := fmt.Sprintf(
		"<@%s>: %s",
		sc.UserId,
		questionExclaimed,
	)

	reply := CommandReply{
		Text:           rant,
		ResponseType:   inChannel,
		DeleteOriginal: true,
	}
	replyBody, err := json.Marshal(reply)
	if err != nil {
		return errors.Wrap(err, "failed to marshall reply")
	}

	req, err := http.NewRequest(http.MethodPost, sc.ResponseURL, bytes.NewReader(replyBody))
	if err != nil {
		return errors.Wrap(err, "failed to create req")
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send http req")
	}
	log.Infof("reply req: %+v", resp)

	return nil
}
