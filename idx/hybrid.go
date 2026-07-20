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

	"github.com/advancevillage/3rd/logx"
	"github.com/tencentyun/cos-go-sdk-v5"
)

const (
	hybridSearchPath = "/datasetquery/hybridsearch"
)

func NewDocSearchRequest(dataset, text string, opts ...SearchOption) HybridSearchRequest {
	o := defaultSearchOption
	for _, opt := range opts {
		opt.Apply(&o)
	}
	return HybridSearchRequest{
		DatasetName:    dataset,
		Mode:           o.mode,
		Templates:      TemplateDocSearch,
		SearchText:     text,
		Limit:          o.limit,
		MatchThreshold: o.matchThreshold,
		Filter:         o.filter,
	}
}

func NewImageSearchRequest(dataset, text string, opts ...SearchOption) HybridSearchRequest {
	o := defaultSearchOption
	for _, opt := range opts {
		opt.Apply(&o)
	}
	return HybridSearchRequest{
		DatasetName:    dataset,
		Mode:           o.mode,
		Templates:      TemplateImageSearch,
		SearchText:     text,
		Limit:          o.limit,
		MatchThreshold: o.matchThreshold,
		Filter:         o.filter,
	}
}

// idx://ak:sk@appid/region
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
	return NewHybridSearcher(ctx, logger, opts...)
}

func NewHybridSearcher(ctx context.Context, logger logx.ILogger, opt ...Option) (HybridSearcher, error) {
	opts := defaultOption
	for _, o := range opt {
		o.Apply(&opts)
	}
	endpoint := fmt.Sprintf("https://%s.ci.%s.myqcloud.com", opts.appId, opts.region)
	endpoint = strings.TrimRight(endpoint, "/")
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("idx: invalid endpoint: %w", err)
	}
	client := &http.Client{
		Timeout: opts.timeout,
		Transport: &cos.AuthorizationTransport{
			SecretID:  opts.ak,
			SecretKey: opts.sk,
		},
	}
	return &txHybridSearcher{opts: opts, logger: logger, endpoint: endpoint, client: client}, nil
}

func (c *txHybridSearcher) HybridSearch(ctx context.Context, req HybridSearchRequest) (*HybridSearchResponse, error) {
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
		c.logger.Errorw(ctx, "idx hybrid search failed", "err", err, "dataset", req.DatasetName)
		return nil, err
	}
	defer resp.Body.Close()

	replyBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var reply HybridSearchResponse
	if err := json.Unmarshal(replyBody, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}
