package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	httpClient = &http.Client{}
)

// GenerateAuth generates an authentication token for the user
func (u *User) GenerateAuth(username string, password string) (string, error) {
	urlToken := "https://mobile.vodafone.com.eg/auth/realms/vf-realm/protocol/openid-connect/token"
	dataToken := map[string]string{
		"username":      username,
		"password":      password,
		"grant_type":    "password",
		"client_secret": "a2ec6fff-0b7f-4aa4-a733-96ceae5c84c3",
		"client_id":     "my-vodafone-app",
	}

	headersToken := map[string]string{
		"Accept":                  "application/json, text/plain, */*",
		"Connection":              "keep-alive",
		"x-agent-operatingsystem": "10.1.0.264C185",
		"clientId":                "AnaVodafoneAndroid",
		"x-agent-device":          "HWDRA-MR",
		"x-agent-version":         "2022.1.2.3",
		"x-agent-build":           "500",
		"Content-Type":            "application/x-www-form-urlencoded",
		"User-Agent":              "okhttp/4.9.1",
	}

	form := url.Values{}
	for key, value := range dataToken {
		form.Add(key, value)
	}

	req, err := http.NewRequest("POST", urlToken, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating token request: %v", err)
	}

	for key, value := range headersToken {
		req.Header.Set(key, value)
	}

	respToken, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making token request: %v", err)
	}
	defer respToken.Body.Close()

	if respToken.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", respToken.StatusCode)
	}

	if err := json.NewDecoder(respToken.Body).Decode(&TokenResponse); err != nil {
		return "", fmt.Errorf("error decoding token response: %v", err)
	}

	return TokenResponse.AccessToken, nil
}

func GetPromotion(username string) error {
	urlProduct := "https://web.vodafone.com.eg/services/dxl/promo/promotion?@type=Promo&$.context.type=rechargeProgram"
	headersProduct := map[string]string{
		"Host":               "web.vodafone.com.eg",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Chromium";v="124", "Android WebView";v="124", "Not-A.Brand";v="99"`,
		"msisdn":             username,
		"Accept-Language":    "AR",
		"sec-ch-ua-mobile":   "?1",
		"Authorization":      "Bearer " + TokenResponse.AccessToken,
		"User-Agent":         "vodafoneandroid",
		"Content-Type":       "application/json",
		"Accept":             "application/json",
		"clientId":           "WebsiteConsumer",
		"channel":            "APP_PORTAL",
		"sec-ch-ua-platform": `"Android"`,
		"Sec-Fetch-Site":     "same-origin",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://web.vodafone.com.eg/portal/bf/rechargeProgram",
		"Accept-Encoding":    "gzip, deflate, br, zstd",
	}

	reqProduct, err := http.NewRequest("GET", urlProduct, nil)
	if err != nil {
		return fmt.Errorf("error creating product request: %v", err)
	}

	for key, value := range headersProduct {
		reqProduct.Header.Set(key, value)
	}

	respProduct, err := httpClient.Do(reqProduct)
	if err != nil {
		return fmt.Errorf("error making product request: %v", err)
	}
	defer respProduct.Body.Close()

	return nil
}

