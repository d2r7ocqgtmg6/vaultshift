package vault

import (
	"fmt"
	"strings"
)

// GraphNode represents a secret path node in the dependency graph.
type GraphNode struct {
	Path     string
	Children []string
	Depth    int
}

// GraphResult holds the full dependency graph output.
type GraphResult struct {
	Nodes []GraphNode
	Edges [][2]string
}

// Grapher builds a dependency graph of secret paths under a prefix.
type Grapher struct {
	client *Client
	logger AuditLogger
}

// NewGrapher constructs a Grapher or returns an error if dependencies are missing.
func NewGrapher(client *Client, logger AuditLogger) (*Grapher, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Grapher{client: client, logger: logger}, nil
}

// Build traverses the secret tree under prefix and returns a GraphResult.
func (g *Grapher) Build(prefix string) (*GraphResult, error) {
	result := &GraphResult{}
	err := g.walk(prefix, 0, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *Grapher) walk(prefix string, depth int, result *GraphResult) error {
	keys, err := g.client.List(prefix)
	if err != nil {
		return fmt.Errorf("list %q: %w", prefix, err)
	}

	var children []string
	for _, key := range keys {
		full := strings.TrimRight(prefix, "/") + "/" + key
		children = append(children, full)
		if strings.HasSuffix(key, "/") {
			result.Edges = append(result.Edges, [2]string{prefix, full})
			if err := g.walk(full, depth+1, result); err != nil {
				return err
			}
		} else {
			result.Edges = append(result.Edges, [2]string{prefix, full})
		}
	}

	result.Nodes = append(result.Nodes, GraphNode{
		Path:     prefix,
		Children: children,
		Depth:    depth,
	})

	g.logger.Log("graph_walk", map[string]interface{}{
		"prefix": prefix,
		"depth":  depth,
		"count":  len(keys),
	})

	return nil
}
