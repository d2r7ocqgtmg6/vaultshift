package vault

import (
	"strings"
	"testing"
)

func TestGraph_NodeDepth_RootIsZero(t *testing.T) {
	srv := newGraphMockServer(t)
	defer srv.Close()
	c := newGraphClient(t, srv)
	g, err := NewGrapher(c, newTraceLogger(t))
	if err != nil {
		t.Fatalf("NewGrapher: %v", err)
	}
	result, err := g.Build("secret/metadata/root")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, node := range result.Nodes {
		if node.Path == "secret/metadata/root" && node.Depth != 0 {
			t.Errorf("root node depth = %d, want 0", node.Depth)
		}
	}
}

func TestGraph_ChildNode_HasGreaterDepth(t *testing.T) {
	srv := newGraphMockServer(t)
	defer srv.Close()
	c := newGraphClient(t, srv)
	g, _ := NewGrapher(c, newTraceLogger(t))
	result, err := g.Build("secret/metadata/root")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, node := range result.Nodes {
		if strings.Contains(node.Path, "child") && node.Depth <= 0 {
			t.Errorf("child node %q has depth %d, want > 0", node.Path, node.Depth)
		}
	}
}

func TestGraph_AllEdgesHaveTwoEndpoints(t *testing.T) {
	srv := newGraphMockServer(t)
	defer srv.Close()
	c := newGraphClient(t, srv)
	g, _ := NewGrapher(c, newTraceLogger(t))
	result, err := g.Build("secret/metadata/root")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for i, edge := range result.Edges {
		if edge[0] == "" || edge[1] == "" {
			t.Errorf("edge[%d] has empty endpoint: %v", i, edge)
		}
	}
}