func GetUserDataConsumption(token string, phoneNumber string) error {
	urlUsage := "https://web.vodafone.com.eg/services/dxl/usage/usageConsumptionReport?@type=adslWallet&relatedParty.id=20504356294"

	headersUsage := map[string]string{
		"Host":               "web.vodafone.com.eg",
		"Connection":         "keep-alive",
		"sec-ch-ua":          `"Chromium";v="124", "Android WebView";v="124", "Not-A.Brand";v="99"`,
		"Accept-Language":    "AR",
		"sec-ch-ua-mobile":   "?1",
		"msisdn":             phoneNumber,
		"Authorization":      "Bearer " + token,
		"User-Agent":         "vodafoneandroid",
		"Content-Type":       "application/json",
		"Accept":             "application/json",
		"clientId":           "WebsiteConsumer",
		"channel":            "APP_PORTAL",
		"sec-ch-ua-platform": `"Android"`,
		"Sec-Fetch-Site":     "same-origin",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
		"Referer":            "https://web.vodafone.com.eg/spa/adslManagement",
	}

	reqUsage, err := http.NewRequest("GET", urlUsage, nil)
	if err != nil {
		return fmt.Errorf("error creating usage request: %v", err)
	}

	for key, value := range headersUsage {
		reqUsage.Header.Set(key, value)
	}

	// TODO: refine the error handling
	respUsage, err := httpClient.Do(reqUsage)
	if err != nil {
		return fmt.Errorf("error making usage request: %v", err)
	}
	defer respUsage.Body.Close()

	if respUsage.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", respUsage.StatusCode)
	}

	body, err := io.ReadAll(respUsage.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var usageConsumptionRes []UsageConsumptionResponse
	err = json.Unmarshal(body, &usageConsumptionRes)
	if err != nil {
		return fmt.Errorf("error unmarhsalling response body %v", err)
	}

	accountType := usageConsumptionRes[0].Type
	remainingBalance := usageConsumptionRes[0].Bucket[0].BucketBalance[0].RemainingValue

	remainingDataUnits := usageConsumptionRes[1].Bucket[0].BucketBalance[0].RemainingValue.Amount
	dataPlan := fmt.Sprintf("%s - %s",
		usageConsumptionRes[1].Bucket[0].Product[0].Name,
		usageConsumptionRes[1].Bucket[0].Product[0].Type,
	)
	renewalFees := usageConsumptionRes[1].Bucket[1].BucketBalance[0].RemainingValue.Amount
	subscriptionDay, err := strconv.ParseInt(usageConsumptionRes[1].Bucket[0].Product[0].ID, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing subscription day: %v", err)
	}
	var active bool
	if usageConsumptionRes[1].Bucket[1].Product[0].Type == "Active" {
		active = true
	}

	fmt.Println(subscriptionDay)
	renewalDay := time.Unix(subscriptionDay*1000, 0).UTC().Local().Day()

	usageConsumption := UsageConsumption{
		Type:             accountType,
		RemainingBalance: remainingBalance.Amount,
		adsl: Adsl{
			DataPlan:           dataPlan,
			RemainingDataUsage: remainingDataUnits / 1024,
			Active:             active,
			RenewalFees:        renewalFees / 100,
			RenewalDay:         renewalDay,
		},
	}

	j, _ := json.MarshalIndent(usageConsumption, "", "  ")
	k, _ := json.MarshalIndent(usageConsumption.adsl, "", "  ")
	fmt.Println(string(j))
	fmt.Println(string(k))

	return nil
}

var TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type UsageConsumptionResponse struct {
	ID     string `json:"id,omitempty"`
	Bucket []struct {
		BucketBalance []struct {
			RemainingValue struct {
				Amount float32 `json:"amount"`
				Units  string  `json:"units"`
			} `json:"remainingValue"`
			Type string `json:"@type"`
		} `json:"bucketBalance"`
		BucketCounter []struct {
			Value struct {
				Amount float32 `json:"amount"`
			} `json:"value"`
			Type string `json:"@type"`
		} `json:"bucketCounter"`
		Product []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"@baseType"`
		} `json:"product"`
	} `json:"bucket"`
	RelatedParty struct {
		Type         string `json:"@type"`
		ReferredType string `json:"@referredType"`
	} `json:"relatedParty,omitempty"`
	Type string `json:"@type"`
}

type UsageConsumption struct {
	RemainingBalance float32 `json:"remainingBalance"`
	Type             string  `json:"type"`
	adsl             Adsl    `json: "adsl"`
}
type Adsl struct {
	DataPlan           string  `json:"dataPlan"`
	RenewalFees        float32 `json:"renewalFees"`
	RemainingDataUsage float32 `json:"remainingDataUsage"`
	Active             bool    `json:"active"`
	RenewalDay         int     `json:"renewalDay"`
}
