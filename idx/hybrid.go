package idx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const (
	ModePic  SearchMode = "pic"
	ModeText SearchMode = "text"

	TemplateImageSearch SearchTemplate = "ImageSearch"
	TemplateDocSearch   SearchTemplate = "DocSearch"

	hybridSearchPath = "/datasetquery/hybridsearch"
)

type SearchMode string
type SearchTemplate string

type HybridSearcher interface {
	HybridSearch(ctx context.Context, req HybridSearchRequest) (*HybridSearchResponse, error)
}

type HybridSearchRequest struct {
	DatasetName    string         `json:"DatasetName"`
	Mode           SearchMode     `json:"Mode,omitempty"`
	Templates      SearchTemplate `json:"Templates"`
	SearchURIs     []string       `json:"SearchURIs,omitempty"`
	SearchText     string         `json:"SearchText,omitempty"`
	Limit          int            `json:"Limit,omitempty"`
	MatchThreshold int            `json:"MatchThreshold,omitempty"`
	Filter         Filter         `json:"Filter,omitempty"`
}

type HybridSearchResponse struct {
	RequestId   string              `json:"RequestId"`
	ImageResult []ImageSearchResult `json:"ImageResult,omitempty"`
	DocResult   []DocSearchResult   `json:"DocResult,omitempty"`
}

type ImageSearchResult struct {
	URI   string `json:"URI"`
	Score int    `json:"Score"`
}

type DocSearchResult struct {
	URI       string            `json:"URI"`
	Text      string            `json:"Text"`
	Score     int               `json:"Score"`
	TextPage  int               `json:"TextPage"`
	ImageUrls map[string]string `json:"ImageUrls,omitempty"`
}

type txHybridSearcher struct {
	opts     option
	logger   logx.ILogger
	endpoint string
	client   *http.Client
}

var _ HybridSearcher = (*txHybridSearcher)(nil)

func NewDocSearchRequest(dataset, text string) HybridSearchRequest {
	return HybridSearchRequest{
		DatasetName: dataset,
		Mode:        ModeText,
		Templates:   TemplateDocSearch,
		SearchText:  text,
	}
}

func NewImageSearchRequest(dataset, text string) HybridSearchRequest {
	return HybridSearchRequest{
		DatasetName: dataset,
		Mode:        ModeText,
		Templates:   TemplateImageSearch,
		SearchText:  text,
	}
}

// idx://ak:sk@appid/region?token=xxx&timeout=30&endpoint=https://appid.ci.ap-shanghai.myqcloud.com
func NewHybridSearchClient(ctx context.Context, logger logx.ILogger, dsn string) (HybridSearcher, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "idx" {
		return nil, fmt.Errorf("idx: invalid scheme: %s", u.Scheme)
	}
	ak := u.User.Username()
	sk, ok := u.User.Password()
	if !ok || ak == "" || sk == "" {
		return nil, fmt.Errorf("idx: invalid ak or sk")
	}
	appId := u.Host
	region := strings.TrimPrefix(u.Path, "/")
	opts := []Option{
		WithSecret(ak, sk),
		WithAppId(appId),
		WithRegion(region),
	}
	if token := u.Query().Get("token"); token != "" {
		opts = append(opts, WithSessionToken(token))
	}
	if endpoint := u.Query().Get("endpoint"); endpoint != "" {
		opts = append(opts, WithEndpoint(endpoint))
	}
	if timeout := u.Query().Get("timeout"); timeout != "" {
		d, err := parseTimeout(timeout)
		if err != nil {
			return nil, fmt.Errorf("idx: invalid timeout: %s", timeout)
		}
		opts = append(opts, WithTimeout(d))
	}
	return NewHybridSearcher(ctx, logger, opts...)
}

func NewHybridSearcher(ctx context.Context, logger logx.ILogger, opt ...Option) (HybridSearcher, error) {
	opts := defaultOption
	for _, o := range opt {
		o.Apply(&opts)
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	endpoint := opts.endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.ci.%s.myqcloud.com", opts.appId, opts.region)
	}
	endpoint = strings.TrimRight(endpoint, "/")
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("idx: invalid endpoint: %w", err)
	}
	client := opts.httpClient
	if client == nil {
		client = &http.Client{
			Timeout: opts.timeout,
			Transport: &cos.AuthorizationTransport{
				SecretID:     opts.ak,
				SecretKey:    opts.sk,
				SessionToken: opts.token,
			},
		}
	}
	return &txHybridSearcher{opts: opts, logger: logger, endpoint: endpoint, client: client}, nil
}

func (c *txHybridSearcher) HybridSearch(ctx context.Context, req HybridSearchRequest) (*HybridSearchResponse, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+hybridSearchPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		if c.logger != nil {
			c.logger.Errorw(ctx, "idx hybrid search failed", "err", err, "dataset", req.DatasetName)
		}
		return nil, err
	}
	defer resp.Body.Close()

	replyBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("idx: hybrid search status %d: %s", resp.StatusCode, string(replyBody))
	}
	var reply HybridSearchResponse
	if err := json.Unmarshal(replyBody, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (r HybridSearchRequest) validate() error {
	if r.DatasetName == "" {
		return fmt.Errorf("idx: dataset name is required")
	}
	switch r.Templates {
	case TemplateImageSearch, TemplateDocSearch:
	default:
		return fmt.Errorf("idx: invalid templates: %s", r.Templates)
	}
	if r.Mode == "" {
		r.Mode = ModePic
	}
	switch r.Mode {
	case ModeText:
		if r.SearchText == "" {
			return fmt.Errorf("idx: search text is required")
		}
	case ModePic:
		if len(r.SearchURIs) == 0 {
			return fmt.Errorf("idx: search uris is required")
		}
	default:
		return fmt.Errorf("idx: invalid mode: %s", r.Mode)
	}
	if r.Templates == TemplateDocSearch && r.Mode != ModeText {
		return fmt.Errorf("idx: doc search requires text mode")
	}
	if r.Limit < 0 || r.Limit > 100 {
		return fmt.Errorf("idx: limit must be in [1, 100]")
	}
	if r.MatchThreshold < 0 || r.MatchThreshold > 100 {
		return fmt.Errorf("idx: match threshold must be in [0, 100]")
	}
	return nil
}

func parseTimeout(s string) (time.Duration, error) {
	if strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' }) < 0 {
		return time.ParseDuration(s + "s")
	}
	return time.ParseDuration(s)
}
