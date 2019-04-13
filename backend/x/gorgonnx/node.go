package gorgonnx

import (
	"github.com/owulveryck/onnx-go"
	"gorgonia.org/tensor"
)

// Node is compatible with graph.Node and
// onnx.DataCarrier
type Node struct {
	id        int64
	t         tensor.Tensor
	operation onnx.Operation
	name      string
}

// ID to fulfill the graph.Node interface
func (n *Node) ID() int64 {
	return n.id
}

// SetTensor assign the tensor N to the underlying node
func (n *Node) SetTensor(t tensor.Tensor) error {
	n.t = t
	return nil
}

// GetTensor value from the node
func (n *Node) GetTensor() tensor.Tensor {
	return n.t
}
