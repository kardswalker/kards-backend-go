package models

// These structs are aligned with Dumper-7 SDK headers under kards/Public.
// They are intended as protocol/domain DTOs for KARDS-compatible payloads.

type SDKPlayerProfile struct {
	PlayerName string   `json:"player_name"`
	PlayerTag  string   `json:"player_tag"`
	Avatar     string   `json:"avatar"`
	Medals     []string `json:"medals"`
	Ribbons    []string `json:"ribbons"`
	CacheTime  string   `json:"cacheTime"`
}

type SDKLaunchMessage struct {
	Title        string `json:"title"`
	Body         string `json:"Body"`
	URL          string `json:"URL"`
	URLText      string `json:"url_text"`
	CloseText    string `json:"close_text"`
	Icon         string `json:"icon"`
	DownloadInfo string `json:"download_info"`
}

type SDKNewPlayerLoginReward struct {
	Day     int    `json:"day"`
	Reset   string `json:"Reset"`
	Seconds int    `json:"Seconds"`
}

type SDKPlayersInfoMisc struct {
	MobileReviewDone     bool     `json:"bMobileReviewDone"`
	CreateDate           string   `json:"CreateDate"`
	OptedOutProfile      bool     `json:"bOptedOutOfProfile"`
	OptedOutMarketing    int      `json:"bOptedOutOfMarketing"`
	FeaturedAchievements []string `json:"featuredAchievements"`
}

type SDKPlayersInfo struct {
	HeartbeatURL string `json:"heartbeat_url"`
	PlayerID     int    `json:"player_id"`
	PlayerName   string `json:"player_name"`
	PlayerTag    int    `json:"player_tag"`

	Gold            int `json:"Gold"`
	Diamonds        int `json:"Diamonds"`
	Dust            int `json:"Dust"`
	DraftAdmissions int `json:"draft_admissions"`

	Packs            string `json:"Packs"`
	URL              string `json:"URL"`
	LibraryURL       string `json:"library_url"`
	DecksURL         string `json:"decks_url"`
	PacksURL         string `json:"packs_url"`
	DailyMissionsURL string `json:"dailymissions_url"`
	AchievementsURL  string `json:"achievements_url"`

	TutorialsFinished []string `json:"tutorials_finished"`
	SeasonEnd         string   `json:"season_end"`
	Stars             int      `json:"stars"`

	IsOfficer      bool   `json:"is_officer"`
	HasBeenOfficer bool   `json:"has_been_officer"`
	Email          string `json:"Email"`

	EmailRewardReceived bool `json:"email_reward_received"`
	EmailVerified       bool `json:"email_verified"`

	LaunchMessages       []SDKLaunchMessage      `json:"launch_messages"`
	NewCards             []string                `json:"new_cards"`
	NewPlayerLoginReward SDKNewPlayerLoginReward `json:"new_player_login_reward"`

	UsaXP     int `json:"usa_xp"`
	GermanyXP int `json:"germany_xp"`
	SovietXP  int `json:"soviet_xp"`
	JapanXP   int `json:"japan_xp"`
	BritainXP int `json:"britain_xp"`

	UsaLevel     int `json:"usa_level"`
	GermanyLevel int `json:"germany_level"`
	SovietLevel  int `json:"soviet_level"`
	JapanLevel   int `json:"japan_level"`
	BritainLevel int `json:"britain_level"`

	UsaLevelClaimed     int `json:"usa_level_claimed"`
	GermanyLevelClaimed int `json:"germany_level_claimed"`
	SovietLevelClaimed  int `json:"soviet_level_claimed"`
	JapanLevelClaimed   int `json:"japan_level_claimed"`
	BritainLevelClaimed int `json:"britain_level_claimed"`

	LastLogonDate       string `json:"last_logon_date"`
	DoubleXpEndDate     string `json:"double_xp_end_date"`
	LastCrateClaimedAt  string `json:"last_crate_claimed_date"`
	ClaimableCrateLevel int    `json:"claimable_crate_level"`

	Misc SDKPlayersInfoMisc `json:"misc"`
}

