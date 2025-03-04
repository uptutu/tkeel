/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// _verSep is a separator for version fields in map keys.
const _verSep = "$$"

type PluginRes struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Repo        string   `json:"repository"` // nolint
	URLs        []string `json:"urls"`       //nolint
	Description string   `json:"description"`
}

func (r PluginRes) ToInstallerBrief() *repository.InstallerBrief {
	return &repository.InstallerBrief{
		Name:      r.Name,
		Repo:      r.Repo,
		Version:   r.Version,
		Installed: false,
	}
}

type PluginResList []*PluginRes

func (r *PluginResList) ToInstallerBrief() []*repository.InstallerBrief {
	list := make([]*repository.InstallerBrief, 0, len(*r))
	for _, res := range *r {
		list = append(list, res.ToInstallerBrief())
	}
	return list
}

type Index struct {
	RepoName  string
	helmIndex *repo.IndexFile
	charts    map[string]*repo.ChartVersion
}

// NewIndex creates a new Index.
func NewIndex(repoName string, data []byte) (*Index, error) {
	i := &repo.IndexFile{}

	if len(data) == 0 {
		return nil, repo.ErrEmptyIndexYaml
	}

	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return nil, errors.Wrap(err, "unmarshal data to IndexFile failed")
	}
	for name, cvs := range i.Entries {
		for idx := len(cvs) - 1; idx >= 0; idx-- {
			if cvs[idx].APIVersion == "" {
				cvs[idx].APIVersion = chart.APIVersionV1
			}
			if err := cvs[idx].Validate(); err != nil {
				log.Infof("skipping loading invalid entry for chart %q %q : %s", name, cvs[idx].Version, err)
				cvs = append(cvs[:idx], cvs[idx+1:]...)
			}
		}
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	index := &Index{
		RepoName:  repoName,
		helmIndex: i,
		charts:    make(map[string]*repo.ChartVersion),
	}

	for name, ref := range i.Entries {
		if len(ref) == 0 {
			continue
		}

		for _, rr := range ref {
			versionedName := name + _verSep + rr.Version
			index.charts[versionedName] = rr
		}
	}

	return index, nil
}

func (r *Index) Search(word string, version string) PluginResList {
	list := make(PluginResList, 0, len(r.helmIndex.Entries))
	if word == "*" {
		for _, ch := range r.charts {
			res := PluginRes{
				Name:        ch.Name,
				Version:     ch.Version,
				Repo:        r.RepoName,
				URLs:        ch.URLs,
				Description: ch.Description,
			}
			list = append(list, &res)
		}
		return list
	}

	for _, ch := range r.charts {
		if ch.Name == word {
			if version == "" || version == ch.Version {
				res := PluginRes{
					Name:        ch.Name,
					Version:     ch.Version,
					Repo:        r.RepoName,
					URLs:        ch.URLs,
					Description: ch.Description,
				}
				list = append(list, &res)
			}
		}
	}

	return list
}
