package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/config"
	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/handler"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/storage"
	"github.com/thawng/velox/internal/transcoder"
	"github.com/thawng/velox/internal/trickplay"
	"github.com/thawng/velox/internal/websocket"
	"github.com/thawng/velox/pkg/fanart"
	"github.com/thawng/velox/pkg/omdb"
	"github.com/thawng/velox/pkg/thetvdb"
	"github.com/thawng/velox/pkg/tmdb"
	"github.com/thawng/velox/pkg/tvmaze"
)

type serverApp struct {
	cfg          *config.Config
	startTime    time.Time
	db           *sql.DB
	backgroundDB *sql.DB
	hwAccel      string
	tmdbClient   *tmdb.Client
	trickplayGen *trickplay.Generator
	jwtManager   *auth.JWTManager
	apiKeyStore  *auth.APIKeyStore
	wsHub        *websocket.Hub
	repos        serverRepos
	services     serverServices
	handlers     serverHandlers
}

type serverRepos struct {
	library            *repository.LibraryRepo
	media              *repository.MediaRepo
	mediaFile          *repository.MediaFileRepo
	series             *repository.SeriesRepo
	season             *repository.SeasonRepo
	episode            *repository.EpisodeRepo
	scanJob            *repository.ScanJobRepo
	subtitle           *repository.SubtitleRepo
	audioTrack         *repository.AudioTrackRepo
	user               *repository.UserRepo
	refreshToken       *repository.RefreshTokenRepo
	session            *repository.SessionRepo
	prefs              *repository.UserPreferencesRepo
	userData           *repository.UserDataRepo
	genre              *repository.GenreRepo
	person             *repository.PersonRepo
	activity           *repository.ActivityRepo
	webhook            *repository.WebhookRepo
	marker             *repository.MediaMarkerRepo
	fingerprint        *repository.AudioFingerprintRepo
	notification       *repository.NotificationRepo
	appSettings        *repository.AppSettingsRepo
	pretranscodeStatus *repository.PretranscodeRepo
	appVersion         *repository.AppVersionRepo
}

type serverServices struct {
	pipeline       *scanner.Pipeline
	metadata       *service.MetadataService
	transcoder     *transcoder.Transcoder
	library        *service.LibraryService
	media          *service.MediaService
	setup          *service.SetupService
	settings       *service.SettingsService
	profile        *service.ProfileService
	playbackConfig *service.PlaybackConfigService
	catalog        *service.CatalogService
	browse         *service.BrowseService
	cinema         *service.CinemaService
	series         *service.SeriesService
	stream         *service.StreamService
	streamManager  *transcoder.Manager
	auth           *service.AuthService
	userData       *service.UserDataService
	subtitle       *service.SubtitleService
	audioTrack     *service.AudioTrackService
	marker         *service.MarkerService
	activity       *service.ActivityService
	admin          *service.AdminService
	webhook        *service.WebhookService
	notification   *service.NotificationService
	pretranscode   *service.PretranscodeService
	subtitleSearch *service.SubtitleSearchService
	scheduler      *service.Scheduler
	verifier       *scanner.Verifier
}

type serverHandlers struct {
	library        *handler.LibraryHandler
	media          *handler.MediaHandler
	stream         *handler.StreamHandler
	streamURL      *handler.StreamURLHandler
	setup          *handler.SetupHandler
	auth           *handler.AuthHandler
	user           *handler.UserHandler
	profile        *handler.ProfileHandler
	playback       *handler.PlaybackHandler
	subtitle       *handler.SubtitleHandler
	audioTrack     *handler.AudioTrackHandler
	settings       *handler.SettingsHandler
	subtitleSearch *handler.SubtitleSearchHandler
	trickplay      *handler.TrickplayHandler
	image          *handler.ImageHandler
	browse         *handler.BrowseHandler
	catalog        *handler.CatalogHandler
	series         *handler.SeriesHandler
	metadata       *handler.MetadataHandler
	cinema         *handler.CinemaHandler
	activity       *handler.ActivityHandler
	admin          *handler.AdminHandler
	webhook        *handler.WebhookHandler
	markerAdmin    *handler.MarkerAdminHandler
	notification   *handler.NotificationHandler
	pretranscode   *handler.PretranscodeHandler
	ws             *handler.WebSocketHandler
	scheduler      *handler.SchedulerHandler
	appVersion     *handler.AppVersionHandler
}