type SDKDeckHeader struct {
	ID                    int    `json:"ID"`
	Name                  string `json:"Name"`
	MainFaction           string `json:"main_faction"`
	AllyFaction           string `json:"ally_faction"`
	CardBack              string `json:"card_back"`
	DeckInvalid           bool   `json:"deck_invalid"`
	SkirmishInvalid       bool   `json:"skirmish_invalid"`
	KnockoutInvalid       bool   `json:"knockout_invalid"`
	DeckInvalidReason     string `json:"deck_invalid_reason"`
	SkirmishInvalidReason string `json:"skirmish_invalid_reason"`
	KnockoutInvalidReason string `json:"knockout_invalid_reason"`
	LastPlayed            string `json:"last_played"`
	Favorite              bool   `json:"favorite"`
	ContainsReserved      bool   `json:"contains_reserved"`
	DeckCode              string `json:"deck_code"`
}

type SDKMatch struct {
	CurrentActionID   int    `json:"current_action_id"`
	ActionPlayerID    int    `json:"action_player_id"`
	ActionSide        string `json:"action_side"`
	MatchURL          string `json:"match_url"`
	MatchType         string `json:"match_type"`
	CurrentTurn       int    `json:"current_turn"`
	MatchID           int    `json:"match_id"`
	Status            string `json:"Status"`
	StartSide         string `json:"start_side"`
	ActionsURL        string `json:"actions_url"`
	PlayerIDLeft      int    `json:"player_id_left"`
	PlayerIDRight     int    `json:"player_id_right"`
	DeckIDLeft        int    `json:"deck_id_left"`
	DeckIDRight       int    `json:"deck_id_right"`
	LeftIsOnline      bool   `json:"left_is_online"`
	RightIsOnline     bool   `json:"right_is_online"`
	PlayerStatusLeft  string `json:"player_status_left"`
	PlayerStatusRight string `json:"player_status_right"`
	WinnerID          int    `json:"winner_id"`
	WinnerSide        string `json:"winner_side"`
	ModifyDate        string `json:"modify_date"`
}

type SDKMatchPollResponse struct {
	Match           SDKMatch `json:"match"`
	OpponentPolling bool     `json:"opponent_polling"`
	SendMatchState  bool     `json:"send_match_state"`
}

type SDKPlayerResponse struct {
	PlayerID         int    `json:"player_id"`
	PlayerName       string `json:"player_name"`
	CounterURL       string `json:"counter_url"`
	CounterTotalsURL string `json:"countertotals_url"`
}

type SDKPlayerFriend struct {
	IsOnline   bool   `json:"is_online"`
	PlayerID   int    `json:"player_id"`
	PlayerName string `json:"player_name"`
	PlayerTag  int    `json:"player_tag"`
	Status     string `json:"Status"`
	BusyStatus string `json:"busy_status"`
}

type SDKPlayerItem struct {
	Details string `json:"details"`
	ItemID  string `json:"item_id"`
	Cnt     int    `json:"cnt"`
}

type SDKPlayerEquippedItem struct {
	ItemID  string `json:"item_id"`
	Slot    string `json:"Slot"`
	Faction string `json:"faction"`
}

type SDKPlayerStat struct {
	StatName string `json:"stat_name"`
	StatInt  int    `json:"stat_int"`
	StatStr  string `json:"stat_str"`
}

type SDKPlayerNotification struct {
	ID           int    `json:"ID"`
	Notification string `json:"notification"`
	Int1         int    `json:"int_1"`
	Int2         int    `json:"int_2"`
	Str1         string `json:"str_1"`
	Str2         string `json:"str_2"`
	FromPlayerID int    `json:"from_player_id"`
}

type SDKPlayerPack struct {
	ID         int    `json:"ID"`
	PlayerID   int    `json:"player_id"`
	CardSet    string `json:"card_set"`
	DateOpened string `json:"date_opened"`
	Details    string `json:"details"`
	UnlockDate string `json:"unlock_date"`
}
