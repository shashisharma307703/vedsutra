package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/shashisharma307703/vedantam/internal/handler"
	"github.com/shashisharma307703/vedantam/internal/logger"
	"github.com/shashisharma307703/vedantam/internal/service"
)

type Server struct {
	router *chi.Mux
	port   string
}

func NewServer(
	port string,
	timeout time.Duration,
	orgHrv *handler.OrgHandler,
	classHrv *handler.ClassHandler,
	authService service.AuthService,
	authHandler *handler.AuthHandler,
	logger logger.Logger,
) *Server {
	r := chi.NewRouter()

	// Apply Core Infrastructure Middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(timeout))

	// Tenant middleware (must be applied early to all routes)
	r.Use(handler.TenantMiddleware(logger))

	// Register auth routes (public + protected)
	authHandler.RegisterRoutes(r)

	// Clean, Flattened API Routes Groupings
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth middleware to all API routes
		r.Use(handler.AuthMiddleware(authService, logger))

		// ORGANIZATIONS DOMAIN (Tenant Context)
		r.Route("/organizations", func(r chi.Router) {
			//r.Post("/", orgHrv.Create) // POST /api/v1/organizations
			//r.Get("/", orgHrv.List)    // GET /api/v1/organizations (With pagination/search filters)

			r.Route("/{orgId}", func(r chi.Router) {
			//	r.Get("/", orgHrv.Get)       // GET /api/v1/organizations/{orgId}
			//	r.Put("/", orgHrv.Update)    // PUT /api/v1/organizations/{orgId}
			//	r.Patch("/", orgHrv.Patch)   // PATCH /api/v1/organizations/{orgId}
			//	r.Delete("/", orgHrv.Delete) // DELETE /api/v1/organizations/{orgId}
			})
		})

		// CLASSES DOMAIN (Global / Super Admin Scope Only)
		r.Route("/classes", func(r chi.Router) {
			// Optional: Attach Super-Admin Auth Role validation middleware here
			// r.Use(auth.RequireSuperAdmin)

			//r.Post("/", classHrv.Create) // POST /api/v1/classes
			//r.Get("/", classHrv.List)    // GET /api/v1/classes (Global List/Search)

			r.Route("/{classId}", func(r chi.Router) {
				//r.Get("/", classHrv.Get)       // GET /api/v1/classes/{classId}
				//r.Put("/", classHrv.Update)    // PUT /api/v1/classes/{classId}
				//r.Patch("/", classHrv.Patch)   // PATCH /api/v1/classes/{classId}
				//r.Delete("/", classHrv.Delete) // DELETE /api/v1/classes/{classId}
			})
		})
	})

	return &Server{
		router: r,
		port:   port,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.port, s.router)
}
