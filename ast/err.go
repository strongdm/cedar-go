package ast

// Must returns the node, panicking if an error was recorded while building it.
func (n Node) Must() Node {
	return wrapNode(n.Node.Must())
}

// Err returns the first error recorded while building the policy, such as an invalid entity type name or annotation
// key. A policy that records an error must not be used.
func (p *Policy) Err() error {
	return p.unwrap().Err()
}

// Must returns the policy, panicking if an error was recorded while building it.
func (p *Policy) Must() *Policy {
	return wrapPolicy(p.unwrap().Must())
}
