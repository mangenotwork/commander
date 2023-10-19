package proxy

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/slve_linux/dao"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"time"
)

// HttpProxyServer RunServer 启动 http/s 代理服务
func HttpProxyServer(port, name string) {
	logger.Info("启动http/s代理服务")
	go func() {
		cert, err := genCertificate()
		if err != nil {
			logger.Error("启动http/s代理服务 失败 : ", err)
			return
		}
		server := &http.Server{
			Addr:      "0.0.0.0:" + port,
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// 是否停止
				proxyObj, err := new(dao.DaoHttpsProxy).Get(name)
				if err == nil {
					if proxyObj.IsClose == "1" {
						http.Error(w, "代理已关闭!", http.StatusInternalServerError)
						return
					}
				}

				// 是否删除, 删除与停止的区别在于，如果该代理有删除标识后slave不会去恢复这个代理
				if proxyObj.IsDel == "1" {
					http.Error(w, "代理已删除!", http.StatusInternalServerError)
					return
				}

				if r.Method == http.MethodConnect {
					destConn, err := net.DialTimeout("tcp", r.Host, 60*time.Second)
					if err != nil {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
						return
					}
					w.WriteHeader(http.StatusOK)
					hijacker, ok := w.(http.Hijacker)
					if !ok {
						http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
						return
					}

					clientConn, _, err := hijacker.Hijack()
					if err != nil {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
					}
					go io.Copy(clientConn, destConn)
					go io.Copy(destConn, clientConn)
				} else {
					res, err := http.DefaultTransport.RoundTrip(r)
					if err != nil {
						http.Error(w, err.Error(), http.StatusServiceUnavailable)
						return
					}
					defer res.Body.Close()
					for k, vv := range res.Header {
						for _, v := range vv {
							w.Header().Add(k, v)
						}
					}
					var bodyBytes []byte
					bodyBytes, _ = ioutil.ReadAll(res.Body)
					res.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
					w.WriteHeader(res.StatusCode)
					io.Copy(w, res.Body)
					res.Body.Close()
				}
			}),
		}
		err = server.ListenAndServe()
		if err != nil {
			logger.Error("启动http/s代理服务 失败 : ", err)
			return
		}
	}()
}

func genCertificate() (cert tls.Certificate, err error) {
	rawCert, rawKey, err := generateKeyPair()
	if err != nil {
		return
	}
	return tls.X509KeyPair(rawCert, rawKey)

}

func generateKeyPair() (rawCert, rawKey []byte, err error) {
	// Create private key and self-signed certificate
	// Adapted from https://golang.org/src/crypto/tls/generate_cert.go

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}
	validFor := time.Hour * 24 * 365 * 10 // ten years
	notBefore := time.Now()
	notAfter := notBefore.Add(validFor)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"ManGe-commander"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return
	}

	rawCert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	rawKey = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return
}
