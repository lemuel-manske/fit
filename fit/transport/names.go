package transport

import (
	"fmt"

	"fit/fit/internal"
)

const DiscoveryExchange = "fit.discovery"

func RepositoryExchange(id internal.RepositoryID) string {
	return fmt.Sprintf("fit.repo.%s", id)
}
