package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/forward"
	"github.com/shipyard/shipyard/controller/manager"
	"github.com/shipyard/shipyard/controller/middleware/access"
	"github.com/shipyard/shipyard/controller/middleware/audit"
	mAuth "github.com/shipyard/shipyard/controller/middleware/auth"
	"github.com/shipyard/shipyard/utils/auth"
	"github.com/shipyard/shipyard/utils/tlsutils"
	"github.com/shipyard/shipyard/utils/emitter"
	"golang.org/x/net/websocket"
)

type (
	Api struct {
		listenAddr         string
		manager            manager.Manager
		authWhitelistCIDRs []string
		enableCors         bool
		serverVersion      string
		allowInsecure      bool
		tlsCACertPath      string
		tlsCertPath        string
		tlsKeyPath         string
		dUrl               string
		fwd                *forward.Forwarder
		globalMux          http.Handler
	}

	ApiConfig struct {
		ListenAddr         string
		Manager            manager.Manager
		AuthWhiteListCIDRs []string
		EnableCORS         bool
		AllowInsecure      bool
		TLSCACertPath      string
		TLSCertPath        string
		TLSKeyPath         string
	}

	Credentials struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
)

var (
	// This can be passed down to handlers to have them broadcast messages
	// to our websocket connection `/ws/updates`
	WsEmmitter *emitter.Emitter
)

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

func NewApi(config ApiConfig) (*Api, error) {
	return &Api{
		listenAddr:         config.ListenAddr,
		manager:            config.Manager,
		authWhitelistCIDRs: config.AuthWhiteListCIDRs,
		enableCors:         config.EnableCORS,
		allowInsecure:      config.AllowInsecure,
		tlsCertPath:        config.TLSCertPath,
		tlsKeyPath:         config.TLSKeyPath,
		tlsCACertPath:      config.TLSCACertPath,
	}, nil
}

