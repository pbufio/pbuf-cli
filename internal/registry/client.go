package registry

import (
	v1 "github.com/pbufio/pbuf-cli/gen/api/v1"
	"github.com/pbufio/pbuf-cli/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewInsecureClient generates insecure grpc client
func NewInsecureClient(config *model.Config) v1.RegistryClient {
	grpcClient, _ := grpc.Dial(
		config.Registry.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	return v1.NewRegistryClient(grpcClient)
}
