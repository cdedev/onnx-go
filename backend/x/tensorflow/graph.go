package tfrt

import (
	"github.com/owulveryck/onnx-go"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"

	"gorgonia.org/tensor"
)

// Graph is the top structure that should be compatible with
//    backend.ComputationGraph
// It holds a gorgonia.ExprGraph that is populated on the first call to the
// Run() method
type Graph struct {
	g       *simple.WeightedDirectedGraph
	tfGraph *tf.Graph
	m       *tf.Session
	roots   []int64
	groups  [][]*Node // a reference of all the nodes that belongs to a group
}

// GetGraph returns the tensorflow graph; if the graph is nil, it populates the graph before returing it
func (g *Graph) GetGraph() (*tf.Graph, error) {
	var err error
	if g.tfGraph == nil {
		err = g.PopulateExprgraph()
	}
	return g.tfGraph, err
}

// ApplyOperation to fulfill the onnx.Backend interface
func (g *Graph) ApplyOperation(o onnx.Operation, ns ...graph.Node) error {
	nodes := make([]*Node, len(ns))
	for i, n := range ns {
		n.(*Node).operation = &o
		nodes[i] = n.(*Node)
	}
	g.groups = append(g.groups, nodes)
	return nil
}

// Run the graph. It populate the underlying exprgraph if the graph is nil
func (g *Graph) Run() error {
	if g.tfGraph == nil {
		err := g.PopulateExprgraph()
		if err != nil {
			return err
		}
	}
	var err error
	if g.m == nil {
		g.m, err = tf.NewSession(g.tfGraph, nil)
		if err != nil {
			return err
		}
	} else {
		g.m.Close()
		g.m, err = tf.NewSession(g.tfGraph, nil)
		if err != nil {
			return err
		}
	}

	// TODO
	output, err := g.m.Run(nil, nil, nil)
	if err != nil {
		return err
	}
	// Now sets the output tensor
	for i := 0; i < len(g.roots); i++ {
		root := g.Node(g.roots[i]).(*Node)
		root.t = tensor.New(
			tensor.WithBacking(output[i].Value()),
			tensor.WithShape(
				toIntSlice(output[i].Shape())...,
			))
	}
	return nil
}

// PopulateExprgraph creates the underlynig graph by walking the current graph
func (g *Graph) PopulateExprgraph() error {
	g.tfGraph = tf.NewGraph()
	// Find the root nodes
	// TODO make it more efficient
	g.roots = make([]int64, 0)
	it := g.g.Nodes()
	for it.Next() {
		n := it.Node()
		if g.g.To(n.ID()).Len() == 0 {
			g.roots = append(g.roots, n.ID())
		}
	}
	return g.populateExprgraph()
}