func (a *Api) Setup() (*http.ServeMux, error) {

	// Initialize our emitter.
	// We will use this to broadcast updates.
	WsEmmitter = emitter.NewEmitter()

	// This launches a goroutine and creates channels for all the plumbing
	var err error

	globalMux := http.NewServeMux()
	controllerManager := a.manager
	client := a.manager.DockerClient()

	// forwarder for swarm
	a.fwd, err = forward.New()
	if err != nil {
		return nil, err
	}

	u := client.URL

	// setup redirect target to swarm
	scheme := "http://"

	// check if TLS is enabled and configure if so
	if client.TLSConfig != nil {
		log.Debug("configuring ssl for swarm redirect")
		scheme = "https://"
		// setup custom roundtripper with TLS transport
		r := forward.RoundTripper(
			&http.Transport{
				TLSClientConfig: client.TLSConfig,
			})
		f, err := forward.New(r)
		if err != nil {
			return nil, err
		}

		a.fwd = f
	}

	a.dUrl = fmt.Sprintf("%s%s", scheme, u.Host)

	log.Debugf("configured docker proxy target: %s", a.dUrl)

	swarmRedirect := http.HandlerFunc(a.swarmRedirect)

	swarmHijack := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.swarmHijack(client.TLSConfig, a.dUrl, w, req)
	})

	apiRouter := mux.NewRouter()
	apiRouter.HandleFunc("/api/accounts", a.accounts).Methods("GET")
	apiRouter.HandleFunc("/api/accounts", a.saveAccount).Methods("POST")
	apiRouter.HandleFunc("/api/accounts/{username}", a.account).Methods("GET")
	apiRouter.HandleFunc("/api/accounts/{username}", a.deleteAccount).Methods("DELETE")
	//Project related routes
	apiRouter.HandleFunc("/api/projects", a.projects).Methods("GET")
	apiRouter.HandleFunc("/api/projects", a.saveProject).Methods("POST")
	apiRouter.HandleFunc("/api/projects/{id}", a.updateProject).Methods("PUT")
	apiRouter.HandleFunc("/api/projects/{id}", a.project).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{id}", a.deleteProject).Methods("DELETE")

	// Endpoints to handle images related to a given project
	apiRouter.HandleFunc("/api/projects/{projectId}/images", a.createImage).Methods("POST")
	apiRouter.HandleFunc("/api/projects/{projectId}/images", a.getImages).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/images/{imageId}", a.getImage).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/images/{imageId}", a.updateImage).Methods("PUT")
	apiRouter.HandleFunc("/api/projects/{projectId}/images/{imageId}", a.deleteImage).Methods("DELETE")

	//ILM related routes
	/*	apiRouter.HandleFunc("/api/ilm_images", a.getImages).Methods("GET")
		apiRouter.HandleFunc("/api/ilm_images", a.createImage).Methods("POST")
		apiRouter.HandleFunc("/api/ilm_images/{id}", a.updateImage).Methods("PUT")
		apiRouter.HandleFunc("/api/ilm_images/{id}", a.getImage).Methods("GET")
		apiRouter.HandleFunc("/api/ilm_images/{id}", a.deleteImage).Methods("DELETE")*/

	//Result Related routes
	apiRouter.HandleFunc("/api/projects/{projectId}/results", a.createResult).Methods("POST")
	apiRouter.HandleFunc("/api/projects/{projectId}/results", a.getResults).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/results/{resultId}", a.getResult).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/results/{resultId}", a.updateResult).Methods("PUT")
	apiRouter.HandleFunc("/api/projects/{projectId}/results/{resultId}", a.deleteResult).Methods("DELETE")
	//end Result related routes

	//Test related routes
	apiRouter.HandleFunc("/api/projects/{projectId}/tests", a.createTest).Methods("POST")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests", a.getTests).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}", a.getTest).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}", a.updateTest).Methods("PUT")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}", a.deleteTest).Methods("DELETE")

	//Build Related routes
	apiRouter.HandleFunc("/api/projects/{projectId}/builds", a.createAllBuilds).Methods("POST")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds", a.createBuild).Methods("POST")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds", a.getBuilds).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds/{buildId}", a.getBuild).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds/{buildId}/results", a.getBuildResults).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds/{buildId}", a.getBuildStatus).Methods("GET")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds/{buildId}/{action}", a.updateBuild).Methods("PUT")
	apiRouter.HandleFunc("/api/projects/{projectId}/tests/{testId}/builds/{buildId}", a.deleteBuild).Methods("DELETE")
	//end Build related routes

	//Provider related routes
	apiRouter.HandleFunc("/api/providers", a.createProvider).Methods("POST")
	apiRouter.HandleFunc("/api/providers", a.getProviders).Methods("GET")
	apiRouter.HandleFunc("/api/providers/{providerId}", a.getProvider).Methods("GET")
	apiRouter.HandleFunc("/api/providers/{providerId}", a.updateProvider).Methods("PUT")
	apiRouter.HandleFunc("/api/providers/{providerId}", a.deleteProvider).Methods("DELETE")
	apiRouter.HandleFunc("/api/providers/{providerId}/jobs", a.getJobsByProviderId).Methods("GET")
	apiRouter.HandleFunc("/api/providers/{providerId}/jobs", a.addJobToProviderId).Methods("POST")

	//Public Registry Proxy
	apiRouter.HandleFunc("/api/v1/search", a.dockerhubSearch).Methods("GET")
	apiRouter.HandleFunc("/api/v2/repositories/library/{imageName}/tags/", a.dockerhubTags).Methods("GET")

	apiRouter.HandleFunc("/api/roles", a.roles).Methods("GET")
	apiRouter.HandleFunc("/api/nodes", a.nodes).Methods("GET")
	apiRouter.HandleFunc("/api/nodes/{name}", a.node).Methods("GET")
	apiRouter.HandleFunc("/api/containers/{id}/scale", a.scaleContainer).Methods("POST")
	apiRouter.HandleFunc("/api/events", a.events).Methods("GET")
	apiRouter.HandleFunc("/api/events", a.purgeEvents).Methods("DELETE")
	apiRouter.HandleFunc("/api/registries", a.registries).Methods("GET")
	apiRouter.HandleFunc("/api/registries", a.addRegistry).Methods("POST")
	apiRouter.HandleFunc("/api/registries/{registryId}", a.registry).Methods("GET")
	apiRouter.HandleFunc("/api/registries/{registryId}", a.removeRegistry).Methods("DELETE")
	apiRouter.HandleFunc("/api/registries/{registryId}/repositories", a.repositories).Methods("GET")
	apiRouter.HandleFunc("/api/registries/{registryId}/repositories/{repo:.*}", a.repository).Methods("GET")
	apiRouter.HandleFunc("/api/registries/{registryId}/repositories/{repo:.*}", a.deleteRepository).Methods("DELETE")
	apiRouter.HandleFunc("/api/servicekeys", a.serviceKeys).Methods("GET")
	apiRouter.HandleFunc("/api/servicekeys", a.addServiceKey).Methods("POST")
	apiRouter.HandleFunc("/api/servicekeys", a.removeServiceKey).Methods("DELETE")
	apiRouter.HandleFunc("/api/webhookkeys", a.webhookKeys).Methods("GET")
	apiRouter.HandleFunc("/api/webhookkeys/{id}", a.webhookKey).Methods("GET")
	apiRouter.HandleFunc("/api/webhookkeys", a.addWebhookKey).Methods("POST")
	apiRouter.HandleFunc("/api/webhookkeys/{id}", a.deleteWebhookKey).Methods("DELETE")
	apiRouter.HandleFunc("/api/consolesession/{container}", a.createConsoleSession).Methods("GET")
	apiRouter.HandleFunc("/api/consolesession/{token}", a.consoleSession).Methods("GET")
	apiRouter.HandleFunc("/api/consolesession/{token}", a.removeConsoleSession).Methods("DELETE")

	// global handler
	globalMux.Handle("/", http.FileServer(http.Dir("controller/static")))

	auditExcludes := []string{
		"^/containers/json",
		"^/images/json",
		"^/api/events",
	}
	apiAuditor := audit.NewAuditor(controllerManager, auditExcludes)

	// api router; protected by auth
	apiAuthRouter := negroni.New()
	apiAuthRequired := mAuth.NewAuthRequired(controllerManager, a.authWhitelistCIDRs)
	apiAccessRequired := access.NewAccessRequired(controllerManager)
	apiAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAccessRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAuditor.HandlerFuncWithNext))
	apiAuthRouter.UseHandler(apiRouter)
	globalMux.Handle("/api/", apiAuthRouter)

	// account router ; protected by auth
	accountRouter := mux.NewRouter()
	accountRouter.HandleFunc("/account/changepassword", a.changePassword).Methods("POST")
	accountAuthRouter := negroni.New()
	accountAuthRequired := mAuth.NewAuthRequired(controllerManager, a.authWhitelistCIDRs)
	accountAuthRouter.Use(negroni.HandlerFunc(accountAuthRequired.HandlerFuncWithNext))
	accountAuthRouter.Use(negroni.HandlerFunc(apiAuditor.HandlerFuncWithNext))
	accountAuthRouter.UseHandler(accountRouter)
	globalMux.Handle("/account/", accountAuthRouter)

	// login handler; public
	loginRouter := mux.NewRouter()
	loginRouter.HandleFunc("/auth/login", a.login).Methods("POST")
	globalMux.Handle("/auth/", loginRouter)
	globalMux.Handle("/exec", websocket.Handler(a.execContainer))
	globalMux.HandleFunc("/ws/updates", a.projectUpdates)

	// hub handler; public
	hubRouter := mux.NewRouter()
	hubRouter.HandleFunc("/hub/webhook/{id}", a.hubWebhook).Methods("POST")
	globalMux.Handle("/hub/", hubRouter)

	// swarm
	swarmRouter := mux.NewRouter()
	// these are pulled from the swarm api code to proxy and allow
	// usage with the standard Docker cli
	m := map[string]map[string]http.HandlerFunc{
		"GET": {
			"/_ping":                          swarmRedirect,
			"/events":                         swarmRedirect,
			"/info":                           swarmRedirect,
			"/version":                        swarmRedirect,
			"/images/json":                    swarmRedirect,
			"/images/viz":                     swarmRedirect,
			"/images/search":                  swarmRedirect,
			"/images/get":                     swarmRedirect,
			"/images/{name:.*}/get":           swarmRedirect,
			"/images/{name:.*}/history":       swarmRedirect,
			"/images/{name:.*}/json":          swarmRedirect,
			"/containers/ps":                  swarmRedirect,
			"/containers/json":                swarmRedirect,
			"/containers/{name:.*}/export":    swarmRedirect,
			"/containers/{name:.*}/changes":   swarmRedirect,
			"/containers/{name:.*}/json":      swarmRedirect,
			"/containers/{name:.*}/top":       swarmRedirect,
			"/containers/{name:.*}/logs":      swarmRedirect,
			"/containers/{name:.*}/stats":     swarmRedirect,
			"/containers/{name:.*}/attach/ws": swarmHijack,
			"/exec/{execid:.*}/json":          swarmRedirect,
		},
		"POST": {
			"/auth":                         swarmRedirect,
			"/commit":                       swarmRedirect,
			"/build":                        swarmRedirect,
			"/images/create":                swarmRedirect,
			"/images/load":                  swarmRedirect,
			"/images/{name:.*}/push":        swarmRedirect,
			"/images/{name:.*}/tag":         swarmRedirect,
			"/containers/create":            swarmRedirect,
			"/containers/{name:.*}/kill":    swarmRedirect,
			"/containers/{name:.*}/pause":   swarmRedirect,
			"/containers/{name:.*}/unpause": swarmRedirect,
			"/containers/{name:.*}/rename":  swarmRedirect,
			"/containers/{name:.*}/restart": swarmRedirect,
			"/containers/{name:.*}/start":   swarmRedirect,
			"/containers/{name:.*}/stop":    swarmRedirect,
			"/containers/{name:.*}/wait":    swarmRedirect,
			"/containers/{name:.*}/resize":  swarmRedirect,
			"/containers/{name:.*}/attach":  swarmHijack,
			"/containers/{name:.*}/copy":    swarmRedirect,
			"/containers/{name:.*}/exec":    swarmRedirect,
			"/exec/{execid:.*}/start":       swarmHijack,
			"/exec/{execid:.*}/resize":      swarmRedirect,
		},
		"DELETE": {
			"/containers/{name:.*}": swarmRedirect,
			"/images/{name:.*}":     swarmRedirect,
		},
		"OPTIONS": {
			"": swarmRedirect,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			localRoute := route
			localFct := fct
			wrap := func(w http.ResponseWriter, r *http.Request) {
				if a.enableCors {
					writeCorsHeaders(w, r)
				}
				localFct(w, r)
			}
			localMethod := method

			// add the new route
			swarmRouter.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(wrap)
			swarmRouter.Path(localRoute).Methods(localMethod).HandlerFunc(wrap)
		}
	}

	swarmAuthRouter := negroni.New()
	swarmAuthRequired := mAuth.NewAuthRequired(controllerManager, a.authWhitelistCIDRs)
	swarmAccessRequired := access.NewAccessRequired(controllerManager)
	swarmAuthRouter.Use(negroni.HandlerFunc(swarmAuthRequired.HandlerFuncWithNext))
	swarmAuthRouter.Use(negroni.HandlerFunc(swarmAccessRequired.HandlerFuncWithNext))
	swarmAuthRouter.Use(negroni.HandlerFunc(apiAuditor.HandlerFuncWithNext))
	swarmAuthRouter.UseHandler(swarmRouter)
	globalMux.Handle("/containers/", swarmAuthRouter)
	globalMux.Handle("/_ping", swarmAuthRouter)
	globalMux.Handle("/commit", swarmAuthRouter)
	globalMux.Handle("/build", swarmAuthRouter)
	globalMux.Handle("/events", swarmAuthRouter)
	globalMux.Handle("/version", swarmAuthRouter)
	globalMux.Handle("/images/", swarmAuthRouter)
	globalMux.Handle("/exec/", swarmAuthRouter)
	globalMux.Handle("/v1.14/", swarmAuthRouter)
	globalMux.Handle("/v1.15/", swarmAuthRouter)
	globalMux.Handle("/v1.16/", swarmAuthRouter)
	globalMux.Handle("/v1.17/", swarmAuthRouter)
	globalMux.Handle("/v1.18/", swarmAuthRouter)
	globalMux.Handle("/v1.19/", swarmAuthRouter)
	globalMux.Handle("/v1.20/", swarmAuthRouter)
	globalMux.Handle("/v1.21/", swarmAuthRouter)
	globalMux.Handle("/v1.22/", swarmAuthRouter)
	globalMux.Handle("/v1.23/", swarmAuthRouter)
	globalMux.Handle("/v1.24/", swarmAuthRouter)

	// check for admin user
	if _, err := controllerManager.Account("admin"); err == manager.ErrAccountDoesNotExist {
		// create roles
		acct := &auth.Account{
			Username:  "admin",
			Password:  "shipyard",
			FirstName: "Shipyard",
			LastName:  "Admin",
			Roles:     []string{"admin"},
		}
		if err := controllerManager.SaveAccount(acct); err != nil {
			log.Fatal(err)
		}
		log.Infof("created admin user: username: admin password: shipyard")
	}
	return globalMux, nil
}

