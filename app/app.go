package app

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"humpback-agent/config"
	"humpback-agent/model"
	"humpback-agent/pkg/utils"
	"humpback-agent/service"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func loadConfig(configPath string) (*config.AppConfig, error) {
	logrus.Info("Loading server config....")
	appConfig, err := config.NewAppConfig(configPath)
	if err != nil {
		return nil, err
	}

	logrus.Info("-----------------HUMPBACK AGENT CONFIG-----------------")
	logrus.Infof("API Bind: %s:%s", appConfig.APIConfig.HostIP, appConfig.APIConfig.Port)
	logrus.Infof("API Versions: %v", appConfig.APIConfig.Versions)
	logrus.Infof("API Middlewares: %v", appConfig.APIConfig.Middlewares)
	// logrus.Infof("API Access Token: %s", appConfig.APIConfig.AccessToken)
	logrus.Infof("Docker Host: %s", appConfig.DockerConfig.Host)
	logrus.Infof("Docker Version: %s", appConfig.DockerConfig.Version)
	logrus.Infof("Docker AutoNegotiate: %v", appConfig.DockerConfig.AutoNegotiate)
	logrus.Info("-------------------------------------------------------")
	return appConfig, nil
}

func initLogger(loggerConfig *config.LoggerConfig) error {
	logDir := filepath.Dir(loggerConfig.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   loggerConfig.LogFile,
		MaxSize:    loggerConfig.MaxSize / 1024 / 1024,
		MaxBackups: loggerConfig.MaxBackups,
		MaxAge:     loggerConfig.MaxAge,
		Compress:   loggerConfig.Compress,
	}

	logrus.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))
	level, err := logrus.ParseLevel(loggerConfig.Level)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)
	if loggerConfig.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	return nil
}

func Bootstrap(ctx context.Context) {
	configFile := flag.String("f", "./config.yaml", "application configuration file path.")
	// 解析命令行参数
	flag.Parse()

	logrus.Info("Humpback Agent starting....")
	appConfig, err := loadConfig(*configFile)
	if err != nil {
		logrus.Errorf("Load application config error, %s", err.Error())
		return
	}

	if err := initLogger(appConfig.LoggerConfig); err != nil {
		logrus.Errorf("Init application logger error, %s", err.Error())
		return
	}

	certBundle, token, err := RegisterWithMaster(appConfig)
	if err != nil {
		logrus.Errorf("Register with master error, %s", err.Error())
		return
	}

	agentService, err := service.NewAgentService(ctx, appConfig, certBundle, token)
	if err != nil {
		logrus.Errorf("Init application agent service error, %s", err.Error())
		return
	}

	defer func() {
		agentService.Shutdown(ctx)
		logrus.Info("Humpback Agent shutdown.")
	}()

	logrus.Info("Humpback Agent started.")
	utils.ProcessWaitForSignal(nil)
}

// RegisterWithMaster 向Master注册并获取证书
func RegisterWithMaster(appConfig *config.AppConfig) (*model.CertificateBundle, string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 第一次连接跳过验证
			},
		},
	}

	var hostIps []string
	if appConfig.HostIP != "" {
		hostIps = []string{appConfig.HostIP}
	} else {
		hostIps = utils.HostIPs()
	}

	// 创建注册请求
	reqBody := struct {
		IpAddress []string `json:"hostIPs"`
		Token     string   `json:"token"`
	}{IpAddress: hostIps, Token: appConfig.ServerConfig.RegisterToken}

	reqBytes, _ := json.Marshal(reqBody)
	fmt.Printf("Register request body: %s\n", string(reqBytes))
	url := fmt.Sprintf("https://%s/api/register", appConfig.ServerConfig.Host)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("registration failed: %s", string(body))
	}

	// 解析响应
	var regResp struct {
		CertPEM string `json:"certPem"`
		KeyPEM  string `json:"keyPem"`
		Token   string `json:"token"`
		CAPEM   string `json:"caPem"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return nil, "", fmt.Errorf("failed to parse registration response: %w", err)
	}

	// 创建证书包
	certBlock, _ := pem.Decode([]byte(regResp.CertPEM))
	if certBlock == nil {
		return nil, "", errors.New("invalid certificate format")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	keyBlock, _ := pem.Decode([]byte(regResp.KeyPEM))
	if keyBlock == nil {
		return nil, "", errors.New("invalid key format")
	}

	privKey, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse private key: %w", err)
	}

	caBlock, _ := pem.Decode([]byte(regResp.CAPEM))
	if caBlock == nil {
		return nil, "", errors.New("invalid CA certificate format")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	certBundle := &model.CertificateBundle{
		Cert:     cert,
		PrivKey:  privKey,
		CertPool: certPool,
		CertPEM:  []byte(regResp.CertPEM),
		KeyPEM:   []byte(regResp.KeyPEM),
	}

	currentToken := regResp.Token

	slog.Info("Worker registered with master successfully")
	return certBundle, currentToken, nil
}
