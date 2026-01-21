package matching

import "sync"

// Object pools for high-performance order processing.
// These pools reduce GC pressure by reusing allocated objects.

// orderNodePool is a pool of OrderNode objects for reuse
var orderNodePool = sync.Pool{
	New: func() interface{} {
		return &OrderNode{}
	},
}

// levelNodePool is a pool of LevelNode objects for reuse
var levelNodePool = sync.Pool{
	New: func() interface{} {
		return &LevelNode{}
	},
}

// AcquireOrderNode gets an OrderNode from the pool
func AcquireOrderNode() *OrderNode {
	return orderNodePool.Get().(*OrderNode)
}

// ReleaseOrderNode returns an OrderNode to the pool
func ReleaseOrderNode(node *OrderNode) {
	if node == nil {
		return
	}
	// Clear references to allow GC of linked objects
	node.Next = nil
	node.Prev = nil
	node.Level = nil
	orderNodePool.Put(node)
}

// AcquireLevelNode gets a LevelNode from the pool
func AcquireLevelNode() *LevelNode {
	return levelNodePool.Get().(*LevelNode)
}

// ReleaseLevelNode returns a LevelNode to the pool
func ReleaseLevelNode(node *LevelNode) {
	if node == nil {
		return
	}
	// Clear references
	node.Parent = nil
	node.Left = nil
	node.Right = nil
	node.OrderList = OrderList{}
	levelNodePool.Put(node)
}

// NewOrderNodePooled creates a new OrderNode from pool and initializes it
func NewOrderNodePooled(order Order) *OrderNode {
	node := AcquireOrderNode()
	node.Order = order
	node.Next = nil
	node.Prev = nil
	node.Level = nil
	return node
}

// NewLevelNodePooled creates a new LevelNode from pool and initializes it
func NewLevelNodePooled(levelType LevelType, price uint64) *LevelNode {
	node := AcquireLevelNode()
	node.Level = NewLevel(levelType, price)
	node.OrderList = OrderList{}
	node.Parent = nil
	node.Left = nil
	node.Right = nil
	node.Balance = 0
	return node
}
