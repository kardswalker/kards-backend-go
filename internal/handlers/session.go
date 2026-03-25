package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kards-backend-go/internal/config"
	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"
	"kards-backend-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ServerOptions 定义结构体以保证 JSON 键值的物理顺序
// websocketurl 被严格放置在 winter_war_date 和 homefront_date 之间
type ServerOptions struct {
	ChristmasMusic         int                      `json:"christmas_music"`
	NuiMobile              int                      `json:"nui_mobile"`
	BetaExpiryDate         string                   `json:"beta_expiry_date"`
	ScalabilityOverride    map[string]interface{}   `json:"scalability_override"`
	AppscaleDesktopDefault float64                  `json:"appscale_desktop_default"`
	AppscaleDesktopMax     float64                  `json:"appscale_desktop_max"`
	AppscaleMobileDefault  float64                  `json:"appscale_mobile_default"`
	AppscaleMobileMax      float64                  `json:"appscale_mobile_max"`
	AppscaleMobileMin      float64                  `json:"appscale_mobile_min"`
	AppscaleTabletMin      float64                  `json:"appscale_tablet_min"`
	BattleWaitTime         int                      `json:"battle_wait_time"`
	BrothersInArmsDate     string                   `json:"brothers_in_arms_date"`
	CovertOpsDate          string                   `json:"covert_ops_date"`
	NavalWarfareDate       string                   `json:"naval_warfare_date"`
	FirstPurchaseBonusDate string                   `json:"first_purchase_bonus_date"`
	Reconnect              int                      `json:"reconnect"`
	LoggerDisabled         int                      `json:"logger_disabled"`
	NewRewards             int                      `json:"new_rewards"`
	MostPopularProducts    string                   `json:"most_popular_products"`
	WinterWarDate          string                   `json:"winter_war_date"`
	WebsocketURL           string                   `json:"websocketurl"` // 位置关键
	HomefrontDate          string                   `json:"homefront_date"`
	ShowFullImage          bool                     `json:"show_full_image"`
	NewEffectBar           int                      `json:"new_effect_bar"`
	NewEffectBarPc         int                      `json:"new_effect_bar_pc"`
	NewEffectIcons         int                      `json:"new_effect_icons"`
	Versions               []string                 `json:"versions"`
	LockedCards            []map[string]interface{} `json:"locked_cards"`
	ReserveChanges         []map[string]interface{} `json:"reserve_changes"`
	DraftCardLimits        map[string]interface{}   `json:"draft_card_limits"`
	GiveGuestName          int                      `json:"give_guest_name"`
}

// DeckHeader 卡组头信息（与 decks.headers 内对象顺序一致）
type DeckHeader struct {
	Name        string `json:"name"`
	MainFaction string `json:"main_faction"`
	AllyFaction string `json:"ally_faction"`
	CardBack    string `json:"card_back"`
	DeckCode    string `json:"deck_code"`
	Favorite    bool   `json:"favorite"`
	ID          uint   `json:"id"`
	PlayerID    uint   `json:"player_id"`
	LastPlayed  string `json:"last_played"`
	CreateDate  string `json:"create_date"`
	ModifyDate  string `json:"modify_date"`
}

// Decks 卡组容器
type Decks struct {
	Headers []DeckHeader `json:"headers"`
}

// Misc 杂项信息
type Misc struct {
	CreateDate           string        `json:"createDate"`
	FeaturedAchievements []interface{} `json:"featuredAchievements"`
}

// Rewards 奖励信息
type Rewards struct {
	Packs   int `json:"packs"`
	GoldMax int `json:"gold_max"`
	GoldMin int `json:"gold_min"`
}

// NewPlayerLoginReward 新手登录奖励
type NewPlayerLoginReward struct {
	Day     int    `json:"day"`
	Reset   string `json:"reset"`
	Seconds int    `json:"seconds"`
}

// CardsBlacklistItem 黑名单卡牌项
type CardsBlacklistItem struct {
	CardType string `json:"card_type"`
	EndDate  string `json:"end_date"`
}