func newServerApp(cfg *config.Config) (*serverApp, error) {
	db, err := openMigratedDatabase(cfg.DatabasePath)
	if err != nil {
		return nil, err
	}

	jwtManager, apiKeyStore, err := newAuthDependencies(cfg.DataDir)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	app := &serverApp{
		cfg:         cfg,
		startTime:   time.Now(),
		db:          db,
		hwAccel:     resolveHWAccel(cfg.HWAccel),
		jwtManager:  jwtManager,
		apiKeyStore: apiKeyStore,
		wsHub:       websocket.NewHub(slog.Default()),
	}

	app.repos = newServerRepos(db)
	go app.wsHub.Run()

	if err := app.initServices(); err != nil {
		app.Close()
		return nil, err
	}

	app.initHandlers()

	return app, nil
}

func (app *serverApp) Close() {
	if app.services.activity != nil {
		app.services.activity.Close()
	}
	if app.backgroundDB != nil {
		_ = app.backgroundDB.Close()
	}
	if app.db != nil {
		_ = app.db.Close()
	}
}

func openMigratedDatabase(path string) (*sql.DB, error) {
	db, err := database.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.Migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return db, nil
}

func newAuthDependencies(dataDir string) (*auth.JWTManager, *auth.APIKeyStore, error) {
	jwtSecret, err := auth.LoadOrCreateSecret(dataDir)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize JWT: %w", err)
	}

	return auth.NewJWTManager(jwtSecret), auth.NewAPIKeyStore(), nil
}

func newServerRepos(db *sql.DB) serverRepos {
	return serverRepos{
		library:            repository.NewLibraryRepo(db),
		media:              repository.NewMediaRepo(db),
		mediaFile:          repository.NewMediaFileRepo(db),
		series:             repository.NewSeriesRepo(db),
		season:             repository.NewSeasonRepo(db),
		episode:            repository.NewEpisodeRepo(db),
		scanJob:            repository.NewScanJobRepo(db),
		subtitle:           repository.NewSubtitleRepo(db),
		audioTrack:         repository.NewAudioTrackRepo(db),
		user:               repository.NewUserRepo(db),
		refreshToken:       repository.NewRefreshTokenRepo(db),
		session:            repository.NewSessionRepo(db),
		prefs:              repository.NewUserPreferencesRepo(db),
		userData:           repository.NewUserDataRepo(db),
		genre:              repository.NewGenreRepo(db),
		person:             repository.NewPersonRepo(db),
		activity:           repository.NewActivityRepo(db),
		webhook:            repository.NewWebhookRepo(db),
		marker:             repository.NewMediaMarkerRepo(db),
		fingerprint:        repository.NewAudioFingerprintRepo(db),
		notification:       repository.NewNotificationRepo(db),
		appSettings:        repository.NewAppSettingsRepo(db),
		pretranscodeStatus: repository.NewPretranscodeRepo(db),
		appVersion:         repository.NewAppVersionRepo(db),
	}
}

