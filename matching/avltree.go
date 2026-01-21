package matching

// AVLTree is a self-balancing binary search tree for price levels
type AVLTree struct {
	root       *LevelNode
	size       int
	descending bool // true for bids (highest first), false for asks (lowest first)
}

// NewAVLTree creates a new AVL tree
func NewAVLTree(descending bool) *AVLTree {
	return &AVLTree{
		root:       nil,
		size:       0,
		descending: descending,
	}
}

// Size returns the number of nodes in the tree
func (t *AVLTree) Size() int {
	return t.size
}

// Empty returns true if the tree is empty
func (t *AVLTree) Empty() bool {
	return t.size == 0
}

// First returns the first (best) price level
func (t *AVLTree) First() *LevelNode {
	if t.root == nil {
		return nil
	}
	node := t.root
	for node.Left != nil {
		node = node.Left
	}
	return node
}

// Last returns the last price level
func (t *AVLTree) Last() *LevelNode {
	if t.root == nil {
		return nil
	}
	node := t.root
	for node.Right != nil {
		node = node.Right
	}
	return node
}

// Find finds a level by price
func (t *AVLTree) Find(price uint64) *LevelNode {
	node := t.root
	for node != nil {
		if price == node.Price {
			return node
		}
		if t.compare(price, node.Price) < 0 {
			node = node.Left
		} else {
			node = node.Right
		}
	}
	return nil
}

