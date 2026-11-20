package setup

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/stores/conn"
)

// composePrefix is the compose project name prefix in datastore/compose.yaml (name: polystore-${DATASET}).
const composePrefix = "polystore-"

// datasetPrepare reshapes the graph per dataset (before the uuid indexes are made); keys are dataset folder
// names in datastore/datasets. Datasets without an entry get only the common steps (uuid indexes).
var datasetPrepare = map[string]func(ctx context.Context, session neo4j.SessionWithContext) error{
	"ldbc_snb_sf1": func(ctx context.Context, session neo4j.SessionWithContext) error {
		return applyTypeLabels(ctx, session, ldbcTypeLabels)
	},
}

// RunningDataset returns the running dataset: the running neo4j container's compose project
// (polystore-<name>) mapped to its dataset folder by l. Only one dataset runs at a time (the ports are shared),
// so none or several is an error.
func RunningDataset(ctx context.Context, l conn.Layout) (string, error) {
	out, err := exec.CommandContext(ctx, "docker", "ps",
		"--filter", "label=com.docker.compose.service=neo4j",
		"--format", `{{.Label "com.docker.compose.project"}}`).Output()
	if err != nil {
		return "", fmt.Errorf("cannot get the running dataset from docker: %w", err)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if ds, ok := strings.CutPrefix(strings.TrimSpace(line), composePrefix); ok && ds != "" {
			seen[ds] = true
		}
	}
	var names []string
	for ds := range seen {
		names = append(names, ds)
	}
	sort.Strings(names)
	switch len(names) {
	case 0:
		return "", fmt.Errorf("no running neo4j container (compose project %s<dataset>)", composePrefix)
	case 1:
		return l.DatasetOfProject(names[0])
	}
	return "", fmt.Errorf("neo4j of several datasets is running: %v", names)
}