func (app *serverApp) initServices() error {
	repos := app.repos

	app.services.pipeline = scanner.NewPipeline(
		app.db,
		repos.library,
		repos.media,
		repos.mediaFile,
		repos.series,
		repos.season,
		repos.episode,
		repos.scanJob,
		repos.subtitle,
		repos.audioTrack,
		repos.marker,
		app.trickplayGen,
	)

	app.services.metadata, app.tmdbClient = app.initMetadataService(app.services.pipeline)
	app.services.transcoder = transcoder.New(app.cfg.TranscodePath, app.hwAccel, app.cfg.MaxTranscodes)
	app.services.library = service.NewLibraryService(repos.library, repos.scanJob, app.services.pipeline)
	app.services.media = service.NewMediaService(repos.media, repos.mediaFile)
	app.services.media.SetEpisodeRepo(repos.episode)
	app.services.media.SetSeasonRepo(repos.season)
	app.services.catalog = service.NewCatalogService(repos.media, repos.series, repos.genre, repos.person)
	app.services.browse = service.NewBrowseService(repos.library, repos.media, repos.mediaFile, repos.user)
	app.services.cinema = service.NewCinemaService(
		repos.appSettings,
		repos.media,
		repos.episode,
		repos.series,
		app.tmdbClient,
		app.cfg.DataDir,
	)
	app.services.series = service.NewSeriesService(repos.series, repos.season, repos.episode)
	app.services.stream = service.NewStreamService(repos.mediaFile, repos.audioTrack, app.services.transcoder)

	outDir := filepath.Join(app.cfg.DataDir, "hls")
	app.services.streamManager = transcoder.NewManager(outDir, app.hwAccel)

	app.services.auth = service.NewAuthService(repos.user, repos.refreshToken, repos.session, app.jwtManager, app.db)
	app.services.userData = service.NewUserDataService(repos.userData)
	app.services.subtitle = service.NewSubtitleService(app.db, repos.subtitle, repos.mediaFile, repos.media, repos.genre)
	app.services.subtitle.SetSettingsRepo(repos.appSettings)
	app.services.subtitle.SetCacheDir(app.cfg.SubtitleCachePath)
	app.services.audioTrack = service.NewAudioTrackService(app.db, repos.audioTrack)
	app.services.setup = service.NewSetupService(app.services.auth, repos.appSettings)
	app.services.settings = service.NewSettingsService(repos.appSettings, builtinSettingsPresence(app.cfg))
	app.services.profile = service.NewProfileService(app.services.auth, repos.prefs, app.services.userData)
	app.services.playbackConfig = service.NewPlaybackConfigService(repos.prefs, repos.appSettings)
	app.services.marker = service.NewMarkerService(repos.marker, repos.mediaFile, repos.fingerprint, repos.episode, repos.season, app.wsHub)
	app.services.activity = service.NewActivityService(repos.activity)
	app.services.admin = service.NewAdminService(app.db, repos.library, repos.mediaFile, repos.user, app.startTime, app.hwAccel, app.cfg.DatabasePath)
	app.services.webhook = service.NewWebhookService(repos.webhook)
	app.services.notification = service.NewNotificationService(repos.notification, repos.user, app.wsHub, slog.Default())
	app.services.notification.SetWebhookService(app.services.webhook)
	app.services.library.SetNotificationService(app.services.notification)
	app.services.stream.SetNotificationService(app.services.notification)

	if app.services.metadata != nil {
		app.services.metadata.SetNotificationService(app.services.notification)
	}

	if err := app.initPretranscodeService(); err != nil {
		return err
	}

	app.services.subtitleSearch = service.NewSubtitleSearchService(
		repos.media,
		repos.mediaFile,
		repos.subtitle,
		repos.appSettings,
		repos.episode,
		repos.season,
		repos.series,
		app.cfg.SubtitleCachePath,
	)
	app.services.subtitleSearch.SetBuiltinSubdlKey(app.cfg.SubdlAPIKey)
	app.services.subtitleSearch.SetNotificationService(app.services.notification)
	app.services.pipeline.SetSubtitleAutoDownloader(app.services.subtitleSearch)
	app.services.scheduler = service.NewScheduler()
	app.services.verifier = scanner.NewVerifier(repos.mediaFile)

	return app.initTrickplayGenerator()
}

func (app *serverApp) initPretranscodeService() error {
	bgDB, err := database.Open(app.cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open background database: %w", err)
	}

	app.backgroundDB = bgDB
	app.services.pretranscode = service.NewPretranscodeService(
		repository.NewPretranscodeRepo(bgDB),
		repository.NewMediaFileRepo(bgDB),
		repository.NewAppSettingsRepo(bgDB),
		repository.NewLibraryRepo(bgDB),
		app.cfg.PretranscodePath,
		app.hwAccel,
	)
	app.services.pretranscode.SetNotificationService(app.services.notification)
	app.services.pretranscode.SetTranscoder(app.services.transcoder)
	app.services.pretranscode.SetStatusRepo(app.repos.pretranscodeStatus)
	app.services.stream.SetPretranscodeService(app.services.pretranscode)
	app.services.library.SetPretranscodeService(app.services.pretranscode)

	return nil
}

