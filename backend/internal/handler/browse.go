package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

type BrowseHandler struct {
	libraryRepo   *repository.LibraryRepo
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	userRepo      *repository.UserRepo
}

type libraryAccess struct {
	isAdmin bool
	allowed map[int64]struct{}
}

func NewBrowseHandler(
	libraryRepo *repository.LibraryRepo,
	mediaRepo *repository.MediaRepo,
	mediaFileRepo *repository.MediaFileRepo,
	userRepo *repository.UserRepo,
) *BrowseHandler {
	return &BrowseHandler{
		libraryRepo:   libraryRepo,
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		userRepo:      userRepo,
	}
}

func (h *BrowseHandler) Browse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	access, err := h.loadLibraryAccess(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "checking library access")
		return
	}

	libraryIDStr := r.URL.Query().Get("library_id")
	relativePath := r.URL.Query().Get("path")

	if strings.Contains(relativePath, "..") {
		respondError(w, http.StatusBadRequest, "invalid path")
		return
	}

	if libraryIDStr == "" || libraryIDStr == "0" {
		result, err := h.browseLibrariesRoot(ctx, access)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "listing libraries")
			return
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	libraryID, err := strconv.ParseInt(libraryIDStr, 10, 64)
	if err != nil || libraryID == 0 {
		respondError(w, http.StatusBadRequest, "invalid library_id")
		return
	}

	library, err := h.libraryRepo.GetByID(ctx, libraryID)
	if err != nil {
		respondError(w, http.StatusNotFound, "library not found")
		return
	}

	if !access.canAccess(libraryID) {
		respondError(w, http.StatusForbidden, "no access to this library")
		return
	}

	if relativePath == "" && len(library.Paths) > 1 {
		result, err := h.browseMultiRootLibrary(ctx, libraryID, library)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, result)
		return
	}

	absDir, err := resolveBrowsePath(library.Paths, relativePath)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.mediaFileRepo.BrowseFolders(ctx, libraryID, absDir, relativePath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (h *BrowseHandler) loadLibraryAccess(ctx context.Context) (libraryAccess, error) {
	userID, isAdmin, _ := auth.UserFromContext(ctx)
	access := libraryAccess{
		isAdmin: isAdmin,
		allowed: make(map[int64]struct{}),
	}

	if isAdmin {
		return access, nil
	}

	ids, err := h.userRepo.GetLibraryIDs(ctx, userID)
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

func (h *BrowseHandler) browseLibrariesRoot(ctx context.Context, access libraryAccess) (*repository.BrowseResult, error) {
	libraries, err := h.libraryRepo.List(ctx)
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
			Poster: h.mediaRepo.FirstPosterByLibrary(ctx, library.ID),
		})
	}

	return &repository.BrowseResult{
		Path:    "",
		Parent:  "",
		Folders: folders,
	}, nil
}

func (h *BrowseHandler) browseMultiRootLibrary(ctx context.Context, libraryID int64, library *model.Library) (*repository.BrowseResult, error) {
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
			Poster: h.mediaRepo.FirstPosterByLibraryPath(ctx, libraryID, rootPath),
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
		return "", fmt.Errorf("library has no root paths")
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
			return "", fmt.Errorf("invalid root index")
		}

		rootPath = paths[rootIndex]
	}

	if subPath == "" {
		return rootPath, nil
	}

	return filepath.Join(rootPath, subPath), nil
}
