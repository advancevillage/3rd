//author: richard
package https

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func NewRequest(headers map[string]string, timeout int64, retryCount uint) *Client {
	return &Client{
		headers:headers,
		timeout:timeout,
		retryCount: retryCount + 1,
	}
}

func (r *Client) GET(uri string,  params map[string]string, headers map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * time.Duration(r.timeout)}
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	//设置请求参数
	query := request.URL.Query()
	for k,v := range params {
		query.Add(k, v)
	}
	request.URL.RawQuery = query.Encode()
	//设置请求头 headers > config.Configure.Header
	for k,v := range headers {
		request.Header.Add(k,v)
	}
	for k,v := range r.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k,v)
		}
	}
	//发送请求
	var response *http.Response
	for i := uint(0); i < r.retryCount; i++ {
		response, err = client.Do(request)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	defer func() { err = response.Body.Close() }()
	//读取响应
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Client) POST(uri string, headers map[string]string, buf []byte) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * time.Duration(r.timeout)}
	request, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(buf))
	request.Header.Add("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	for k,v := range headers {
		request.Header.Add(k,v)
	}
	for k,v := range r.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k,v)
		}
	}
	var response *http.Response
	for i := uint(0); i < r.retryCount; i++ {
		response, err = client.Do(request)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	defer func() { err = response.Body.Close() }()
	//读取响应
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Client) PostForm(uri string, params map[string]string, headers map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * time.Duration(r.timeout)}
	form := url.Values{}
	for k, v := range params {
		form.Add(k, v)
	}
	request, err := http.NewRequest(http.MethodGet, uri, nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	//设置请求头 headers > config.Configure.Header
	for k,v := range headers {
		request.Header.Add(k,v)
	}
	for k,v := range r.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k,v)
		}
	}
	//发送请求
	var response *http.Response
	for i := uint(0); i < r.retryCount; i++ {
		response, err = client.Do(request)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	defer func() { err = response.Body.Close() }()
	//读取响应
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

//@link: https://matt.aimonetti.net/posts/2013-07-golang-multipart-file-upload-example/
func (r *Client) Upload(uri string, params map[string]string, headers map[string]string, uploadFile string) ([]byte, error) {
	client := &http.Client{Timeout: time.Second * time.Duration(r.timeout)}
	file, err := os.Open(uploadFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(uploadFile))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	for k,v := range headers {
		request.Header.Add(k,v)
	}
	for k,v := range r.headers {
		if _, ok := headers[k]; ok {
			continue
		} else {
			request.Header.Add(k,v)
		}
	}
	//请求头
	request.Header.Set("Content-Type", writer.FormDataContentType())
	var response *http.Response
	for i := uint(0); i < r.retryCount; i++ {
		response, err = client.Do(request)
		if err != nil {
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}