// SessionResponse 会话响应结构体（字段顺序与示例完全一致）
type SessionResponse struct {
	AchievementsURL         string                 `json:"achievements_url"`
	AllKnockoutTourneys     []interface{}          `json:"all_knockout_tourneys"`
	BritainLevel            int                    `json:"britain_level"`
	BritainLevelClaimed     int                    `json:"britain_level_claimed"`
	BritainXP               int                    `json:"britain_xp"`
	CardsBlacklist          []CardsBlacklistItem   `json:"cards_blacklist"`
	ClaimableCrateLevel     int                    `json:"claimable_crate_level"`
	ClientID                uint                   `json:"client_id"`
	Currency                string                 `json:"currency"`
	CurrentKnockoutTourney  map[string]interface{} `json:"current_knockout_tourney"`
	DailymissionsURL        string                 `json:"dailymissions_url"`
	Decks                   Decks                  `json:"decks"`
	DecksURL                string                 `json:"decks_url"`
	Diamonds                int                    `json:"diamonds"`
	DoubleXPEndDate         string                 `json:"double_xp_end_date"`
	DraftAdmissions         int                    `json:"draft_admissions"`
	Dust                    int                    `json:"dust"`
	Email                   interface{}            `json:"email"`
	EmailRewardReceived     bool                   `json:"email_reward_received"`
	EmailVerified           bool                   `json:"email_verified"`
	ExtendedRewards         bool                   `json:"extended_rewards"`
	GermanyLevel            int                    `json:"germany_level"`
	GermanyLevelClaimed     int                    `json:"germany_level_claimed"`
	GermanyXP               int                    `json:"germany_xp"`
	Gold                    int                    `json:"gold"`
	HasBeenOfficer          bool                   `json:"has_been_officer"`
	HeartbeatURL            string                 `json:"heartbeat_url"`
	IsOfficer               bool                   `json:"is_officer"`
	IsOnline                bool                   `json:"is_online"`
	JapanLevel              int                    `json:"japan_level"`
	JapanLevelClaimed       int                    `json:"japan_level_claimed"`
	JapanXP                 int                    `json:"japan_xp"`
	Jti                     string                 `json:"jti"`
	Jwt                     string                 `json:"jwt"`
	LastCrateClaimedDate    string                 `json:"last_crate_claimed_date"`
	LastDailyMissionCancel  interface{}            `json:"last_daily_mission_cancel"`
	LastDailyMissionRenewal string                 `json:"last_daily_mission_renewal"`
	LastLogonDate           string                 `json:"last_logon_date"`
	LaunchMessages          []interface{}          `json:"launch_messages"`
	LibraryURL              string                 `json:"library_url"`
	LinkerAccount           string                 `json:"linker_account"`
	Locale                  string                 `json:"locale"`
	Misc                    Misc                   `json:"misc"`
	NewCards                []string               `json:"new_cards"`
	NewPlayerLoginReward    NewPlayerLoginReward   `json:"new_player_login_reward"`
	Npc                     bool                   `json:"npc"`
	OnlineFlag              bool                   `json:"online_flag"`
	PacksURL                string                 `json:"packs_url"`
	PlayerID                uint                   `json:"player_id"`
	PlayerName              string                 `json:"player_name"`
	PlayerTag               int                    `json:"player_tag"`
	Rewards                 Rewards                `json:"rewards"`
	SeasonEnd               string                 `json:"season_end"`
	SeasonID                int                    `json:"season_id"`
	SeasonWins              int                    `json:"season_wins"`
	ServerOptions           string                 `json:"server_options"`
	ServerTime              string                 `json:"server_time"`
	SovietLevel             int                    `json:"soviet_level"`
	SovietLevelClaimed      int                    `json:"soviet_level_claimed"`
	SovietXP                int                    `json:"soviet_xp"`
	Stars                   int                    `json:"stars"`
	TutorialsDone           int                    `json:"tutorials_done"`
	TutorialsFinished       []string               `json:"tutorials_finished"`
	UsaLevel                int                    `json:"usa_level"`
	UsaLevelClaimed         int                    `json:"usa_level_claimed"`
	UsaXP                   int                    `json:"usa_xp"`
	UserID                  uint                   `json:"user_id"`
}

