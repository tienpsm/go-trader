package matching

import (
	"testing"
)

func TestOrderNodePool(t *testing.T) {
	// Get from pool
	node := AcquireOrderNode()
	if node == nil {
		t.Error("Expected non-nil node from pool")
	}

	// Initialize
	node.Order = Order{ID: 1, SymbolID: 1}
	node.Next = nil
	node.Prev = nil

	// Return to pool
	ReleaseOrderNode(node)

	// Get again - should be recycled
	node2 := AcquireOrderNode()
	if node2 == nil {
		t.Error("Expected non-nil node from pool")
	}
}

func TestLevelNodePool(t *testing.T) {
	// Get from pool
	node := AcquireLevelNode()
	if node == nil {
		t.Error("Expected non-nil node from pool")
	}

	// Initialize
	node.Level = NewLevel(LevelTypeBid, 10000)

	// Return to pool
	ReleaseLevelNode(node)

	// Get again - should be recycled
	node2 := AcquireLevelNode()
	if node2 == nil {
		t.Error("Expected non-nil node from pool")
	}
}

func TestNewOrderNodePooled(t *testing.T) {
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
	}

	node := NewOrderNodePooled(order)
	if node == nil {
		t.Error("Expected non-nil node")
	}
	if node.ID != 1 {
		t.Errorf("Expected ID 1, got %d", node.ID)
	}
	if node.Price != 10000 {
		t.Errorf("Expected price 10000, got %d", node.Price)
	}

	ReleaseOrderNode(node)
}

func TestNewLevelNodePooled(t *testing.T) {
	node := NewLevelNodePooled(LevelTypeBid, 10000)
	if node == nil {
		t.Error("Expected non-nil node")
	}
	if node.Price != 10000 {
		t.Errorf("Expected price 10000, got %d", node.Price)
	}
	if !node.IsBid() {
		t.Error("Expected bid level")
	}

	ReleaseLevelNode(node)
}

// Benchmark comparing pooled vs non-pooled allocation

func BenchmarkOrderNodeNonPooled(b *testing.B) {
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := NewOrderNode(order)
		_ = node
	}
}

func BenchmarkOrderNodePooled(b *testing.B) {
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := NewOrderNodePooled(order)
		ReleaseOrderNode(node)
	}
}

func BenchmarkLevelNodeNonPooled(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := NewLevelNode(LevelTypeBid, 10000)
		_ = node
	}
}

func BenchmarkLevelNodePooled(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node := NewLevelNodePooled(LevelTypeBid, 10000)
		ReleaseLevelNode(node)
	}
}

func BenchmarkOrderNodePooledParallel(b *testing.B) {
	order := Order{
		ID:       1,
		SymbolID: 1,
		Type:     OrderTypeLimit,
		Side:     OrderSideBuy,
		Price:    10000,
		Quantity: 100,
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			node := NewOrderNodePooled(order)
			ReleaseOrderNode(node)
		}
	})
}
