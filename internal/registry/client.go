package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net"

	"github.com/jdx/go-netrc"
	v1 "github.com/pbufio/pbuf-cli/gen/pbuf-registry/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// NewInsecureClient generates insecure grpc client
func NewInsecureClient(config *model.Config, netrcAuth *netrc.Netrc) v1.RegistryClient {
	addr := canonicalizeAddr(config.Registry.Addr)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	perRPCCredentials := credentialsFromNetrc(getRegistryHost(addr), netrcAuth)
	if perRPCCredentials != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(perRPCCredentials))
	}

	grpcClient, _ := grpc.NewClient(addr, opts...)

	return v1.NewRegistryClient(grpcClient)
}

// NewSecureClient generates secure grpc client
// Should use TLS to secure the connection
func NewSecureClient(config *model.Config, netrcAuth *netrc.Netrc) v1.RegistryClient {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("failed to load system cert pool: %v", err)
	}

	addr := canonicalizeAddr(config.Registry.Addr)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{
				RootCAs: certPool,
			}),
		),
	}

	perRPCCredentials := credentialsFromNetrc(getRegistryHost(addr), netrcAuth)
	if perRPCCredentials != nil {
		opts = append(opts, grpc.WithPerRPCCredentials(perRPCCredentials))
	}

	grpcClient, _ := grpc.NewClient(addr, opts...)

	return v1.NewRegistryClient(grpcClient)
}

// canonicalizeAddr check has the address port or not
// and add the 6777 port by default
func canonicalizeAddr(addr string) string {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return net.JoinHostPort(addr, "6777")
	}

	return addr
}

// getRegistryHost returns the registry host
func getRegistryHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	return host
}

type netrcCredential struct {
	token string
}

func (n netrcCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": n.token,
	}, nil
}

func (n netrcCredential) RequireTransportSecurity() bool {
	return false
}

// credentialsFromNetrc returns the credentials from netrc
func credentialsFromNetrc(host string, netrcAuth *netrc.Netrc) credentials.PerRPCCredentials {
	if netrcAuth == nil {
		return nil
	}

	machine := netrcAuth.Machine(host)
	if machine == nil {
		return nil
	}

	return &netrcCredential{
		token: machine.Get("token"),
	}
}