func HandleSession(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// 1. 查找或创建用户
	var user models.User
	result := database.DB.Preload("Decks").Where("username = ?", credentials.Username).First(&user)
	if result.Error != nil {
		user = models.User{
			Username:   credentials.Username,
			Password:   credentials.Password,
			PlayerName: "<anon>",
			PlayerTag:  0,
		}
		database.DB.Create(&user)
	}

	// 2. 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, _ := token.SignedString(config.JWTKey)
	database.DB.Model(&user).Update("PlayerJWT", tokenString)

	// 3. 查询卡组（关键修复）
	var decks []models.Deck
	if err := database.DB.Where("user_id = ?", user.ID).Find(&decks).Error; err != nil {
		decks = []models.Deck{}
	}

	// 构建 deck headers
	deckHeaders := make([]DeckHeader, 0, len(decks))
	for _, deck := range decks {
		deckHeaders = append(deckHeaders, DeckHeader{
			Name:        deck.Name,
			MainFaction: deck.MainFaction,
			AllyFaction: deck.AllyFaction,
			CardBack:    deck.CardBack,
			DeckCode:    deck.DeckCode,
			Favorite:    deck.Favorite,
			ID:          deck.ID,
			PlayerID:    user.ID, // ⚠️ 用 user.ID 更稳
			LastPlayed:  deck.LastPlayed,
			CreateDate:  deck.CreateDate,
			ModifyDate:  deck.ModifyDate,
		})
	}

	// 4. 精确复刻庞大的 ServerOptions 结构
	opts := ServerOptions{
		ChristmasMusic: 0,
		NuiMobile:      1,
		BetaExpiryDate: "2023-10-01 00:00:00",
		ScalabilityOverride: map[string]interface{}{
			"Android_Low":  map[string]interface{}{"console_commands": []string{"r.Screenpercentage 100"}},
			"Android_Mid":  map[string]interface{}{"console_commands": []string{"r.Screenpercentage 100"}},
			"Android_High": map[string]interface{}{"console_commands": []string{"r.Screenpercentage 100"}},
		},
		AppscaleDesktopDefault: 1.0,
		AppscaleDesktopMax:     1.4,
		AppscaleMobileDefault:  1.4,
		AppscaleMobileMax:      1.4,
		AppscaleMobileMin:      1.0,
		AppscaleTabletMin:      1.0,
		BattleWaitTime:         600,
		BrothersInArmsDate:     "2023.06.18-09.30.00",
		CovertOpsDate:          "2024.06.11-11.00.00",
		NavalWarfareDate:       "2025.05.22-12.00.00",
		FirstPurchaseBonusDate: "2025.07.24-11.10.00",
		Reconnect:              1,
		LoggerDisabled:         0,
		NewRewards:             1,
		MostPopularProducts:    "304;238;318;319;320;321;322;324;11;270;1;52;45;18;56;72;7;149;70;143;9;53;57;75;151;228;153;79",
		WinterWarDate:          "2023.11.29-09.00.00",
		WebsocketURL:           fmt.Sprintf("ws://%s:%s/ws", config.Host, config.WSPort),
		HomefrontDate:          "2025.11.27-09.00.00",
		ShowFullImage:          true,
		NewEffectBar:           1,
		NewEffectBarPc:         1,
		NewEffectIcons:         1,
		Versions:               []string{"Kards 1.47", "Kards 1.49", "Kards 1.50"},
		LockedCards: []map[string]interface{}{
			{
				"cards":       []string{"card_unit_whirlwind", "card_event_pound", "card_unit_tiger_moth", "card_event_harass", "card_unit_salamander", "card_unit_henschel_he_129", "card_unit_ilyushin_10", "card_unit_p_39_airacobra", "card_event_out_with_the_old", "card_unit_tigercat", "card_unit_seahawk", "card_event_screening_force", "card_unit_n1k1_kyofu", "card_event_flight_to_oblivion", "card_unit_kyushu_j7w3", "card_event_sally", "card_event_air_strips", "card_event_pilot_escape", "card_unit_fokker_finland"},
				"unlock_date": "2025-09-23T12:10:00",
			},
		},
		ReserveChanges: []map[string]interface{}{
			{
				"due_date":       "2025-11-27T10:00:00",
				"coming_back":    []string{"card_unit_fw_190_f", "card_unit_royal_ulster_rifles", "card_unit_c47_skytrain", "card_event_monastyr", "card_event_last_rites", "card_event_u_99", "card_unit_seaforth_highlanders", "card_event_five_year_plan", "card_unit_red_devils", "card_event_for_the_emperor", "card_event_joint_operation", "card_unit_m3a3_honey_desert", "card_unit_bryansk_irregulars", "card_unit_m16_halftrack", "card_unit_ki_61_tony", "card_unit_me_bf_109_v2", "card_unit_spitfire_v", "card_unit_25th_infantry_regiment", "card_unit_17th_infantry_regiment", "card_event_jungle_warfare", "card_unit_french_expedition", "card_event_test_resistance2", "card_event_arming_resistance", "card_event_liberation"},
				"to_be_reserved": []string{"card_unit_french_p_36", "card_event_compromise", "card_unit_1er_regiment_etranger", "card_unit_2e_brigade", "card_unit_obice_da_75_13", "card_event_eastern_front", "card_event_frozen_assets", "card_event_agency_africa", "card_event_envelop", "card_event_alpenfestung", "card_unit_comet", "card_unit_king_tiger_ii", "card_unit_short_wellington", "card_event_hms_talbot", "card_unit_sexton", "card_unit_comet1", "card_event_yank_the_army_weekly", "card_event_victory_march", "card_unit_13th_engineers_battalion", "card_unit_m26_pershing", "card_event_winter_offensive", "card_unit_kv_1_1941", "card_event_gambit", "card_event_tractor_factories", "card_unit_1st_yokosuka", "card_unit_kawanishi_h6k", "card_unit_pete", "card_unit_kokuras_sword", "card_unit_kurmark_aufklarungs", "card_unit_139_gebirgsjager", "card_event_mountain_offense", "card_unit_141_gebirgsjager", "card_event_blitz_doctrine", "card_event_baker_street_irregulars", "card_unit_vildebeest", "card_unit_green_howards", "card_event_enemy_contact", "card_unit_sdf", "card_event_bomber_mafia", "card_unit_blue_and_gray", "card_event_embargo", "card_event_mobilization", "card_unit_p_40_k", "card_unit_125_rifle_regiment", "card_event_by_the_sword", "card_unit_60th_cavalry_regiment", "card_unit_85mm_d44", "card_event_mass_attack", "card_unit_mikawa_regiment", "card_unit_winter_regiment", "card_unit_dinah", "card_event_code_of_bushido", "card_event_divine_wind", "card_event_tip_of_the_spear", "card_unit_stug_iv", "card_unit_171_infantry_regiment", "card_unit_panzer_iv_f2", "card_event_hit_the_drop_point", "card_event_commando_raid", "card_unit_15th_recon_regiment", "card_unit_blenheim_mk_i", "card_unit_skua_mk_ii", "card_event_infiltrate", "card_event_firebomb", "card_unit_the_deuce", "card_unit_37mm_m1_aa_gun", "card_event_supply_priorities", "card_event_furor", "card_unit_35th_mountain_rifles", "card_unit_13th_rifle_regiment", "card_event_turmoil", "card_unit_392nd_rifles", "card_unit_266th_guards_rifles", "card_unit_nakajima_b5n", "card_unit_tsu_regiment", "card_unit_type_89", "card_unit_sally", "card_event_take_liberty", "card_unit_sturm_brigade_rhodos", "card_event_recuperation", "card_unit_13_gebirgsjager", "card_unit_136_gebirgsjager", "card_unit_50_infantry_regiment", "card_unit_15_aufklarungs", "card_unit_panzer_iv_g", "card_event_lure", "card_unit_5th_parachute_brigade", "card_event_sea_embargo", "card_event_delaying_tactics", "card_unit_qf_40mm_mk_iii", "card_event_stiff_upper_lip", "card_unit_no_9_commando", "card_event_america_goes_to_war", "card_event_hidden_plans", "card_unit_164th_infantry_regiment", "card_event_uncle_sam", "card_unit_4th_marines", "card_unit_103rd_cavalry_recon", "card_unit_142nd_infantry_regiment", "card_event_field_recon", "card_unit_95th_rifles", "card_unit_1357th_rifles", "card_unit_t_34_85", "card_event_covert_operation", "card_unit_bt_7", "card_unit_34th_guards", "card_unit_e7k", "card_event_shock_attack", "card_event_diplomatic_attache", "card_event_grand_plans", "card_unit_g4m1_betty", "card_event_imperial_strength", "card_unit_51st_recon", "card_event_contest_doctrine"},
			},
		},
		DraftCardLimits: map[string]interface{}{
			"blacklist": []string{"card_event_tactical_withdrawal", "card_event_hms_spectre", "card_event_lure", "card_event_naval_power", "card_event_fortification", "card_event_for_the_king", "card_event_overrun", "card_event_creeping_barrage", "card_unit_qf_40mm_mk_iii", "card_unit_no43_commando", "card_event_bpf", "card_event_hms_formidable", "card_event_national_fire_service", "card_event_grounded", "card_unit_the_rangers", "card_event_breakout", "card_unit_7th_scottish_borderers", "card_event_royal_research", "card_event_eagle_day", "card_event_reich_research", "card_event_fliegerfuhrer_atlantik", "card_event_fast_heinz", "card_event_foiled_plans", "card_event_order_no_227", "card_event_no_retreat", "card_event_turning_point", "card_event_workers_unite", "card_event_out_with_the_old", "card_event_blood_red_sky", "card_unit_175th_infantry_regiment", "card_unit_104th_infantry_regiment", "card_event_yank_the_army_weekly", "card_unit_cyclone_division", "card_unit_blue_and_gray", "card_unit_2nd_michigan", "card_event_fortunes_of_war", "card_unit_121st_infantry_regiment", "card_event_us_military_research", "card_unit_cedar_division", "card_event_the_tide_turns", "card_event_surprise_attack", "card_event_isolation", "card_event_imperial_decree", "card_event_honor", "card_event_burning_sun", "card_event_distant_front", "card_unit_nakajima_b5n", "card_event_outmaneuver", "card_unit_dinah", "card_event_for_prosperety", "card_unit_fukuoka_regiment", "card_unit_388th_independent", "card_unit_40th_cavalry_regiment", "card_event_yamamoto", "card_event_redeployment", "card_event_shock_attack", "card_event_great_expanse", "card_event_reconnaissance", "card_event_yamato", "card_event_flight_to_oblivion", "card_unit_infantry_regiment_56", "card_unit_infantry_regiment_36", "card_event_friendly_fire_fin", "card_unit_sissiosasto_5d", "card_unit_stug_iii_g_fin", "card_event_free_french_navy", "card_event_honor_and_loyalty", "card_event_vanguard_ita", "card_event_lysander_pl", "card_event_act_on_the_intel", "card_event_orp_blyskawica", "card_unit_506th_airborne", "card_event_overwhelming_force", "card_event_rationing", "card_unit_8th_marines", "card_unit_f6f_hellcat_home", "card_unit_layforce", "card_unit_royal_ulster_rifles", "card_unit_27_fusiliers", "card_unit_ace_of_spades", "card_event_u_99", "card_unit_106th_infantry_regiment", "card_event_last_rites", "card_event_way_of_subjects", "card_event_persian_corridor", "card_event_rm_bersaglieri", "card_unit_iron_division", "card_unit_7tp", "card_event_decoy_tactics", "card_event_orp_orzel"},
			"whitelist": []string{"card_unit_baluch_regiment", "card_unit_the_glamour_boys", "card_unit_east_surray_regiment", "card_unit_3rd_canadian_division", "card_unit_spitfire_v", "card_unit_rnzaf_kittyhawk", "card_unit_british_buffalo", "card_unit_typhoon_mk_ib", "card_unit_stirling", "card_event_bomber_run_brit", "card_event_shelling", "card_event_strong_bond", "card_event_supply_shortage", "card_event_sea_patrol", "card_event_lend_lease", "card_event_sincerely_yours", "card_unit_no1_commando", "card_unit_royal_engineers", "card_unit_mosquito", "card_unit_mathilda_mk_ii", "card_unit_p38_lightning_raf", "card_unit_coastal_gun", "card_unit_2_pounder", "card_unit_1st_airlanding_brigade", "card_unit_spitfire_mk_ii", "card_unit_me_410_hornisse", "card_unit_4_pioneer", "card_unit_stug_iii_g", "card_unit_panther_d_winter", "card_unit_78th_steel_regiment", "card_event_hold_the_line", "card_unit_spitfire_mk_v_pol", "card_unit_pzl_23_karas", "card_event_first_to_fight", "card_unit_6th_alpini_regiment", "card_event_med_raid", "card_event_out_of_the_mist", "card_event_under_fire_japan", "card_event_audacity", "card_unit_24th_cavalry_regiment", "card_unit_utsunomiya_regiment", "card_unit_japan_ki_100_goshiken", "card_unit_shiden", "card_unit_toyama_regiment", "card_unit_1st_taipei_regiment", "card_unit_28cm_coasttal_howitzer", "card_unit_ki_44_tojo", "card_event_deadly_duty", "card_unit_d4y_suisei", "card_event_naval_supply_run", "card_event_bombing_raid", "card_unit_34th_infantry_regiment", "card_unit_3rd_yokosuka", "card_unit_51st_recon", "card_unit_aichi_b7a2", "card_unit_b6n_tenzan", "card_unit_a5m4_claude", "card_unit_a6m3_zeke", "card_event_home_defense", "card_event_honorable_death", "card_event_on_the_ascend", "card_unit_type_3_chi_nu", "card_unit_okayama_regiment", "card_unit_petlyakov_pe_2", "card_unit_321st_rifles", "card_unit_2nd_motor", "card_unit_t_34_flame", "card_event_urban_fighting", "card_unit_35th_rifles", "card_event_mass_deployment", "card_unit_devils_brigade", "card_event_shore_bombardment", "card_unit_75th_rangers", "card_unit_p38_lightning", "card_event_torpedo_attack", "card_unit_11th_infantry_regiment", "card_unit_2_5_marines", "card_unit_m24_chaffee", "card_event_cadet_nurse_corps", "card_unit_4th_marines", "card_event_america_goes_to_war", "card_unit_164th_infantry_regiment", "card_event_firebomb", "card_unit_5th_parachute_brigade", "card_unit_comet1", "card_unit_stug_iv", "card_unit_king_tiger_ii", "card_unit_51st_recon", "card_unit_pete", "card_unit_me_bf_110_waw", "card_event_home_defense", "card_unit_60th_cavalry_regiment"},
		},
		GiveGuestName: 0,
	}

	// 序列化 ServerOptions 结构体
	serverOptionsBytes, _ := json.Marshal(opts)

	// 5. 组装最终响应
	ip := config.Host + ":" + config.Port
	userID := user.ID

	// 构建 decks 数据
	decksData := Decks{
		Headers: deckHeaders,
	}

	// 构建 misc 数据
	miscData := Misc{
		CreateDate:           time.Now().Format(time.RFC3339Nano),
		FeaturedAchievements: []interface{}{},
	}

	// 构建 rewards 数据
	rewardsData := Rewards{
		Packs:   0,
		GoldMax: 114,
		GoldMin: 514,
	}

	// 构建 new_player_login_reward 数据
	newPlayerLoginReward := NewPlayerLoginReward{
		Day:     8,
		Reset:   "0001-01-01 00:00:00",
		Seconds: 0,
	}

	// 构建 cards_blacklist 数据
	cardsBlacklist := []CardsBlacklistItem{
		{
			CardType: "card_location_british_scen2",
			EndDate:  "2026-12-31T23:59:59.999Z",
		},
	}

	// 构建响应结构体
	resp := SessionResponse{
		AchievementsURL:         fmt.Sprintf("http://%s/players/%d/achievements", ip, userID),
		AllKnockoutTourneys:     []interface{}{},
		BritainLevel:            500,
		BritainLevelClaimed:     500,
		BritainXP:               0,
		CardsBlacklist:          cardsBlacklist,
		ClaimableCrateLevel:     1,
		ClientID:                userID,
		Currency:                "USD",
		CurrentKnockoutTourney:  map[string]interface{}{},
		DailymissionsURL:        fmt.Sprintf("http://%s/players/%d/dailymissions", ip, userID),
		Decks:                   decksData,
		DecksURL:                fmt.Sprintf("http://%s/players/%d/decks", ip, userID),
		Diamonds:                91,
		DoubleXPEndDate:         "2026-07-03T12:13:36.889692Z",
		DraftAdmissions:         1,
		Dust:                    0,
		Email:                   nil,
		EmailRewardReceived:     false,
		EmailVerified:           false,
		ExtendedRewards:         true,
		GermanyLevel:            500,
		GermanyLevelClaimed:     500,
		GermanyXP:               0,
		Gold:                    78,
		HasBeenOfficer:          true,
		HeartbeatURL:            fmt.Sprintf("http://%s/players/%d/heartbeat", ip, userID),
		IsOfficer:               true,
		IsOnline:                true,
		JapanLevel:              500,
		JapanLevelClaimed:       500,
		JapanXP:                 0,
		Jti:                     "1",
		Jwt:                     tokenString,
		LastCrateClaimedDate:    "2025-07-02T11:24:15.567042Z",
		LastDailyMissionCancel:  nil,
		LastDailyMissionRenewal: "2025-07-05T15:21:43.653915Z",
		LastLogonDate:           "2025-07-05T15:21:06.168847Z",
		LaunchMessages:          []interface{}{},
		LibraryURL:              fmt.Sprintf("http://%s/players/%d/library", ip, userID),
		LinkerAccount:           "",
		Locale:                  "zh-hans",
		Misc:                    miscData,
		NewCards:                []string{}, // 可动态填充
		NewPlayerLoginReward:    newPlayerLoginReward,
		Npc:                     false,
		OnlineFlag:              true,
		PacksURL:                fmt.Sprintf("http://%s/players/%d/packs", ip, userID),
		PlayerID:                userID,
		PlayerName:              user.PlayerName,
		PlayerTag:               user.PlayerTag,
		Rewards:                 rewardsData,
		SeasonEnd:               "2027-01-01T00:00:00Z",
		SeasonID:                85, // 示例中为 85，可配置
		SeasonWins:              91,
		ServerOptions:           string(serverOptionsBytes),
		ServerTime:              utils.GetKardsNow(),
		SovietLevel:             500,
		SovietLevelClaimed:      500,
		SovietXP:                0,
		Stars:                   0,
		TutorialsDone:           0,
		TutorialsFinished: []string{
			"unlocking_germany_1", "unlocking_germany_2", "unlocking_germany_0",
			"germany_cards_rewarded", "unlocking_usa_8", "recruit_missions_done",
			"draft_1", "draft_ally", "draft_kredits", "unlocking_japan_0",
			"japan_cards_rewarded", "unlocking_soviet_0", "soviet_cards_rewarded",
			"unlocking_usa_0", "usa_cards_rewarded", "unlocking_britain_0",
			"britain_cards_rewarded",
		},
		UsaLevel:        500,
		UsaLevelClaimed: 500,
		UsaXP:           0,
		UserID:          userID,
	}

	h := c.Writer.Header()
	h.Set("Connection", "keep-alive")
	h.Set("Keep-Alive", "timeout=5")
	c.JSON(http.StatusOK, resp)
}
