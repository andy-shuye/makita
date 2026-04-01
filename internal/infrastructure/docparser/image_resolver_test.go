package docparser

import (
	"context"
	"io"
	"mime/multipart"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

type mockFileService struct{}

func (m *mockFileService) CheckConnectivity(ctx context.Context) error { return nil }
func (m *mockFileService) SaveFile(ctx context.Context, file *multipart.FileHeader, tenantID uint64, knowledgeID string) (string, error) {
	return "", nil
}
func (m *mockFileService) SaveBytes(ctx context.Context, data []byte, tenantID uint64, fileName string, temp bool) (string, error) {
	return "minio://bucket/1/exports/uploaded.jpg", nil
}
func (m *mockFileService) GetFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	return nil, nil
}
func (m *mockFileService) GetFileURL(ctx context.Context, filePath string) (string, error) { return "", nil }
func (m *mockFileService) DeleteFile(ctx context.Context, filePath string) error            { return nil }

func TestResolveAndStore_NormalizeRefAndReplace(t *testing.T) {
	resolver := NewImageResolver()
	result := &types.ReadResult{
		MarkdownContent: `段落 ![](images/a.jpg)`,
		ImageRefs: []types.ImageRef{
			{
				OriginalRef: "./images/a.jpg",
				MimeType:    "image/jpeg",
				ImageData:   []byte{1, 2, 3},
			},
		},
	}

	updated, images, err := resolver.ResolveAndStore(context.Background(), result, &mockFileService{}, 1)
	if err != nil {
		t.Fatalf("ResolveAndStore returned error: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 stored image, got %d", len(images))
	}
	if updated == result.MarkdownContent || updated == "" {
		t.Fatalf("expected markdown to be updated, got: %q", updated)
	}
}

func TestResolveAndStore_WithMarkdownTitle(t *testing.T) {
	resolver := NewImageResolver()
	result := &types.ReadResult{
		MarkdownContent: `![](images/a.jpg "title")`,
		ImageRefs: []types.ImageRef{
			{
				OriginalRef: "images/a.jpg",
				MimeType:    "image/jpeg",
				ImageData:   []byte{1, 2, 3},
			},
		},
	}

	updated, images, err := resolver.ResolveAndStore(context.Background(), result, &mockFileService{}, 1)
	if err != nil {
		t.Fatalf("ResolveAndStore returned error: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 stored image, got %d", len(images))
	}
	if updated == result.MarkdownContent || updated == "" {
		t.Fatalf("expected markdown to be updated, got: %q", updated)
	}
}

