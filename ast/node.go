package ast

import "github.com/cedar-policy/cedar-go/internal/ast"

type Node struct {
	ast.Node
}

func wrapNode(n ast.Node) Node {
	return Node{n}
}