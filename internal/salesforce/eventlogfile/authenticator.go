package eventlogfile

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"newrelic/multienv/pkg/config"
	"newrelic/multienv/pkg/connect"
	"os"
	"strings"
)

type salesforceAuthenticator struct {
	tokenUrl     string
	clientId     string
	clientSecret string
	username     string
	password     string
}

type accessTokenResponse struct {
	Id          string
	IssuedAt    string `json:"issued_at"`
	InstanceUrl string `json:"instance_url"`
	Signature   string
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func NewAuthenticatorFromConfig(
	pipeConfig *config.PipelineConfig,
	instanceUrl string,
) (*salesforceAuthenticator, error) {
	var ok bool

	clientId := os.Getenv("SF_API_CLIENT_ID")
	if clientId == "" {
		if clientId, ok = pipeConfig.GetString("clientId"); !ok {
			return nil, fmt.Errorf("missing salesforce API client ID")
		}
	}

	clientSecret := os.Getenv("SF_API_CLIENT_SECRET")
	if clientSecret == "" {
		if clientSecret, ok = pipeConfig.GetString("clientSecret"); !ok {
			return nil, fmt.Errorf("missing salesforce API client secret")
		}
	}

	username := os.Getenv("SF_API_CLIENT_USERNAME")
	if username == "" {
		if username, ok = pipeConfig.GetString("username"); !ok {
			return nil, fmt.Errorf("missing salesforce API username")
		}
	}

	password := os.Getenv("SF_API_CLIENT_PASSWORD")
	if password == "" {
		if password, ok = pipeConfig.GetString("password"); !ok {
			return nil, fmt.Errorf("missing salesforce API password")
		}
	}

	return &salesforceAuthenticator{
		fmt.Sprintf("%s/services/oauth2/token", instanceUrl),
		clientId,
		clientSecret,
		username,
		password,
	}, nil
}

func (s *salesforceAuthenticator) Authenticate(
	httpConfig *connect.HttpConfig,
	req *http.Request,
) error {
	client := &http.Client{}

	params := url.Values{
		"grant_type":    {"password"},
		"client_id":     {s.clientId},
		"client_secret": {s.clientSecret},
		"username":      {s.username},
		"password":      {s.password},
	}

	authReq, err := http.NewRequest("POST", s.tokenUrl, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	authReq.Header.Add("Accept", "application/json")
	authReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	authResp, err := client.Do(authReq)
	if err != nil {
		return err
	}

	defer authResp.Body.Close()

	if authResp.StatusCode != 200 {
		return fmt.Errorf("fetch results failed: %s", authResp.Status)
	}

	body, err := io.ReadAll(authResp.Body)
	if err != nil {
		return err
	}

	tokenResponse := &accessTokenResponse{}

	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))

	fmt.Print(tokenResponse.AccessToken)
	return nil
}
