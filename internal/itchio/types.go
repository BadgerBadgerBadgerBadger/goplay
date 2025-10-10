package itchio

import "time"

type GameInfo struct {
	OwnedKey    OwnedKey     `json:"game"`
	GameID      int          `json:"game_id"`
	DownloadKey int          `json:"download_key"`
	Uploads     []GameUpload `json:"uploads"`

	size int
}

func (g GameInfo) Size() int {

	// If size has already been calculated, return it.
	if g.size > 0 {
		return g.size
	}

	for _, upload := range g.Uploads {
		g.size += upload.Size
	}

	return g.size
}

type OwnedKeysResponse struct {
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
	OwnedKeys []OwnedKey `json:"owned_keys"`
}

type OwnedKey struct {
	Game       Game      `json:"game"`
	GameID     int       `json:"game_id"`
	Downloads  int       `json:"downloads"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ID         int       `json:"id"` // This is also the download key for this game.
	PurchaseID int       `json:"purchase_id"`
}

type Game struct {
	PublishedAt    time.Time `json:"published_at"`
	Title          string    `json:"title"`
	ShortText      string    `json:"short_text"`
	MinPrice       int       `json:"min_price"`
	CreatedAt      time.Time `json:"created_at"`
	Classification string    `json:"classification"`
	User           struct {
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
		CoverURL    string `json:"cover_url"`
		ID          int    `json:"id"`
		Username    string `json:"username"`
	} `json:"user"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	ID       int    `json:"id"`
	CoverURL string `json:"cover_url"`
}

type GameUpload struct {
	Position int `json:"position"`
	GameID   int `json:"game_id"`
	Build    struct {
		UserVersion   string    `json:"user_version"`
		UpdatedAt     time.Time `json:"updated_at"`
		Version       int       `json:"version"`
		CreatedAt     time.Time `json:"created_at"`
		ParentBuildID int       `json:"parent_build_id"`
		ID            int       `json:"id"`
		UploadID      int       `json:"upload_id"`
	} `json:"build"`
	CreatedAt   time.Time `json:"created_at"`
	BuildID     int       `json:"build_id"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	Filename    string    `json:"filename"`
	ChannelName string    `json:"channel_name"`
	Size        int       `json:"size"`
	Storage     string    `json:"storage"`
	UpdatedAt   time.Time `json:"updated_at"`
	ID          int       `json:"id"`
}
