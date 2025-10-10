package itchio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const pageSize = 50

type Client interface {
	OwnedGames(withUploads bool) <-chan GameItemsResult
	Uploads(gameId, downloadKey int) ([]GameUpload, error)
	DownloadGame(uploadId, downloadKey int) (io.ReadCloser, error)
}

type apiKeyClient struct {
	concurrency int
	httpClient  *http.Client
	apiKey      string
}

func (a *apiKeyClient) DownloadGame(uploadId, downloadKey int) (io.ReadCloser, error) {

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.itch.io/uploads/%d/download?download_key_id=%d", uploadId, downloadKey),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Set("Transfer-Encoding", "chunked")
	a.decorateRequest(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.Errorf("got bad status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (a *apiKeyClient) Uploads(gameId, downloadKey int) ([]GameUpload, error) {

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.itch.io/games/%d/uploads?download_key_id=%d", gameId, downloadKey),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	a.decorateRequest(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("got bad status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response")
	}

	type uploadsResp struct {
		Uploads []GameUpload `json:"uploads"`
	}
	var uploads uploadsResp
	err = json.Unmarshal(body, &uploads)

	// When no uploads are present, the API, very annoyingly,
	// returns and empty object instead of an empty array 😤
	if string(body) == `{"uploads":{}}` {
		return nil, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to decode response for %d, %d", gameId, downloadKey))
	}

	return uploads.Uploads, nil
}

type GameItemsResult struct {
	GameInfo GameInfo
	Err      error
}

// OwnedGames fetches all the keys owned by the user and returns them on a channel.
// If an error occurs, it will be sent on the error channel. The error channel will
// be closed when the owned keys channel is closed.
func (a *apiKeyClient) OwnedGames(withUploads bool) <-chan GameItemsResult {

	gameItemChan := make(chan GameItemsResult)

	sendErrAndClose := func(err error) {
		gameItemChan <- GameItemsResult{
			Err: err,
		}
		close(gameItemChan)
	}

	go func() {

		page := 1

		for {

			req, err := http.NewRequest(
				"GET",
				fmt.Sprintf("https://api.itch.io/profile/owned-keys?page=%d", page),
				nil,
			)
			if err != nil {
				sendErrAndClose(errors.Wrap(err, "failed to create request"))
				break
			}
			a.decorateRequest(req)

			resp, err := a.httpClient.Do(req)
			if err != nil {
				sendErrAndClose(errors.Wrap(err, "failed to make request"))
				break
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				sendErrAndClose(errors.Errorf("got bad status code %d", resp.StatusCode))
				break
			}

			ownedKeysResp := OwnedKeysResponse{}
			err = json.NewDecoder(resp.Body).Decode(&ownedKeysResp)
			if err != nil {
				sendErrAndClose(errors.Wrap(err, "failed to decode response"))
				break
			}

			for _, ownedKey := range ownedKeysResp.OwnedKeys {

				gameDetails := GameInfo{
					OwnedKey:    ownedKey,
					GameID:      ownedKey.GameID,
					DownloadKey: ownedKey.ID,
				}

				if withUploads {
					uploads, err := a.Uploads(ownedKey.GameID, ownedKey.ID)
					if err != nil {
						gameItemChan <- GameItemsResult{
							Err: err,
						}
						continue
					}
					gameDetails.Uploads = uploads
				}

				gameItemChan <- GameItemsResult{
					GameInfo: gameDetails,
				}
			}

			if len(ownedKeysResp.OwnedKeys) < pageSize {
				break
			}
			page += 1
		}

		close(gameItemChan)
	}()

	return gameItemChan
}

func (a *apiKeyClient) decorateRequest(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
}

func NewApiKeyClient(apiKey string) Client {

	return &apiKeyClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}
