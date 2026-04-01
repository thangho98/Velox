package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// BrowseService orchestrates library browsing and access checks.
type BrowseService struct {
	libraryRepo   *repository.LibraryRepo
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	userRepo      *repository.UserRepo
}

type libraryAccess struct {
	isAdmin bool
	allowed map[int64]struct{}
}

// NewBrowseService creates a new browse service.
func NewBrowseService(
	libraryRepo *repository.LibraryRepo,
	mediaRepo *repository.MediaRepo,
	mediaFileRepo *repository.MediaFileRepo,
	userRepo *repository.UserRepo,
) *BrowseService {
	return &BrowseService{
		libraryRepo:   libraryRepo,
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		userRepo:      userRepo,
	}
}

func (s *BrowseService) Browse(
	ctx context.Context,
	userID int64,
	isAdmin bool,
	libraryID int64,
	relativePath string,
) (*repository.BrowseResult, error) {
	access, err := s.loadLibraryAccess(ctx, userID, isAdmin)
	if err != nil {
		return nil, err
	}

	if libraryID == 0 {
		return s.browseLibrariesRoot(ctx, access)
	}

	library, err := s.libraryRepo.GetByID(ctx, libraryID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if !access.canAccess(libraryID) {
		return nil, ErrForbidden
	}

	if relativePath == "" && len(library.Paths) > 1 {
		return s.browseMultiRootLibrary(ctx, libraryID, library)
	}

	absDir, err := resolveBrowsePath(library.Paths, relativePath)
	if err != nil {
		return nil, err
	}

	return s.mediaFileRepo.BrowseFolders(ctx, libraryID, absDir, relativePath)
}

func (s *BrowseService) loadLibraryAccess(ctx context.Context, userID int64, isAdmin bool) (libraryAccess, error) {
	access := libraryAccess{
		isAdmin: isAdmin,
		allowed: make(map[int64]struct{}),
	}

	if isAdmin {
		return access, nil
	}

	ids, err := s.userRepo.GetLibraryIDs(ctx, userID)
	if err != nil {
		return access, err
	}

	for _, id := range ids {
		access.allowed[id] = struct{}{}
	}

	return access, nil
}

func (a libraryAccess) canAccess(libraryID int64) bool {
	if a.isAdmin {
		return true
	}

	_, ok := a.allowed[libraryID]
	return ok
}

func (s *BrowseService) browseLibrariesRoot(ctx context.Context, access libraryAccess) (*repository.BrowseResult, error) {
	libraries, err := s.libraryRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	folders := make([]repository.BrowseFolderItem, 0, len(libraries))
	for _, library := range libraries {
		if !access.canAccess(library.ID) {
			continue
		}

		folders = append(folders, repository.BrowseFolderItem{
			Name:   library.Name,
			Path:   fmt.Sprintf("lib:%d", library.ID),
			Poster: s.mediaRepo.FirstPosterByLibrary(ctx, library.ID),
		})
	}

	return &repository.BrowseResult{
		Path:    "",
		Parent:  "",
		Folders: folders,
	}, nil
}

func (s *BrowseService) browseMultiRootLibrary(
	ctx context.Context,
	libraryID int64,
	library *model.Library,
) (*repository.BrowseResult, error) {
	folders := make([]repository.BrowseFolderItem, 0, len(library.Paths))
	nameCounts := make(map[string]int)

	for index, rootPath := range library.Paths {
		base := filepath.Base(rootPath)
		nameCounts[base]++

		name := base
		if nameCounts[base] > 1 {
			name = fmt.Sprintf("%s-%d", base, nameCounts[base])
		}

		folders = append(folders, repository.BrowseFolderItem{
			Name:   name,
			Path:   fmt.Sprintf("root:%d", index),
			Poster: s.mediaRepo.FirstPosterByLibraryPath(ctx, libraryID, rootPath),
		})
	}

	return &repository.BrowseResult{
		LibraryID: libraryID,
		Path:      "",
		Parent:    "",
		Folders:   folders,
	}, nil
}

func resolveBrowsePath(paths []string, relativePath string) (string, error) {
	if len(paths) == 0 {
		return "", ErrLibraryHasNoRootPaths
	}

	rootPath := paths[0]
	subPath := relativePath

	if len(paths) > 1 && strings.HasPrefix(relativePath, "root:") {
		rest := strings.TrimPrefix(relativePath, "root:")
		rootIndexText := rest
		subPath = ""

		if slashIndex := strings.Index(rest, "/"); slashIndex > 0 {
			rootIndexText = rest[:slashIndex]
			subPath = rest[slashIndex+1:]
		}

		rootIndex, err := strconv.Atoi(rootIndexText)
		if err != nil || rootIndex < 0 || rootIndex >= len(paths) {
			return "", ErrInvalidBrowseRoot
		}

		rootPath = paths[rootIndex]
	}

	if subPath == "" {
		return rootPath, nil
	}

	return filepath.Join(rootPath, subPath), nil
}
