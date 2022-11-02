package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/form/v4"
)

// New initializes the server with its default routers.
func New(opts ...Option) (*Server, error) {
	sbl, err := pongo2.NewSandboxedFilesystemLoader("theme/p2")
	if err != nil {
		return nil, err
	}

	s := Server{
		r:     chi.NewRouter(),
		n:     &http.Server{},
		f:     form.NewDecoder(),
		forms: make(map[string]Form),
		tmpls: pongo2.NewSet("html", sbl),
	}

	// Apply options
	for _, o := range opts {
		o(&s)
	}

	pongo2.RegisterFilter("key", s.filterGetValueByKey)
	pongo2.RegisterFilter("index", s.filterGetValueAtIndex)
	s.loadForms()
	s.tmpls.Debug = true

	s.r.Use(middleware.Logger)
	s.r.Use(middleware.Heartbeat("/healthz"))
	s.r.Use(s.checkStorage)

	s.fileServer(s.r, "/static", http.Dir("theme/static"))
	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/public/bigboard", http.StatusSeeOther)
	})

	s.r.Get("/public/bigboard", s.viewBigBoard)

	s.r.Route("/admin", func(r chi.Router) {
		r.Get("/", s.viewAdminLanding)
		r.Post("/", s.submitAdminLanding)
		r.Get("/form/{form}/{id}", s.viewForm)
		r.Post("/form/{form}/{id}", s.submitForm)
		r.Get("/formset/{form}", s.viewFormSet)
	})

	if s.kv == nil {
		s.l.Error("KV is nil!")
		return nil, errors.New("storage is undefined")
	}
	return &s, nil
}

// Serve binds, initializes the mux, and serves forever.
func (s *Server) Serve(bind string) error {
	s.l.Info("HTTP is starting")
	s.n.Addr = bind
	s.n.Handler = s.r
	return s.n.ListenAndServe()
}

// Shutdown requests the underlying server gracefully cease operation.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.n.Shutdown(ctx)
}

func (s *Server) checkStorage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.kv.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Fatal Error: Storage service is unavailable: %v (%T)", err, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) templateErrorHandler(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "Error while rendering template: %s\n", err)
}

func (s *Server) fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func (s *Server) doTemplate(w http.ResponseWriter, r *http.Request, tmpl string, ctx pongo2.Context) {
	if ctx == nil {
		ctx = pongo2.Context{}
	}
	t, err := s.tmpls.FromCache(tmpl)
	if err != nil {
		s.templateErrorHandler(w, err)
		return
	}
	if err := t.ExecuteWriter(ctx, w); err != nil {
		s.templateErrorHandler(w, err)
	}
}