func (a *Api) Run(globalMux http.Handler) error {

	log.Infof("controller listening on %s", a.listenAddr)

	s := &http.Server{
		Addr:    a.listenAddr,
		Handler: context.ClearHandler(globalMux),
	}

	var runErr error

	if a.tlsCertPath != "" && a.tlsKeyPath != "" {
		log.Infof("using TLS for communication: cert=%s key=%s",
			a.tlsCertPath,
			a.tlsKeyPath,
		)

		// setup TLS config
		var caCert []byte
		if a.tlsCACertPath != "" {
			ca, err := ioutil.ReadFile(a.tlsCACertPath)
			if err != nil {
				return err
			}

			caCert = ca
		}

		serverCert, err := ioutil.ReadFile(a.tlsCertPath)
		if err != nil {
			return err
		}

		serverKey, err := ioutil.ReadFile(a.tlsKeyPath)
		if err != nil {
			return err
		}

		tlsConfig, err := tlsutils.GetServerTLSConfig(caCert, serverCert, serverKey, a.allowInsecure)
		if err != nil {
			return err
		}

		s.TLSConfig = tlsConfig

		runErr = s.ListenAndServeTLS(a.tlsCertPath, a.tlsKeyPath)
	} else {
		runErr = s.ListenAndServe()
	}

	return runErr
}
