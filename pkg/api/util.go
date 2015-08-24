package api

import (
	"fmt"
)

func ToDomainName(domainClass, domainInstance string) string {
	return fmt.Sprintf("%s.%s", domainInstance, domainClass)
}
