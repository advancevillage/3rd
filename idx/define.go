package idx

import (
	"context"
	"net/http"

	"github.com/advancevillage/3rd/logx"
)

const (
	ModePic  SearchMode = "pic"
	ModeText SearchMode = "text"

	TemplateImageSearch SearchTemplate = "ImageSearch"
	TemplateDocSearch   SearchTemplate = "DocSearch"
)

type (
	SearchMode     string
	SearchTemplate string
)

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
	ImageResult []ImageSearchResult `json:"ImageResult,omitempty"`
	DocResult   []DocSearchResult   `json:"DocResult,omitempty"`
}

type ImageSearchResult struct {
	URI   string `json:"URI"`
	Score int    `json:"Score"`
}

type DocSearchResult struct {
	URI   string `json:"URI"`
	Score int    `json:"Score"`
}

type txHybridSearcher struct {
	opts     option
	logger   logx.ILogger
	endpoint string
	client   *http.Client
}

var _ HybridSearcher = (*txHybridSearcher)(nil)
