package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docker/app/internal/log"
	"github.com/docker/cli/cli/config/configfile"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/remotes"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

//
type BundleStore interface {
	Store(ref reference.Named, bndle *bundle.Bundle) error
	Read(ref reference.Named) (*bundle.Bundle, error)
	List() ([]reference.Named, error)

	LookupOrPullBundle(ref reference.Named, pullRef bool, config *configfile.ConfigFile, insecureRegistries []string) (*bundle.Bundle, error)
}

var _ BundleStore = &bundleStore{}

type bundleStore struct {
	path string
}

func (b *bundleStore) Store(ref reference.Named, bndle *bundle.Bundle) error {
	path, err := b.storePath(ref)
	if err != nil {
		return errors.Wrapf(err, "failed to store bundle %q", ref)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Wrapf(err, "failed to store bundle %q", ref)
	}
	err = bndle.WriteFile(path, 0644)
	return errors.Wrapf(err, "failed to store bundle %q", ref)
}

func (b *bundleStore) Read(ref reference.Named) (*bundle.Bundle, error) {
	path, err := b.storePath(ref)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read bundle %q", ref)
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read bundle %q", ref)
	}
	var bndle bundle.Bundle
	if err := json.Unmarshal(data, &bndle); err != nil {
		return nil, errors.Wrapf(err, "failed to read bundle %q", ref)
	}
	return &bndle, nil
}

// Returns the list of all bundles present in the bundle store
func (b *bundleStore) List() ([]reference.Named, error) {
	var references []reference.Named
	if err := filepath.Walk(b.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		ref, err := b.pathToReference(path)
		if err != nil {
			return err
		}

		references = append(references, ref)

		return nil
	}); err != nil {
		return nil, err
	}

	sort.Slice(references, func(i, j int) bool {
		return references[i].Name() < references[j].Name()
	})

	return references, nil
}

// LookupOrPullBundle will fetch the given bundle from the local
// bundle store, or if it is missing from the registry, and returns
// it. Always pulls if pullRef is true. If it pulls then the local
// bundle store is updated.
func (b *bundleStore) LookupOrPullBundle(ref reference.Named, pullRef bool, config *configfile.ConfigFile, insecureRegistries []string) (*bundle.Bundle, error) {
	if !pullRef {
		bndl, err := b.Read(ref)
		if err == nil {
			return bndl, nil
		}
		if !os.IsNotExist(errors.Cause(err)) {
			return nil, err
		}
	}
	bndl, err := remotes.Pull(log.WithLogContext(context.Background()), reference.TagNameOnly(ref), remotes.CreateResolver(config, insecureRegistries...))
	if err != nil {
		return nil, errors.Wrap(err, ref.String())
	}
	if err := b.Store(ref, bndl); err != nil {
		return nil, err
	}
	return bndl, nil
}

func (b *bundleStore) storePath(ref reference.Named) (string, error) {
	name := ref.Name()
	// A name is safe for use as a filesystem path (it is
	// alphanumerics + "." + "/") except for the ":" used to
	// separate domain from port which is not safe on Windows.
	// Replace it with "_" which is not valid in the name.
	//
	// There can be at most 1 ":" in a valid reference so only
	// replace one -- if there are more (and this wasn't caught
	// when parsing the ref) then there will be errors when we try
	// to use this as a path later.
	name = strings.Replace(name, ":", "_", 1)
	storeDir := filepath.Join(b.path, filepath.FromSlash(name))

	// We rely here on _ not being valid in a name meaning there can be no clashes due to nesting of repositories.
	switch t := ref.(type) {
	case reference.Digested:
		digest := t.Digest()
		storeDir = filepath.Join(storeDir, "_digests", digest.Algorithm().String(), digest.Encoded())
	case reference.Tagged:
		storeDir = filepath.Join(storeDir, "_tags", t.Tag())
	default:
		return "", errors.Errorf("%s: not tagged or digested", ref.String())
	}

	return storeDir + ".json", nil
}

func (b *bundleStore) pathToReference(path string) (reference.Named, error) {
	// Clean the path and remove the local bundle store path
	cleanpath := filepath.ToSlash(path)
	cleanpath = strings.TrimPrefix(cleanpath, filepath.ToSlash(b.path)+"/")

	// get the hierarchy of directories, so we can get digest algorithm or tag
	paths := strings.Split(cleanpath, "/")
	if len(paths) < 3 {
		return nil, fmt.Errorf("invalid path %q in the bundle store", path)
	}

	// path must point to a json file
	if !strings.Contains(paths[len(paths)-1], ".json") {
		return nil, fmt.Errorf("invalid path %q, not referencing a CNAB bundle in json format", path)
	}

	// remove the json suffix from the filename
	paths[len(paths)-1] = strings.TrimSuffix(paths[len(paths)-1], ".json")

	name, err := reconstructNamedReference(path, paths)
	if err != nil {
		return nil, err
	}

	return reference.ParseNamed(name)
}

func reconstructNamedReference(path string, paths []string) (string, error) {
	name, paths := strings.Replace(paths[0], "_", ":", 1), paths[1:]
	for i, p := range paths {
		switch p {
		case "_tags":
			if i != len(paths)-2 {
				return "", fmt.Errorf("invalid path %q in the bundle store", path)
			}
			return fmt.Sprintf("%s:%s", name, paths[i+1]), nil
		case "_digests":
			if i != len(paths)-3 {
				return "", fmt.Errorf("invalid path %q in the bundle store", path)
			}
			return fmt.Sprintf("%s@%s:%s", name, paths[i+1], paths[i+2]), nil
		default:
			name += "/" + p
		}
	}
	return name, nil
}
