package ftps

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Client FTPS Client
type Client struct {
	UserName  string
	Password  string
	Host      string
	TLSConfig *tls.Config
}

// NewClient return new FTPS Client
func NewClient(user string, pass string, host string) *Client {
	return &Client{
		UserName: user,
		Password: pass,
		Host:     host,
	}
}

func NewClientWithTLSConfig(user string, pass string, host string, config *tls.Config) *Client {
	client := NewClient(user, pass, host)
	client.TLSConfig = config
	return client
}

func (c *Client) tlsConfig() *tls.Config {
	if c.TLSConfig != nil {
		return c.TLSConfig
	}
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

// Upload file to Server
func (c *Client) Upload(filePath string) error {

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Open file failed[%q]: %s", filePath, err)
	}
	defer f.Close()

	return c.UploadFile(filepath.Base(filePath), f)
}

// UploadFile file to Server
func (c *Client) UploadFile(remoteFilepath string, file *os.File) error {
	return c.UploadReader(remoteFilepath, file)
}

// UploadFile file to Server
func (c *Client) UploadReader(remoteFilepath string, source io.Reader) error {
	rawClient := &FTPS{
		TLSConfig: *c.tlsConfig(),
	}

	err := rawClient.Connect(c.Host, 21)
	if err != nil {
		return fmt.Errorf("Connect FTP failed: %#v", err)
	}

	err = rawClient.Login(c.UserName, c.Password)
	if err != nil {
		return fmt.Errorf("Auth FTP failed: %#v", err)
	}

	err = rawClient.StoreReader(remoteFilepath, source)
	if err != nil {
		return fmt.Errorf("Storefile FTP failed: %#v", err)
	}

	err = rawClient.Quit()
	if err != nil {
		return fmt.Errorf("Quit FTP failed: %#v", err)
	}

	return nil
}

// Download file from server
func (c *Client) Download(filePath string) error {

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return c.DownloadFile(file)
}

// DownloadFile file from server
func (c *Client) DownloadFile(file *os.File) error {
	return c.DownloadWriter(file)
}

// DownloadFile file from server
func (c *Client) DownloadWriter(writer io.Writer) error {

	rawClient := &FTPS{
		TLSConfig: *c.tlsConfig(),
	}

	err := rawClient.Connect(c.Host, 21)
	if err != nil {
		return fmt.Errorf("Connect FTP failed: %#v", err)
	}

	err = rawClient.Login(c.UserName, c.Password)
	if err != nil {
		return fmt.Errorf("Auth FTP failed: %#v", err)
	}

	entries, err := rawClient.List()
	if err != nil {
		return fmt.Errorf("FTP List Entry failed: %#v", err)
	}

	var serverFilePath string
	for _, e := range entries {
		if e.Type == EntryTypeFile && !strings.HasPrefix(e.Name, ".") {
			serverFilePath = e.Name
			break
		}
	}
	if serverFilePath == "" {
		return errors.New("FTP retrieve filename failed")
	}

	// download
	err = rawClient.RetrieveWriter(serverFilePath, writer)
	if err != nil {
		return fmt.Errorf("FTP download file is failed: %#v", err)
	}

	err = rawClient.Quit()
	if err != nil {
		return fmt.Errorf("Quit FTP failed: %#v", err)
	}

	return nil
}
