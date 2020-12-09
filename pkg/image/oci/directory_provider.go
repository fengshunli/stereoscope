package oci

import (
	"fmt"

	"github.com/anchore/stereoscope/pkg/image"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

// DirectoryImageProvider is an image.Provider for an OCI image (V1) for an existing tar on disk (from a buildah push <img> oci:<img> command).
type DirectoryImageProvider struct {
	path string
}

// NewProviderFromPath creates a new provider instance for the specific image already at the given path.
func NewProviderFromPath(path string) *DirectoryImageProvider {
	return &DirectoryImageProvider{
		path: path,
	}
}

// Provide an image object that represents the OCI image as a directory.
func (p *DirectoryImageProvider) Provide() (*image.Image, error) {
	pathObj, err := layout.FromPath(p.path)
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory path=%q : %w", p.path, err)
	}

	index, err := layout.ImageIndexFromPath(p.path)
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory index: %w", err)
	}

	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory indexManifest: %w", err)
	}

	// for now, lets only support one image indexManifest (it is not clear how to handle multiple manifests)
	if len(indexManifest.Manifests) != 1 {
		return nil, fmt.Errorf("unexpected number of OCI directory manifests (found %d)", len(indexManifest.Manifests))
	}

	manifest := indexManifest.Manifests[0]
	img, err := pathObj.Image(manifest.Digest)
	if err != nil {
		return nil, fmt.Errorf("unable to parse OCI directory as an image: %w", err)
	}

	var metadata = []image.AdditionalMetadata{
		image.WithManifestDigest(manifest.Digest.String()),
	}

	// make a best-effort attempt at getting the raw indexManifest
	rawManifest, err := img.RawManifest()
	if err == nil {
		metadata = append(metadata, image.WithManifest(rawManifest))
	}

	return image.NewImage(img, metadata...), nil
}