func (app *serverApp) initTrickplayGenerator() error {
	if !app.cfg.TrickplayEnabled {
		return nil
	}

	if err := os.MkdirAll(app.cfg.TrickplayPath, 0755); err != nil {
		return fmt.Errorf("create trickplay dir: %w", err)
	}

	app.trickplayGen = trickplay.New(app.cfg.TrickplayPath, app.cfg.TrickplayInterval, app.hwAccel)
	log.Printf("trickplay enabled (interval: %ds, hwaccel: %s)", app.cfg.TrickplayInterval, app.hwAccel)

	return nil
}

func (app *serverApp) initHandlers() {
	services := app.services

	app.handlers.library = handler.NewLibraryHandler(services.library)
	app.handlers.media = handler.NewMediaHandler(services.media)
	app.handlers.stream = handler.NewStreamHandler(services.stream, services.streamManager)
	app.handlers.streamURL = handler.NewStreamURLHandler(app.apiKeyStore, services.stream)
	app.handlers.setup = handler.NewSetupHandler(services.setup)
	app.handlers.auth = handler.NewAuthHandler(services.auth)
	app.handlers.user = handler.NewUserHandler(services.auth)
	app.handlers.profile = handler.NewProfileHandler(services.profile)
	app.handlers.playback = handler.NewPlaybackHandler(
		services.media,
		services.stream,
		services.userData,
		services.subtitle,
		services.audioTrack,
		services.marker,
		services.playbackConfig,
		app.apiKeyStore,
	)
	app.handlers.subtitle = handler.NewSubtitleHandler(services.subtitle)
	app.handlers.audioTrack = handler.NewAudioTrackHandler(services.audioTrack)
	app.handlers.settings = handler.NewSettingsHandler(services.settings)
	app.handlers.subtitleSearch = handler.NewSubtitleSearchHandler(services.subtitleSearch)
	app.handlers.trickplay = handler.NewTrickplayHandler(app.trickplayGen, services.stream)
	app.handlers.image = handler.NewImageHandler()
	app.handlers.browse = handler.NewBrowseHandler(services.browse)
	app.handlers.catalog = handler.NewCatalogHandler(services.catalog)
	app.handlers.series = handler.NewSeriesHandler(services.series)
	app.handlers.metadata = handler.NewMetadataHandler(services.media, services.metadata, storage.NewImageStorage(app.cfg.DataDir))
	app.handlers.cinema = handler.NewCinemaHandler(services.cinema)
	app.handlers.activity = handler.NewActivityHandler(services.activity)
	app.handlers.admin = handler.NewAdminHandler(services.admin)
	app.handlers.webhook = handler.NewWebhookHandler(services.webhook)
	app.handlers.markerAdmin = handler.NewMarkerAdminHandler(services.marker)
	app.handlers.notification = handler.NewNotificationHandler(services.notification)
	app.handlers.pretranscode = handler.NewPretranscodeHandler(services.pretranscode)
	app.handlers.ws = handler.NewWebSocketHandler(app.wsHub, app.jwtManager, slog.Default())
	app.handlers.scheduler = handler.NewSchedulerHandler(services.scheduler)
	app.handlers.appVersion = handler.NewAppVersionHandler(app.repos.appVersion)

	app.handlers.auth.SetActivityService(services.activity)
	app.handlers.library.SetActivityService(services.activity)
	app.handlers.playback.SetActivityService(services.activity)
}

