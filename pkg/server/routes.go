package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juststeveking/package-api/pkg/packagist"
)

type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	r := gin.Default()
	s := &Server{
		httpServer: &http.Server{
			Handler: r,
			Addr:    ":3000", // default address, can be overridden in ListenAndServe
		},
	}

	r.GET("/", s.PackageHandler)

	return s
}

type packageDetail struct {
	Pkg    string             `json:"pkg"`
	Detail *packagist.Package `json:"detail"`
}

func (s *Server) RegisterRoutes() http.Handler {
	return s.httpServer.Handler
}

func generateETag(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	hash := md5.Sum(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

func (s *Server) PackageHandler(c *gin.Context) {
	vendor := c.Query("vendor")
	if vendor == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vendor query parameter is required"})
		return
	}

	pkgClient := packagist.NewPackagist(vendor)

	packages, err := pkgClient.FetchPackages()
	if err != nil {
		log.Printf("Error fetching packages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch packages"})
		return
	}

	var wg sync.WaitGroup
	detailCh := make(chan *packageDetail)
	errorCh := make(chan error)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fetchDetailWorker := func(ctx context.Context, pkg string) {
		defer wg.Done()
		detail, err := pkgClient.FetchDetails(pkg)
		if err != nil {
			select {
			case errorCh <- err:
			case <-ctx.Done():
			}
			return
		}
		select {
		case detailCh <- &packageDetail{pkg, detail}:
		case <-ctx.Done():
		}
	}

	for _, pkg := range packages.Packages {
		wg.Add(1)
		go fetchDetailWorker(ctx, pkg)
	}

	go func() {
		wg.Wait()
		close(detailCh)
		close(errorCh)
	}()

	response := make(map[string]*packagist.Package)

	// Collect details and build the response
	for detail := range detailCh {
		response[detail.Pkg] = detail.Detail
	}

	// Generate ETag for the response
	etag, err := generateETag(response)
	if err != nil {
		log.Printf("Error generating ETag: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate ETag"})
		return
	}

	// Check the If-None-Match header
	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch == etag {
		c.Status(http.StatusNotModified)
		return
	}

	// Set the ETag header
	c.Header("ETag", etag)

	c.JSON(http.StatusOK, response)
}

func (s *Server) ListenAndServe(port string) error {
	s.httpServer.Addr = ":" + port
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
