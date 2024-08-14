package config

import (
	"encoding/json"
	"flag"
	"os"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

const (
	TimeDateFormat = "2006-01-02 15:04:05"
	FiberLogFormat = "${ip} - [${time}] ${method} ${url} ${status} ${referer} ${ua}\n"
	WebpMax        = 16383
	AvifMax        = 65536
	HttpRegexp     = `^https?://`
	SampleConfig   = `
{
  "HOST": "127.0.0.1",
  "PORT": "3333",
  "QUALITY": "80",
  "IMG_PATH": "./pics",
  "EXHAUST_PATH": "./exhaust",
  "IMG_MAP": {},
  "ALLOWED_TYPES": ["jpg","png","jpeg","gif","bmp","svg","heic","nef"],
  "CONVERT_TYPES": ["webp"],
  "STRIP_METADATA": true,
  "ENABLE_EXTRA_PARAMS": false,
  "EXTRA_PARAMS_CROP_INTERESTING": "InterestingAttention",
  "READ_BUFFER_SIZE": 4096,
  "CONCURRENCY": 262144,
  "DISABLE_KEEPALIVE": false,
  "CACHE_TTL": 259200,
  "MAX_CACHE_SIZE": 0
}`
)

var (
	ConfigPath     string
	Jobs           int
	DumpSystemd    bool
	DumpConfig     bool
	ShowVersion    bool
	ProxyMode      bool
	Prefetch       bool
	Config         = NewWebPConfig()
	Version        = "0.1.0"
	WriteLock      = cache.New(5*time.Minute, 10*time.Minute)
	ConvertLock    = cache.New(5*time.Minute, 10*time.Minute)
	LocalHostAlias = "local"
	RemoteCache    *cache.Cache
)

// MetaFile struct
type MetaFile struct {
	Id       string `json:"id"`       // hash of below path️, also json file name id.webp
	Path     string `json:"path"`     // local: path with width and height, proxy: full url
	Checksum string `json:"checksum"` // hash of original file or hash(etag). Use this to identify changes
}

// WebpConfig struct
type WebpConfig struct {
	Host          string            `json:"HOST"`
	Port          string            `json:"PORT"`
	ImgPath       string            `json:"IMG_PATH"`
	Quality       int               `json:"QUALITY,string"`
	AllowedTypes  []string          `json:"ALLOWED_TYPES"`
	ConvertTypes  []string          `json:"CONVERT_TYPES"`
	ImageMap      map[string]string `json:"IMG_MAP"`
	ExhaustPath   string            `json:"EXHAUST_PATH"`
	MetadataPath  string            `json:"METADATA_PATH"`
	RemoteRawPath string            `json:"REMOTE_RAW_PATH"`

	EnableWebP bool `json:"ENABLE_WEBP"`
	EnableAVIF bool `json:"ENABLE_AVIF"`
	EnableJXL  bool `json:"ENABLE_JXL"`

	EnableExtraParams          bool   `json:"ENABLE_EXTRA_PARAMS"`
	ExtraParamsCropInteresting string `json:"EXTRA_PARAMS_CROP_INTERESTING"`

	StripMetadata    bool `json:"STRIP_METADATA"`
	ReadBufferSize   int  `json:"READ_BUFFER_SIZE"`
	Concurrency      int  `json:"CONCURRENCY"`
	DisableKeepalive bool `json:"DISABLE_KEEPALIVE"`
	CacheTTL         int  `json:"CACHE_TTL"` // In minutes

	MaxCacheSize int `json:"MAX_CACHE_SIZE"` // In MB, for max cached exhausted/metadata files(plus remote-raw if applicable), 0 means no limit
}

// NewWebPConfig default configuration
func NewWebPConfig() *WebpConfig {
	return &WebpConfig{
		Host:          "0.0.0.0",
		Port:          "3333",
		ImgPath:       "./pics",
		Quality:       80,
		AllowedTypes:  []string{"jpg", "png", "jpeg", "bmp", "gif", "svg", "nef", "heic", "webp"},
		ConvertTypes:  []string{"webp"},
		ImageMap:      map[string]string{},
		ExhaustPath:   "./exhaust",
		MetadataPath:  "./metadata",
		RemoteRawPath: "./remote-raw",

		EnableWebP: false,
		EnableAVIF: false,
		EnableJXL:  false,

		EnableExtraParams:          false,
		ExtraParamsCropInteresting: "InterestingAttention",
		StripMetadata:              true,
		ReadBufferSize:             4096,
		Concurrency:                262144,
		DisableKeepalive:           false,
		CacheTTL:                   259200,

		MaxCacheSize: 0,
	}
}

func init() {
	flag.StringVar(&ConfigPath, "config", "config.json", "/path/to/config.json. (Default: ./config.json)")
	flag.BoolVar(&Prefetch, "prefetch", false, "Prefetch and convert image to WebP format.")
	flag.IntVar(&Jobs, "jobs", runtime.NumCPU(), "Prefetch thread, default is all.")
	flag.BoolVar(&DumpConfig, "dump-config", false, "Print sample config.json.")
	flag.BoolVar(&ShowVersion, "V", false, "Show version information.")
}

// LoadConfig file from config.json
func LoadConfig() {

	jsonObject, err := os.Open(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(jsonObject)
	_ = decoder.Decode(&Config)
	_ = jsonObject.Close()

	Config.ImageMap = parseImgMap(Config.ImageMap)

	if slices.Contains(Config.ConvertTypes, "webp") {
		Config.EnableWebP = true
	}
	if slices.Contains(Config.ConvertTypes, "avif") {
		Config.EnableAVIF = true
	}
	if slices.Contains(Config.ConvertTypes, "jxl") {
		Config.EnableJXL = true
	}

	// Read from ENV for override
	if os.Getenv("WEBP_HOST") != "" {
		Config.Host = os.Getenv("WEBP_HOST")
	}
	if os.Getenv("WEBP_PORT") != "" {
		Config.Port = os.Getenv("WEBP_PORT")
	}

	log.Debugln("Config init complete")
	log.Debugln("Config", Config)
}

func parseImgMap(imgMap map[string]string) map[string]string {
	var parsedImgMap = map[string]string{}
	httpRegexpMatcher := regexp.MustCompile(HttpRegexp)
	for uriMap, uriMapTarget := range imgMap {
		if httpRegexpMatcher.Match([]byte(uriMap)) || strings.HasPrefix(uriMap, "/") {
			// Valid
			parsedImgMap[uriMap] = uriMapTarget
		} else {
			// Invalid
			log.Warnf("IMG_MAP key '%s' does matches '%s' or starts with '/' - skipped", uriMap, HttpRegexp)
		}
	}
	return parsedImgMap
}

// ExtraParams struct
type ExtraParams struct {
	Width     int // in px
	Height    int // in px
	MaxWidth  int // in px
	MaxHeight int // in px
}
