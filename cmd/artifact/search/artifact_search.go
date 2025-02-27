// Copyright 2022 The Falco Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package search

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/falcosecurity/falcoctl/pkg/oci"
	"github.com/falcosecurity/falcoctl/pkg/options"
	"github.com/falcosecurity/falcoctl/pkg/output"
)

const (
	defaultMinScore = 0.65
	// CommandName name of the command. It has to be the first word in the use line.
	CommandName = "search"
)

type artifactSearchOptions struct {
	*options.Common
	minScore     float64
	artifactType oci.ArtifactType
}

func (o *artifactSearchOptions) Validate() error {
	if o.minScore <= 0 || o.minScore > 1 {
		return fmt.Errorf("minScore must be a number within (0,1]")
	}

	return nil
}

// NewArtifactSearchCmd returns the artifact search command.
func NewArtifactSearchCmd(ctx context.Context, opt *options.Common) *cobra.Command {
	o := artifactSearchOptions{
		Common: opt,
	}

	cmd := &cobra.Command{
		Use:                   fmt.Sprintf("%s [keyword1 [keyword2 ...]] [flags]", CommandName),
		DisableFlagsInUseLine: true,
		Short:                 "Search an artifact by keywords",
		Long:                  "Search an artifact by keywords",
		Args:                  cobra.MinimumNArgs(1),
		SilenceErrors:         true,
		SilenceUsage:          true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return o.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunArtifactSearch(ctx, args)
		},
	}

	cmd.Flags().Float64VarP(&o.minScore, "min-score", "", defaultMinScore,
		"the minimum score used to match artifact names with search keywords")

	cmd.Flags().Var(&o.artifactType, "type", `Only search artifacts with a specific type. Allowed values: "rulesfile", "plugin""`)

	return cmd
}

func (o *artifactSearchOptions) RunArtifactSearch(_ context.Context, args []string) error {
	resultEntries := o.IndexCache.MergedIndexes.SearchByKeywords(o.minScore, args...)

	var data [][]string
	for _, entry := range resultEntries {
		if o.artifactType != "" && o.artifactType != oci.ArtifactType(entry.Type) {
			continue
		}
		indexName := o.IndexCache.MergedIndexes.IndexByEntry(entry).Name
		row := []string{indexName, entry.Name, entry.Type, entry.Registry, entry.Repository}
		data = append(data, row)
	}

	return o.Printer.PrintTable(output.ArtifactSearch, data)
}
