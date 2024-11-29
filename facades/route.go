package facades

import "github.com/nd-tools/capyvel/router"

func Route() *router.Router {
	return &router.RouterManager
}
