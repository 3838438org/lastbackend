package api

import (
	"github.com/gorilla/mux"
	"github.com/lastbackend/lastbackend/pkg/daemon/api/routes"
	"github.com/lastbackend/lastbackend/pkg/util/http"
)

func Listen(host string, port int) error {
	router := mux.NewRouter()
	router.Methods("OPTIONS").HandlerFunc(http.Headers)
	for _, route := range Routes {
		router.Handle(route.Path, http.Handle(route.Handler, route.Middleware...)).Methods(route.Method)
	}
	return http.Listen(host, port, router)
}

var Routes = []http.Route{
	{Path: "/session", Method: http.MethodPost, Handler: routes.SessionCreateH},

	// User handlers
	{Path: "/user", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.UserGetH},

	// Vendor handlers
	{Path: "/oauth/{vendor}", Method: http.MethodDelete, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.OAuthDisconnectH},
	{Path: "/oauth/{vendor}/{code}", Method: http.MethodPost, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.OAuthConnectH},

	// VCS handlers extends
	{Path: "/vcs/{vendor}/repos", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.VCSRepositoriesListH},
	{Path: "/vcs/{vendor}/branches", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.VCSBranchesListH},

	// Build handlers
	{Path: "/build", Method: http.MethodGet, Handler: routes.BuildListH},
	{Path: "/build", Method: http.MethodPost, Handler: routes.BuildCreateH},

	// Project handlers
	{Path: "/project", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ProjectListH},
	{Path: "/project", Method: http.MethodPost, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ProjectCreateH},
	{Path: "/project/{project}", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ProjectInfoH},
	//{ Path: "/project/{project}", Method: http.MethodPut, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ProjectUpdateH},
	{Path: "/project/{project}", Method: http.MethodDelete, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ProjectRemoveH},
	//{ Path: "/project/{project}/activity", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ProjectActivityListH},
	{Path: "/project/{project}/service", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ServiceListH},
	{Path: "/project/{project}/service/{service}", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ServiceInfoH},
	//{ Path: "/project/{project}/service/{service}", Method: http.MethodPut, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ServiceUpdateH},
	{Path: "/project/{project}/service/{service}", Method: http.MethodDelete, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.ServiceRemoveH},

	//{ Path: "/project/{project}/service/{service}/activity", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ServiceActivityListH},
	//{ Path: "/project/{project}/service/{service}/hook", Method: http.MethodPost, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ServiceHookCreateH},
	//{ Path: "/project/{project}/service/{service}/hook", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ServiceHookListH},
	//{ Path: "/project/{project}/service/{service}/hook/{hook}", Method: http.MethodDelete, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ServiceHookRemoveH},
	//{ Path: "/project/{project}/service/{service}/logs", Method: http.MethodGet, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.ServiceLogsH},

	// Deploy template/docker/source/repo
	//{ Path: "/deploy", Method: http.MethodPost, Middleware: []http.Middleware{http.Authenticate}, Handler: routes.routes.DeployH},

	// Hook routes.
	{Path: "/hook/{token}", Method: http.MethodPost, Handler: routes.HookExecuteH},

	// Docker routes.
	{Path: "/docker/repo/search", Method: http.MethodGet, Handler: routes.DockerRepositorySearchH},
	{Path: "/docker/repo/tags", Method: http.MethodGet, Handler: routes.DockerRepositoryTagListH},
}
