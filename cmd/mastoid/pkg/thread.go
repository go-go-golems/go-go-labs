package pkg

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/mattn/go-mastodon"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Node struct {
	ID          mastodon.ID
	Ancestors   *orderedmap.OrderedMap[mastodon.ID, *Node]
	Descendants *orderedmap.OrderedMap[mastodon.ID, *Node]
	Status      *mastodon.Status
}

type Thread struct {
	Nodes map[mastodon.ID]*Node
}

// GetNode returns the tree node for the given ID, creating it if not already present.
func (t *Thread) GetNode(id mastodon.ID) *Node {
	if node, ok := t.Nodes[id]; ok {
		return node
	}
	node := &Node{
		ID:          id,
		Ancestors:   orderedmap.New[mastodon.ID, *Node](),
		Descendants: orderedmap.New[mastodon.ID, *Node](),
	}

	t.Nodes[id] = node
	return node
}

// GetRoots returns the roots (nodes without ancestors) of the thread, in order.
func (t *Thread) GetRoots() []*Node {
	var ret []*Node

	for k, v := range t.Nodes {
		if v.Ancestors.Len() == 0 {
			ret = append(ret, t.Nodes[k])
		}
	}

	return ret
}

func (t *Thread) AddStatus(status *mastodon.Status) {
	node := t.GetNode(status.ID)
	node.Status = status
}

// AddContextAndGetMissingIDs adds the given context of the given status to the thread.
// Returns the list of missing IDs to be added.
func (t *Thread) AddContextAndGetMissingIDs(id mastodon.ID, context *mastodon.Context) {
	node := t.GetNode(id)

	curNode := node
	// ancestors are a straight lineage
	for _, ancestor := range context.Ancestors {
		ancestorNode := t.GetNode(ancestor.ID)
		ancestorNode.Status = ancestor
		curNode.Ancestors.Set(ancestor.ID, ancestorNode)
		ancestorNode.Descendants.Set(curNode.ID, curNode)
		curNode = ancestorNode
	}

	for _, descendant := range context.Descendants {
		descendantNode := t.GetNode(descendant.ID)
		descendantNode.Status = descendant

		inReplyToID := mastodon.ID(descendant.InReplyToID.(string))
		parentNode := t.GetNode(inReplyToID)
		parentNode.Descendants.Set(descendant.ID, descendantNode)
		descendantNode.Ancestors.Set(inReplyToID, t.GetNode(inReplyToID))
	}
}

type queueEntry struct {
	node  *Node
	depth int
}

// WalkBreadthFirst walks the thread in order, calling the given function on each node.
func (t *Thread) WalkBreadthFirst(f func(n *Node, depth int) error) error {
	roots := t.GetRoots()

	for _, root := range roots {
		queue := []*queueEntry{
			{
				node:  root,
				depth: 0,
			},
		}
		for len(queue) > 0 {
			entry := queue[0]
			queue = queue[1:]
			err := f(entry.node, entry.depth)
			if err != nil {
				return err
			}
			for pair := entry.node.Descendants.Oldest(); pair != nil; pair = pair.Next() {
				queue = append(queue, &queueEntry{
					node:  pair.Value,
					depth: entry.depth + 1,
				})
			}
		}
	}
	return nil
}

func (t *Thread) WalkDepthFirst(f func(n *Node, depth int) error) error {
	roots := t.GetRoots()

	for _, root := range roots {
		queue := []*queueEntry{
			{
				node:  root,
				depth: 0,
			},
		}

		for len(queue) > 0 {
			entry := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			err := f(entry.node, entry.depth)
			if err != nil {
				return err
			}
			for pair := entry.node.Descendants.Oldest(); pair != nil; pair = pair.Next() {
				queue = append(queue, &queueEntry{
					node:  pair.Value,
					depth: entry.depth + 1,
				})
			}
		}
	}

	return nil
}

func (t *Thread) walkDepthFirst(node *Node, depth int, f func(n *Node, depth int) error) error {
	err := f(node, depth)
	if err != nil {
		return err
	}
	for pair := node.Descendants.Oldest(); pair != nil; pair = pair.Next() {
		err := t.walkDepthFirst(pair.Value, depth+1, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Thread) OutputToProcessor(ctx context.Context, gp middlewares.Processor) error {
	for _, root := range t.GetRoots() {
		err := gp.AddRow(ctx, types.NewRowFromStruct(root.Status, true))
		if err != nil {
			return err
		}
	}

	return nil
}
