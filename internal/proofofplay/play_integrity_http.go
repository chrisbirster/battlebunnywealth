package proofofplay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type AccessTokenSource interface {
	AccessToken(context.Context) (string, error)
}
type StaticAccessToken string

func (t StaticAccessToken) AccessToken(context.Context) (string, error) {
	if strings.TrimSpace(string(t)) == "" {
		return "", errors.New("Play Integrity access token not configured")
	}
	return string(t), nil
}

type PlayIntegrityHTTPDecoder struct {
	Client  *http.Client
	Tokens  AccessTokenSource
	BaseURL string
}

func (d PlayIntegrityHTTPDecoder) DecodeIntegrityToken(ctx context.Context, packageName, token string) (PlayIntegrityVerdict, error) {
	if d.Tokens == nil {
		return PlayIntegrityVerdict{}, errors.New("Play Integrity token source not configured")
	}
	access, err := d.Tokens.AccessToken(ctx)
	if err != nil {
		return PlayIntegrityVerdict{}, err
	}
	client := d.Client
	if client == nil {
		client = http.DefaultClient
	}
	base := strings.TrimRight(d.BaseURL, "/")
	if base == "" {
		base = "https://playintegrity.googleapis.com"
	}
	endpoint := base + "/v1/" + url.PathEscape(packageName) + ":decodeIntegrityToken"
	body, _ := json.Marshal(map[string]string{"integrity_token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return PlayIntegrityVerdict{}, err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return PlayIntegrityVerdict{}, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return PlayIntegrityVerdict{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return PlayIntegrityVerdict{}, fmt.Errorf("Play Integrity decode status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var decoded struct {
		TokenPayloadExternal struct {
			RequestDetails struct {
				RequestPackageName string `json:"requestPackageName"`
				RequestHash        string `json:"requestHash"`
			} `json:"requestDetails"`
			AccountDetails struct {
				AppLicensingVerdict string `json:"appLicensingVerdict"`
			} `json:"accountDetails"`
			AppIntegrity struct {
				AppRecognitionVerdict   string   `json:"appRecognitionVerdict"`
				PackageName             string   `json:"packageName"`
				CertificateSha256Digest []string `json:"certificateSha256Digest"`
			} `json:"appIntegrity"`
			DeviceIntegrity struct {
				DeviceRecognitionVerdict []string `json:"deviceRecognitionVerdict"`
				RecentDeviceActivity     struct {
					DeviceActivityLevel string `json:"deviceActivityLevel"`
				} `json:"recentDeviceActivity"`
			} `json:"deviceIntegrity"`
		} `json:"tokenPayloadExternal"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return PlayIntegrityVerdict{}, fmt.Errorf("decode Play Integrity verdict: %w", err)
	}
	payload := decoded.TokenPayloadExternal
	packageFromVerdict := payload.AppIntegrity.PackageName
	if packageFromVerdict == "" {
		packageFromVerdict = payload.RequestDetails.RequestPackageName
	}
	return PlayIntegrityVerdict{
		RequestHash:           payload.RequestDetails.RequestHash,
		PackageName:           packageFromVerdict,
		CertificateDigests:    payload.AppIntegrity.CertificateSha256Digest,
		AppRecognitionVerdict: payload.AppIntegrity.AppRecognitionVerdict,
		AppLicensingVerdict:   payload.AccountDetails.AppLicensingVerdict,
		DeviceIntegrity:       payload.DeviceIntegrity.DeviceRecognitionVerdict,
		RecentDeviceActivity:  payload.DeviceIntegrity.RecentDeviceActivity.DeviceActivityLevel,
	}, nil
}
