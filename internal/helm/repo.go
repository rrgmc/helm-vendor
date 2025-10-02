package helm

import (
	"fmt"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
)

type Repository struct {
	repository *repo.ChartRepository
	index      *repo.IndexFile
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

func (r *Repository) ResolveReferenceURL(url string) (string, error) {
	return repo.ResolveReferenceURL(r.repository.Config.URL, url)
}

func (r *Repository) GetChart(name, version string) (*Chart, error) {
	chart, err := r.index.Get(name, version)
	if err != nil {
		return nil, fmt.Errorf("error getting chart from index: %w", err)
	}
	return LoadChart(r, chart)
}

func (r *Repository) Close() error {
	if r.repository.CachePath == "" {
		return nil
	}
	_ = os.RemoveAll(filepath.Join(r.repository.CachePath, helmpath.CacheChartsFile(r.repository.Config.Name)))
	_ = os.RemoveAll(filepath.Join(r.repository.CachePath, helmpath.CacheIndexFile(r.repository.Config.Name)))
	return nil
}
