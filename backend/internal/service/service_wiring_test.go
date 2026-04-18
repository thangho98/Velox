package service

import (
	"context"
	"testing"

	"github.com/thawng/velox/internal/model"
)

type nilLibraryImagemetaStub struct{}

func (*nilLibraryImagemetaStub) ComputeBatch(context.Context, []string) map[string]error { return nil }
func (*nilLibraryImagemetaStub) InvalidatePaths(context.Context, []string) error         { return nil }

type nilMetadataImagemetaStub struct{}

func (*nilMetadataImagemetaStub) Enqueue(string) {}
func (*nilMetadataImagemetaStub) SaveManual(context.Context, *model.ImageMetadata) error {
	return nil
}

func TestSetImagemetaServiceIgnoresTypedNil(t *testing.T) {
	t.Parallel()

	var svc *nilLibraryImagemetaStub

	librarySvc := &LibraryService{}
	librarySvc.SetImagemetaService(svc)

	if librarySvc.imagemetaSvc != nil {
		t.Fatalf("expected typed-nil imagemeta service to be ignored")
	}
}

func TestSetImageMetaServiceIgnoresTypedNil(t *testing.T) {
	t.Parallel()

	var svc *nilMetadataImagemetaStub

	metadataSvc := &MetadataService{}
	metadataSvc.SetImageMetaService(svc)

	if metadataSvc.imagemetaSvc != nil {
		t.Fatalf("expected typed-nil imagemeta processor to be ignored")
	}
}
