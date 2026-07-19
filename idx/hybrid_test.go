package idx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewHybridSearchClient(t *testing.T) {
	c, err := NewHybridSearchClient(context.Background(), nil, "idx://ak:sk@1234567890/ap-shanghai?token=tok&timeout=30")
	require.NoError(t, err)
	require.NotNil(t, c)

	_, err = NewHybridSearchClient(context.Background(), nil, "srch://ak:sk@1234567890/ap-shanghai")
	require.ErrorContains(t, err, "invalid scheme")

	_, err = NewHybridSearchClient(context.Background(), nil, "idx://ak:sk@1234567890/ap-guangzhou")
	require.ErrorContains(t, err, "unsupported region")
}

func TestRequestFactories(t *testing.T) {
	doc := NewDocSearchRequest("cards", "身份证")
	require.Equal(t, TemplateDocSearch, doc.Templates)
	require.Equal(t, ModeText, doc.Mode)
	require.Equal(t, "身份证", doc.SearchText)

	img := NewImageSearchRequest("images", "蓝色汽车")
	require.Equal(t, TemplateImageSearch, img.Templates)
	require.Equal(t, ModeText, img.Mode)
	require.Equal(t, "蓝色汽车", img.SearchText)
}

func TestHybridSearchValidatesRequest(t *testing.T) {
	tests := []struct {
		name string
		req  HybridSearchRequest
		err  string
	}{
		{name: "dataset", req: HybridSearchRequest{Mode: ModeText, Templates: TemplateDocSearch, SearchText: "x"}, err: "dataset name is required"},
		{name: "template", req: HybridSearchRequest{DatasetName: "d", Mode: ModeText, SearchText: "x"}, err: "invalid templates"},
		{name: "doc mode", req: HybridSearchRequest{DatasetName: "d", Mode: ModePic, Templates: TemplateDocSearch, SearchURIs: []string{"cos://b/a.jpg"}}, err: "doc search requires text mode"},
		{name: "text", req: HybridSearchRequest{DatasetName: "d", Mode: ModeText, Templates: TemplateImageSearch}, err: "search text is required"},
		{name: "pic", req: HybridSearchRequest{DatasetName: "d", Mode: ModePic, Templates: TemplateImageSearch}, err: "search uris is required"},
		{name: "limit", req: HybridSearchRequest{DatasetName: "d", Mode: ModeText, Templates: TemplateImageSearch, SearchText: "x", Limit: 101}, err: "limit must be"},
		{name: "threshold", req: HybridSearchRequest{DatasetName: "d", Mode: ModeText, Templates: TemplateImageSearch, SearchText: "x", MatchThreshold: 101}, err: "match threshold must be"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorContains(t, tt.req.validate(), tt.err)
		})
	}
}

func TestHybridSearchSendsSignedJSONRequest(t *testing.T) {
	var gotReq *http.Request
	var gotBody []byte
	transport := &cos.AuthorizationTransport{
		SecretID:     "ak",
		SecretKey:    "sk",
		SessionToken: "token",
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotReq = req
			var err error
			gotBody, err = io.ReadAll(req.Body)
			require.NoError(t, err)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"RequestId":"req-1","ImageResult":[{"URI":"cos://bucket/a.jpg","Score":91}]}`)),
				Request:    req,
			}, nil
		}),
	}
	client := &http.Client{Transport: transport}
	searcher, err := NewHybridSearcher(
		context.Background(),
		nil,
		WithSecret("ak", "sk"),
		WithSessionToken("token"),
		WithAppId("1234567890"),
		WithRegion("ap-shanghai"),
		WithEndpoint("https://1234567890.ci.ap-shanghai.myqcloud.com"),
		WithHTTPClient(client),
	)
	require.NoError(t, err)

	reply, err := searcher.HybridSearch(context.Background(), HybridSearchRequest{
		DatasetName:    "images",
		Mode:           ModeText,
		Templates:      TemplateImageSearch,
		SearchText:     "蓝色汽车",
		Limit:          10,
		MatchThreshold: 80,
		Filter:         And(In("MediaType", "image", "document"), Gt("Size", 123)),
	})
	require.NoError(t, err)
	require.Equal(t, "req-1", reply.RequestId)
	require.Len(t, reply.ImageResult, 1)
	require.Equal(t, "cos://bucket/a.jpg", reply.ImageResult[0].URI)

	require.Equal(t, http.MethodPost, gotReq.Method)
	require.Equal(t, "/datasetquery/hybridsearch", gotReq.URL.Path)
	require.Equal(t, "application/json", gotReq.Header.Get("Content-Type"))
	require.Equal(t, "application/json", gotReq.Header.Get("Accept"))
	require.NotEmpty(t, gotReq.Header.Get("Authorization"))
	require.Equal(t, "token", gotReq.Header.Get("x-cos-security-token"))

	var payload map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &payload))
	require.Equal(t, "images", payload["DatasetName"])
	require.Equal(t, "text", payload["Mode"])
	require.Equal(t, "ImageSearch", payload["Templates"])
	require.Equal(t, "蓝色汽车", payload["SearchText"])
}

func TestHybridSearchParsesDocResult(t *testing.T) {
	searcher, err := NewHybridSearcher(
		context.Background(),
		nil,
		WithSecret("ak", "sk"),
		WithAppId("1234567890"),
		WithRegion("ap-shanghai"),
		WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"RequestId":"req-2","DocResult":[{"URI":"cos://bucket/a.pdf","Text":"内容","Score":88,"TextPage":2,"ImageUrls":{"Image_0":"cos://bucket/img.jpg"}}]}`)),
				Request:    req,
			}, nil
		})}),
	)
	require.NoError(t, err)

	reply, err := searcher.HybridSearch(context.Background(), NewDocSearchRequest("cards", "内容"))
	require.NoError(t, err)
	require.Equal(t, "req-2", reply.RequestId)
	require.Len(t, reply.DocResult, 1)
	require.Equal(t, 2, reply.DocResult[0].TextPage)
	require.Equal(t, "cos://bucket/img.jpg", reply.DocResult[0].ImageUrls["Image_0"])
}
