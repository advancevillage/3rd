//author: richard
package notices

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)


func NewTlsEmail(host string, username string, password string) *Email {
	return &Email{
		host: host,
		port: 465,
		username: username,
		password: password,
	}
}

func (e *Email) SendMailUsingTLS(to []string, alias string, subject string, body []byte) (err error) {
	//auth
	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	//创建TLS 加密链接
	conn , err := tls.Dial("tcp", fmt.Sprintf("%s:%d", e.host, e.port), nil)
	if err != nil {
		return err
	}
	//基于加密链接创建smtp客户端
	client, err := smtp.NewClient(conn, e.host)
	if err != nil {
		return err
	}
	//校验smtp客户端
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				//关闭Smtp客户端连接
				_ = client.Close()
				return err
			}
		}
	} else {
		//关闭Smtp客户端连接
		_ = client.Close()
		return errors.New("auth is nil")
	}
	//设置发件人
	if err = client.Mail(e.username); err != nil {
		_ = client.Close()
		return err
	}
	//设置收件人
	for i := range to {
		if err = client.Rcpt(to[i]); err != nil {
			_ = client.Close()
			return err
		}
	}
	//设置邮件协议内容
	w, err := client.Data()
	if err != nil {
		_ = client.Close()
		return err
	}
	//设置邮件 start
	message := "To: " + strings.Replace(strings.Trim(fmt.Sprint(to), "[]"), " ", ",", -1) + "\n" +
		"From: " + alias + "<" + e.username + ">\n" +
		"Subject: " + subject + "\n" +
		"Content-Type: text/html; charset=utf-8" + "\n" +
		"Content-Transfer-Encoding: base64" + "\n\n" +
		base64.StdEncoding.EncodeToString(body)

	_, err = w.Write([]byte(message))
	if err != nil {
		_ = client.Close()
		return err
	}
	//end
	err = w.Close()
	if err != nil {
		_ = client.Close()
		return err
	}
	//停止连接
	return client.Quit()
}
