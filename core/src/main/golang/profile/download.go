package profile

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/Dreamacro/clash/adapters/inbound"
	"github.com/Dreamacro/clash/component/socks5"
	"github.com/Dreamacro/clash/config"
	"github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/tunnel"
)

const defaultFileMode = 0600

var client = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			if network != "tcp" && network != "tcp4" && network != "tcp6" {
				return nil, errors.New("Unsupported network type " + network)
			}

			client, server := net.Pipe()

			tunnel.Instance().Add(inbound.NewSocket(socks5.ParseAddr(address), server, constant.HTTP, constant.TCP))

			go func() {
				if ctx == nil || ctx.Done() == nil {
					return
				}

				<-ctx.Done()

				client.Close()
				server.Close()
			}()

			return client, nil
		},
	},
}

func DownloadAndCheck(url, output string) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	_, err = config.Parse(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, defaultFileMode)
}

func SaveAndCheck(data []byte, output string) error {
	_, err := config.Parse(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, defaultFileMode)
}