package service

import "errors"

var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrForbidden is returned when the caller lacks permission to access a resource.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidDisplayName is returned when a profile update omits display_name.
	ErrInvalidDisplayName = errors.New("display_name is required")

	// ErrInvalidBrowseRoot is returned when a library root selector is malformed.
	ErrInvalidBrowseRoot = errors.New("invalid root index")

	// ErrLibraryHasNoRootPaths is returned when a library has no configured root paths.
	ErrLibraryHasNoRootPaths = errors.New("library has no root paths")
)
