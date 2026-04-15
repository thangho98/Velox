package imagemeta

import (
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bbrks/go-blurhash"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/storage"
	"golang.org/x/time/rate"

	// Register image decoders
	_ "golang.org/x/image/webp"
	_ "image/jpeg"
	_ "image/png"
)

type Service struct {
	repo        *repository.ImageMetadataRepo
	storage     *storage.ImageStorage
	httpClient  *http.Client
	tmdbRateLtd *rate.Limiter
	worker      *Worker
}

func NewService(
	repo *repository.ImageMetadataRepo,
	storage *storage.ImageStorage,
	httpClient *http.Client,
	tmdbRateLtd *rate.Limiter,
) *Service {
	s := &Service{
		repo:        repo,
		storage:     storage,
		httpClient:  httpClient,
		tmdbRateLtd: tmdbRateLtd,
	}
	s.worker = NewWorker(s)
	return s
}

// Worker returns the asynchronous worker instance for this service.
func (s *Service) Worker() *Worker {
	return s.worker
}

// Compute fetches (if remote), decodes, computes blurhash + dimensions for a
// raw stored path. Upserts into image_metadata. Idempotent.
func (s *Service) Compute(ctx context.Context, rawPath string) (*model.ImageMetadata, error) {
	if rawPath == "" {
		return nil, fmt.Errorf("empty rawPath")
	}

	if existing, _ := s.repo.Get(ctx, rawPath); existing != nil {
		return existing, nil
	}

	img, err := s.fetchImage(ctx, rawPath)
	if err != nil {
		return nil, fmt.Errorf("fetchImage: %w", err)
	}

	hash, w, h, err := computeHash(img)
	if err != nil {
		return nil, fmt.Errorf("computeHash: %w", err)
	}

	meta := &model.ImageMetadata{
		Path:     rawPath,
		Blurhash: hash,
		Width:    w,
		Height:   h,
	}

	if err := s.repo.Upsert(ctx, meta); err != nil {
		return nil, fmt.Errorf("upsert meta: %w", err)
	}

	return meta, nil
}

// SaveManual directly upserts metadata (useful for local uploads where hash is already computed).
func (s *Service) SaveManual(ctx context.Context, meta *model.ImageMetadata) error {
	return s.repo.Upsert(ctx, meta)
}

// InvalidatePaths removes cached metadata for the given paths. Compute() is
// idempotent (short-circuits when a row exists), so force-rescan needs to
// wipe rows first, then enqueue for recomputation.
func (s *Service) InvalidatePaths(ctx context.Context, paths []string) error {
	return s.repo.DeleteBatch(ctx, paths)
}

// Enqueue adds a path to the background worker queue.
func (s *Service) Enqueue(path string) {
	if s.worker != nil {
		s.worker.Enqueue(path)
	}
}

// Worker handles asynchronous blurhash computation.
type Worker struct {
	svc   *Service
	queue chan string
}

func NewWorker(svc *Service) *Worker {
	return &Worker{
		svc:   svc,
		queue: make(chan string, 256),
	}
}

func (w *Worker) Enqueue(path string) {
	select {
	case w.queue <- path:
	default:
		log.Printf("[imagemeta] dropped path (queue full): %s", path)
	}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-w.queue:
			if _, err := w.svc.Compute(ctx, p); err != nil {
				log.Printf("[imagemeta] compute failed for %s: %v", p, err)
			}
		}
	}
}

// ComputeBatch processes N paths with concurrency. Returns per-path errors.
func (s *Service) ComputeBatch(ctx context.Context, paths []string) map[string]error {
	errs := make(map[string]error)
	for _, p := range paths {
		if _, err := s.Compute(ctx, p); err != nil {
			errs[p] = err
		}
	}
	return errs
}

func (s *Service) fetchImage(ctx context.Context, rawPath string) (image.Image, error) {
	switch {
	case strings.HasPrefix(rawPath, "local://"):
		parts := strings.SplitN(strings.TrimPrefix(rawPath, "local://"), "/", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid local path: %s", rawPath)
		}
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		absPath := s.storage.AbsPath(parts[0], id, parts[2])

		f, err := os.Open(absPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		return img, err
	case strings.HasPrefix(rawPath, "/"):
		// TMDb relative path — rate-limited to respect TMDb CDN.
		return s.fetchTMDbImage(ctx, rawPath)
	case strings.HasPrefix(rawPath, "http://") || strings.HasPrefix(rawPath, "https://"):
		// External full URL (fanart.tv logos/thumbs, or anything else upstream enrichers stored verbatim).
		return s.fetchHTTPImage(ctx, rawPath)
	default:
		return nil, fmt.Errorf("unsupported path scheme: %s", rawPath)
	}
}

func (s *Service) fetchTMDbImage(ctx context.Context, relPath string) (image.Image, error) {
	if err := s.tmdbRateLtd.Wait(ctx); err != nil {
		return nil, err
	}
	// Fetch w185 — small, fast, plenty for blurhash (4x3 components).
	return s.fetchHTTPImage(ctx, "https://image.tmdb.org/t/p/w185"+relPath)
}

func (s *Service) fetchHTTPImage(ctx context.Context, url string) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned %d for %s", resp.StatusCode, url)
	}

	img, _, err := image.Decode(resp.Body)
	return img, err
}

func computeHash(img image.Image) (string, int, int, error) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Ensure we only use 4x3 (blurhash constraints are 1-9 for components)
	hash, err := blurhash.Encode(4, 3, img)
	if err != nil {
		log.Printf("blurhash.Encode error: %v", err)
	}
	return hash, w, h, err
}
