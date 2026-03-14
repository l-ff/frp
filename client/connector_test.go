package client

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func TestDefaultConnectorImpl_getServerEndpoint_UseConfigPort(t *testing.T) {
	require := require.New(t)

	c := &defaultConnectorImpl{
		ctx: context.Background(),
		cfg: &v1.ClientCommonConfig{
			ServerAddr: "frps.example.com",
			ServerPort: 7000,
		},
		lookupSRV: func(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
			t.Fatal("lookupSRV should not be called when serverPort is non-zero")
			return "", nil, nil
		},
	}

	endpoint, err := c.getServerEndpoint()
	require.NoError(err)
	require.Equal("frps.example.com:7000", endpoint)
}

func TestDefaultConnectorImpl_getServerEndpoint_UseSRVWhenPortIsZero(t *testing.T) {
	require := require.New(t)

	c := &defaultConnectorImpl{
		ctx: context.Background(),
		cfg: &v1.ClientCommonConfig{
			ServerAddr: "_frp._tcp.example.com",
			ServerPort: 0,
		},
		lookupSRV: func(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
			require.Equal("", service)
			require.Equal("", proto)
			require.Equal("_frp._tcp.example.com", name)
			return "", []*net.SRV{
				{Target: "frps-1.example.com.", Port: 7100},
				{Target: "frps-2.example.com.", Port: 7200},
			}, nil
		},
	}

	endpoint, err := c.getServerEndpoint()
	require.NoError(err)
	require.Equal("frps-1.example.com:7100", endpoint)
}

func TestDefaultConnectorImpl_getServerEndpoint_SRVError(t *testing.T) {
	require := require.New(t)

	c := &defaultConnectorImpl{
		ctx: context.Background(),
		cfg: &v1.ClientCommonConfig{
			ServerAddr: "_frp._tcp.example.com",
			ServerPort: 0,
		},
		lookupSRV: func(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
			return "", nil, context.DeadlineExceeded
		},
	}

	_, err := c.getServerEndpoint()
	require.Error(err)
	require.Contains(err.Error(), "lookup SRV")
}

func TestDefaultConnectorImpl_getServerEndpoint_EmptySRVRecords(t *testing.T) {
	require := require.New(t)

	c := &defaultConnectorImpl{
		ctx: context.Background(),
		cfg: &v1.ClientCommonConfig{
			ServerAddr: "_frp._tcp.example.com",
			ServerPort: 0,
		},
		lookupSRV: func(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
			return "", nil, nil
		},
	}

	_, err := c.getServerEndpoint()
	require.Error(err)
	require.Contains(err.Error(), "no records")
}
