package frontend

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-multierror"
	"github.com/jinayshah7/distributedSearchEngine/services/linkgraph/graph"
	"github.com/jinayshah7/distributedSearchEngine/services/textindexer/index"
)

const (
	searchEndpoint        = "/search"
	submitLinkEndpoint    = "/submit/site"
	ResultsPerSearch      = 10
	defaultResultsPerPage = 10
)

type LinkRepository interface {
	SaveLink(*graph.Link) error
}

type DocumentRepository interface {
	Search(query index.Query) (index.Iterator, error)
}

type Config struct {
	LinkRepository     LinkRepository
	DocumentRepository DocumentRepository
	ListenAddr         string
}

func (cfg *Config) validate() error {
	var err error
	if cfg.ListenAddr == "" {
		err = multierror.Append(err, errors.New("listen address has not been specified"))
	}
	if cfg.LinkRepository == nil {
		err = multierror.Append(err, errors.New("link repository URL has not been provided"))
	}
	if cfg.DocumentRepository == nil {
		err = multierror.Append(err, errors.New("document repository URL has not been provided"))
	}
	return err
}

type Service struct {
	cfg    Config
	router *mux.Router
}

func NewService(cfg Config) (*Service, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.New("front-end service: config validation failed: %w", err)
	}

	service := &Service{
		router: mux.NewRouter(),
		cfg:    cfg,
	}

	service.router.HandleFunc(searchEndpoint, svc.renderSearchResults).Methods("GET")
	service.router.HandleFunc(submitLinkEndpoint, svc.submitLink).Methods("POST")
	service.router.NotFoundHandler = http.HandlerFunc(svc.render404Page)
	return service, nil
}

func (svc *Service) Name() string { return "front-end" }

func (svc *Service) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", svc.cfg.ListenAddr)
	if err != nil {
		return err
	}
	defer func() { _ = l.Close() }()

	srv := &http.Server{
		Addr:    svc.cfg.ListenAddr,
		Handler: svc.router,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	log.Info("starting front-end server")
	if err = srv.Serve(l); err == http.ErrServerClosed {
		err = nil
	}

	return err
}

func (svc *Service) submitLink(w http.ResponseWriter, r *http.Request) {
	var msg string

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		link, err := url.Parse(r.Form.Get("link"))
		if err != nil || (link.Scheme != "http" && link.Scheme != "https") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		link.Fragment = ""
		newLink := &graph.Link{URL: link.String()}

		if err = svc.cfg.LinkRepository.SaveLink(newLink); err != nil {

			log.Errorf("could not upsert link into link graph %w", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (svc *Service) renderSearchResults(w http.ResponseWriter, r *http.Request) {
	searchTerms := r.URL.Query().Get("q")
	offset, _ := strconv.ParseUint(r.URL.Query().Get("offset"), 10, 64)

	matchedDocs, pagination, err := svc.runQuery(searchTerms, offset)
	if err != nil {
		svc.cfg.Logger.WithField("err", err).Errorf("search query execution failed")
		svc.renderSearchErrorPage(w, searchTerms)
		return
	}

}

func (svc *Service) runQuery(searchTerms string, offset uint64) ([]matchedDoc, error) {
	var query = index.Query{Type: index.QueryTypeMatch, Expression: searchTerms, Offset: offset}

	if strings.HasPrefix(searchTerms, `"`) && strings.HasSuffix(searchTerms, `"`) {

		searchTerms = strings.Trim(searchTerms, `"`)
	}

	resultIt, err := svc.cfg.DocumentRepository.Search(query)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resultIt.Close() }()

	matchedDocs := make([]matchedDoc, 0, ResultsPerSearch)
	for resultCount := 0; resultIt.Next() && resCount < svc.cfg.ResultsPerPage; resultCount++ {
		doc := resultIt.Document()
		matchedDocs = append(matchedDocs, matchedDoc{
			doc: doc,
		})
	}

	if err = resultIt.Error(); err != nil {
		return nil, nil, err
	}

	return matchedDocs, nil
}
