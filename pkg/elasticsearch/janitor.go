package es

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/kubed/pkg/config"
	elastic "gopkg.in/olivere/elastic.v3"
)

type Janitor struct {
	Spec     config.ElasticsearchSpec
	AuthInfo *config.JanitorAuthInfo
	TTL      time.Duration
}

func (j *Janitor) Cleanup() error {
	var httpClient *http.Client

	if j.AuthInfo != nil {
		mTLSConfig := &tls.Config{}
		if j.AuthInfo.CACertData != nil {
			certs := x509.NewCertPool()
			certs.AppendCertsFromPEM(j.AuthInfo.CACertData)
			mTLSConfig.RootCAs = certs
			if j.AuthInfo.ClientCertData != nil && j.AuthInfo.ClientKeyData != nil {
				cert, err := tls.X509KeyPair(j.AuthInfo.ClientCertData, j.AuthInfo.ClientKeyData)
				if err == nil {
					mTLSConfig.Certificates = []tls.Certificate{cert}
				}
			}
		}

		if j.AuthInfo.InsecureSkipVerify {
			mTLSConfig.InsecureSkipVerify = true
		}

		// https://github.com/golang/go/blob/eca45997dfd6cd14a59fbdea2385f6648a0dc786/src/net/http/transport.go#L40
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       mTLSConfig,
		}
		httpClient = &http.Client{Transport: tr}
	}

	client, err := elastic.NewClient(
		elastic.SetHttpClient(httpClient),
		// elastic.SetSniff(false),
		elastic.SetURL(j.Spec.Endpoint),
	)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	oldDate := now.Add(-j.TTL)

	// how many index should we check to delete? I set it to 7
	for i := 1; i <= 7; i++ {
		date := oldDate.AddDate(0, 0, -i)
		prefix := fmt.Sprintf("%s%s", j.Spec.LogIndexPrefix, date.Format("2006.01.02"))

		if _, err := client.Search(prefix).Do(); err == nil {
			if _, err := client.DeleteIndex(prefix).Do(); err != nil {
				log.Errorln(err)
				return err
			}
			log.Debugf("Index [%s] deleted", prefix)
		}
	}
	log.Debugf("ElasticSearch cleanup process complete")
	return nil
}
