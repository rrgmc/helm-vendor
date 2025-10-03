package helm

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
)

type Repository struct {
	repository *repo.ChartRepository
	index      *repo.IndexFile
	registry   *registry.Client
}

func LoadRepository(repoURL string) (*Repository, error) {
	c := repo.Entry{
		URL:                   repoURL,
		Username:              "",
		Password:              "",
		PassCredentialsAll:    false,
		CertFile:              "",
		KeyFile:               "",
		CAFile:                "",
		Name:                  randomName(),
		InsecureSkipTLSverify: false,
	}
	if registry.IsOCI(repoURL) {
		return loadRepositoryOCI(&c)
	}
	repository, err := repo.NewChartRepository(&c, allGetters)
	if err != nil {
		return nil, fmt.Errorf("error loading repository %s: %w", repoURL, err)
	}
	return loadRepository(repository)
}

func loadRepository(repository *repo.ChartRepository) (*Repository, error) {
	indexFilename, err := repository.DownloadIndexFile()
	if err != nil {
		return nil, fmt.Errorf("error downloading repository index file: %w", err)
	}

	// Read the index file for the repository to get chart information and return chart URL
	repoIndex, err := repo.LoadIndexFile(indexFilename)
	if err != nil {
		return nil, fmt.Errorf("error loading repository index file: %w", err)
	}

	return &Repository{
		repository: repository,
		index:      repoIndex,
	}, nil
}

func loadRepositoryOCI(entry *repo.Entry) (*Repository, error) {
	// ref := strings.TrimPrefix(entry.URL, fmt.Sprintf("%s://", registry.OCIScheme))

	registryClient, err := registry.NewClient(
		// registry.ClientOptDebug(true),
		registry.ClientOptEnableCache(false),
		// registry.ClientOptWriter(os.Stdout),
		// registry.ClientOptCredentialsFile(credentialsFile),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating registry client: %w", err)
	}

	// err = registryClient.Login(
	// 	ref,
	// 	// registry.LoginOptBasicAuth(srv.TestUsername, srv.TestPassword),
	// 	// registry.LoginOptInsecure(true),
	// 	// registry.LoginOptPlainText(true),
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("error logging into registry with good credentials: %w", err)
	// }

	return &Repository{
		repository: &repo.ChartRepository{
			Config: entry,
		},
		index:    nil,
		registry: registryClient,
	}, nil
}

func (r *Repository) ResolveReferenceURL(url string) (string, error) {
	return repo.ResolveReferenceURL(r.repository.Config.URL, url)
}

func (r *Repository) GetChart(name, version string) (*Chart, error) {
	if r.index == nil {
		findChart, err := r.FindChartVersion(name, version)
		if err != nil || findChart == nil {
			// errors may be because it is not possible to list tags
			return LoadChart(r, &repo.ChartVersion{
				Metadata: &chart.Metadata{
					Name:    name,
					Version: version,
				},
				URLs: []string{JoinHTTPPaths(r.repository.Config.URL, name)},
			})
		}
		return LoadChart(r, findChart)
		// return nil, errors.New("cannot get chart from OCI registry")
	}

	c, err := r.index.Get(name, version)
	if err != nil {
		return nil, fmt.Errorf("error getting chart from index: %w", err)
	}
	return LoadChart(r, c)
}

func (r *Repository) FindChartVersion(name string, version string) (*repo.ChartVersion, error) {
	for cv, err := range r.ChartVersions(name, 0) {
		if err != nil {
			return nil, err
		}
		if version == "" {
			// return the first one
			return cv, nil
		}
		if version == cv.Version {
			return cv, nil
		}
	}
	return nil, nil
}

func (r *Repository) ChartVersions(name string, maxAmount int) iter.Seq2[*repo.ChartVersion, error] {
	if r.index == nil {
		return r.chartVersionsOCI(name, maxAmount)
	}
	return func(yield func(*repo.ChartVersion, error) bool) {
		chartEntries, ok := r.index.Entries[name]
		if !ok {
			yield(nil, fmt.Errorf("unknown chart %s", name))
			return
		}
		var ct int
		for _, entry := range chartEntries {
			if !yield(entry, nil) {
				return
			}
			ct++
			if maxAmount > 0 && ct >= maxAmount {
				return
			}
		}
	}
}

func (r *Repository) chartVersionsOCI(name string, maxAmount int) iter.Seq2[*repo.ChartVersion, error] {
	return func(yield func(*repo.ChartVersion, error) bool) {
		// if r.index == nil {
		// 	yield(nil, errors.New("cannot list chart from OCI registry"))
		// }

		ref := strings.TrimPrefix(JoinHTTPPaths(r.repository.Config.URL, name), fmt.Sprintf("%s://", registry.OCIScheme))
		tags, err := r.registry.Tags(ref)
		if err != nil {
			yield(nil, fmt.Errorf("error getting tags for chart %s: %w", name, err))
			return
		}
		var ct int
		for _, entry := range tags {
			if !yield(&repo.ChartVersion{
				Metadata: &chart.Metadata{
					Name:    name,
					Version: entry,
				},
				URLs: []string{JoinHTTPPaths(r.repository.Config.URL, name) + ":" + entry},
			}, nil) {
				return
			}
			ct++
			if maxAmount > 0 && ct >= maxAmount {
				return
			}
		}
	}
}

func (r *Repository) Close() error {
	if r.repository.CachePath == "" {
		return nil
	}
	_ = os.RemoveAll(filepath.Join(r.repository.CachePath, helmpath.CacheChartsFile(r.repository.Config.Name)))
	_ = os.RemoveAll(filepath.Join(r.repository.CachePath, helmpath.CacheIndexFile(r.repository.Config.Name)))
	return nil
}