func (app *serverApp) initMetadataService(pipeline *scanner.Pipeline) (*service.MetadataService, *tmdb.Client) {
	ctx := context.Background()

	tmdbAPIKey, tmdbBuiltin := app.configuredAPIKey(ctx, model.SettingTMDbAPIKey, app.cfg.TMDbAPIKey)
	if tmdbAPIKey == "" {
		log.Println("TMDb API key not configured. Metadata enrichment disabled.")
		log.Println("Set VELOX_TMDB_API_KEY env var or configure in Settings → Metadata")
		return nil, nil
	}

	logAPIKeySource("TMDb", tmdbBuiltin)

	tmdbClient := tmdb.New(tmdbAPIKey)
	metadataSvc := service.NewMetadataService(
		tmdbClient,
		app.repos.media,
		app.repos.mediaFile,
		app.repos.series,
		app.repos.season,
		app.repos.episode,
		app.repos.genre,
		app.repos.person,
	)
	if metadataSvc == nil {
		return nil, tmdbClient
	}

	pipeline.SetMetadataMatcher(metadataSvc)
	log.Println("TMDb metadata enrichment enabled")

	omdbAPIKey, omdbBuiltin := app.configuredAPIKey(ctx, model.SettingOMDbAPIKey, app.cfg.OMDbAPIKey)
	if omdbAPIKey != "" {
		logAPIKeySource("OMDb", omdbBuiltin)
		metadataSvc.SetOMDbClient(omdb.New(omdbAPIKey))
		log.Println("OMDb ratings enrichment enabled (IMDb, Rotten Tomatoes, Metacritic)")
	}

	tvdbAPIKey, tvdbBuiltin := app.configuredAPIKey(ctx, model.SettingTVDBAPIKey, app.cfg.TVDBAPIKey)
	if tvdbAPIKey != "" {
		logAPIKeySource("TheTVDB", tvdbBuiltin)
		metadataSvc.SetTVDBClient(thetvdb.New(tvdbAPIKey))
		log.Println("TheTVDB metadata enrichment enabled")
	}

	fanartAPIKey, fanartBuiltin := app.configuredAPIKey(ctx, model.SettingFanartAPIKey, app.cfg.FanartAPIKey)
	if fanartAPIKey != "" {
		logAPIKeySource("fanart.tv", fanartBuiltin)
		metadataSvc.SetFanartClient(fanart.New(fanartAPIKey))
		log.Println("fanart.tv artwork enrichment enabled (logos, thumbs)")
	}

	metadataSvc.SetTVmazeClient(tvmaze.New())
	log.Println("TVmaze enrichment enabled (network, schedule, ID cross-reference)")

	return metadataSvc, tmdbClient
}

func (app *serverApp) configuredAPIKey(ctx context.Context, settingKey, builtin string) (string, bool) {
	value, _ := app.repos.appSettings.Get(ctx, settingKey)
	if value != "" {
		return value, false
	}
	if builtin != "" {
		return builtin, true
	}
	return "", false
}

func logAPIKeySource(provider string, usingBuiltin bool) {
	if usingBuiltin {
		log.Printf("%s using built-in API key from env (override in Settings → Metadata)", provider)
		return
	}

	log.Printf("%s using custom API key from settings", provider)
}

func resolveHWAccel(mode string) string {
	switch mode {
	case "auto":
		hwAccel := playback.DetectHWAccel()
		if hwAccel != "" {
			log.Printf("hardware acceleration: %s", hwAccel)
		} else {
			log.Printf("hardware acceleration: none detected, using software encoder")
		}
		return hwAccel
	case "none":
		return ""
	default:
		if !playback.IsHWAccelAvailable(mode) {
			log.Printf("hardware acceleration: %s requested but unavailable, using software encoder", mode)
			return ""
		}
		log.Printf("hardware acceleration: %s (configured)", mode)
		return mode
	}
}

func builtinSettingsPresence(cfg *config.Config) map[string]bool {
	return map[string]bool{
		"tmdb":   cfg.TMDbAPIKey != "",
		"omdb":   cfg.OMDbAPIKey != "",
		"tvdb":   cfg.TVDBAPIKey != "",
		"fanart": cfg.FanartAPIKey != "",
		"subdl":  cfg.SubdlAPIKey != "",
	}
}