// compare compares two prices for ordering
func (t *AVLTree) compare(a, b uint64) int {
	if t.descending {
		// Descending order for bids (higher prices first)
		if a > b {
			return -1
		}
		if a < b {
			return 1
		}
		return 0
	}
	// Ascending order for asks (lower prices first)
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// Insert inserts a new level into the tree
func (t *AVLTree) Insert(level *LevelNode) {
	if t.root == nil {
		t.root = level
		t.size++
		return
	}

	// Find the parent node
	parent := t.root
	var isLeft bool
	for {
		cmp := t.compare(level.Price, parent.Price)
		if cmp < 0 {
			if parent.Left == nil {
				parent.Left = level
				level.Parent = parent
				isLeft = true
				break
			}
			parent = parent.Left
		} else {
			if parent.Right == nil {
				parent.Right = level
				level.Parent = parent
				isLeft = false
				break
			}
			parent = parent.Right
		}
	}

	t.size++

	// Rebalance the tree
	t.rebalanceInsert(level, parent, isLeft)
}

// Remove removes a level from the tree
func (t *AVLTree) Remove(level *LevelNode) {
	if level == nil {
		return
	}

	var replacement *LevelNode
	var parent *LevelNode

	if level.Left == nil && level.Right == nil {
		// Leaf node
		replacement = nil
		parent = level.Parent
	} else if level.Left == nil {
		// Only right child
		replacement = level.Right
		parent = level.Parent
	} else if level.Right == nil {
		// Only left child
		replacement = level.Left
		parent = level.Parent
	} else {
		// Two children - find successor
		successor := level.Right
		for successor.Left != nil {
			successor = successor.Left
		}

		// Copy successor data
		level.Level = successor.Level
		level.OrderList = successor.OrderList

		// Update orders to point to new level
		for order := level.OrderList.Head; order != nil; order = order.Next {
			order.Level = level
		}

		// Remove successor instead
		if successor.Parent == level {
			level.Right = successor.Right
			if successor.Right != nil {
				successor.Right.Parent = level
			}
			parent = level
		} else {
			successor.Parent.Left = successor.Right
			if successor.Right != nil {
				successor.Right.Parent = successor.Parent
			}
			parent = successor.Parent
		}
		t.size--
		t.rebalanceRemove(parent)
		return
	}

	// Update parent's child pointer
	if parent == nil {
		t.root = replacement
	} else if parent.Left == level {
		parent.Left = replacement
	} else {
		parent.Right = replacement
	}

	if replacement != nil {
		replacement.Parent = parent
	}

	t.size--

	// Rebalance
	if parent != nil {
		t.rebalanceRemove(parent)
	}
}

// rebalanceInsert rebalances the tree after insertion
func (t *AVLTree) rebalanceInsert(node, parent *LevelNode, isLeft bool) {
	for parent != nil {
		if isLeft {
			parent.Balance--
		} else {
			parent.Balance++
		}

		if parent.Balance == 0 {
			break
		}

		if parent.Balance == -2 || parent.Balance == 2 {
			t.rebalance(parent)
			break
		}

		node = parent
		parent = node.Parent
		if parent != nil {
			isLeft = parent.Left == node
		}
	}
}

// rebalanceRemove rebalances the tree after removal
func (t *AVLTree) rebalanceRemove(node *LevelNode) {
	for node != nil {
		oldBalance := node.Balance

		// Recalculate balance
		leftHeight := t.height(node.Left)
		rightHeight := t.height(node.Right)
		node.Balance = rightHeight - leftHeight

		if node.Balance == -2 || node.Balance == 2 {
			node = t.rebalance(node)
			if node.Balance == -1 || node.Balance == 1 {
				break
			}
		} else if oldBalance == 0 {
			break
		}

		node = node.Parent
	}
}

// height returns the height of a subtree
func (t *AVLTree) height(node *LevelNode) int {
	if node == nil {
		return 0
	}
	leftHeight := t.height(node.Left)
	rightHeight := t.height(node.Right)
	if leftHeight > rightHeight {
		return leftHeight + 1
	}
	return rightHeight + 1
}

// rebalance performs AVL tree rotations
func (t *AVLTree) rebalance(node *LevelNode) *LevelNode {
	if node.Balance == -2 {
		if node.Left.Balance <= 0 {
			return t.rotateRight(node)
		}
		t.rotateLeft(node.Left)
		return t.rotateRight(node)
	}

	if node.Balance == 2 {
		if node.Right.Balance >= 0 {
			return t.rotateLeft(node)
		}
		t.rotateRight(node.Right)
		return t.rotateLeft(node)
	}

	return node
}

// rotateLeft performs a left rotation
func (t *AVLTree) rotateLeft(node *LevelNode) *LevelNode {
	pivot := node.Right
	parent := node.Parent

	node.Right = pivot.Left
	if node.Right != nil {
		node.Right.Parent = node
	}

	pivot.Left = node
	node.Parent = pivot

	pivot.Parent = parent
	if parent == nil {
		t.root = pivot
	} else if parent.Left == node {
		parent.Left = pivot
	} else {
		parent.Right = pivot
	}

	// Update balance factors
	node.Balance = node.Balance - 1 - max(0, pivot.Balance)
	pivot.Balance = pivot.Balance - 1 + min(0, node.Balance)

	return pivot
}

// rotateRight performs a right rotation
func (t *AVLTree) rotateRight(node *LevelNode) *LevelNode {
	pivot := node.Left
	parent := node.Parent

	node.Left = pivot.Right
	if node.Left != nil {
		node.Left.Parent = node
	}

	pivot.Right = node
	node.Parent = pivot

	pivot.Parent = parent
	if parent == nil {
		t.root = pivot
	} else if parent.Left == node {
		parent.Left = pivot
	} else {
		parent.Right = pivot
	}

	// Update balance factors
	node.Balance = node.Balance + 1 - min(0, pivot.Balance)
	pivot.Balance = pivot.Balance + 1 + max(0, node.Balance)

	return pivot
}

// ForEach iterates over all levels in order
func (t *AVLTree) ForEach(fn func(*LevelNode) bool) {
	t.forEach(t.root, fn)
}

func (t *AVLTree) forEach(node *LevelNode, fn func(*LevelNode) bool) bool {
	if node == nil {
		return true
	}
	if !t.forEach(node.Left, fn) {
		return false
	}
	if !fn(node) {
		return false
	}
	return t.forEach(node.Right, fn)
}
