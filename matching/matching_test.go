package matching

import (
	"testing"
)

func TestNewSymbol(t *testing.T) {
	symbol := NewSymbol(1, "AAPL")
	if symbol.ID != 1 {
		t.Errorf("Expected ID 1, got %d", symbol.ID)
	}
	if symbol.Name != "AAPL" {
		t.Errorf("Expected Name AAPL, got %s", symbol.Name)
	}
}

func TestNewSymbolTruncation(t *testing.T) {
	symbol := NewSymbol(1, "LONGSYMBOLNAME")
	if len(symbol.Name) > 8 {
		t.Errorf("Expected name to be truncated to 8 chars, got %s", symbol.Name)
	}
}

func TestOrderSideString(t *testing.T) {
	if OrderSideBuy.String() != "BUY" {
		t.Errorf("Expected BUY, got %s", OrderSideBuy.String())
	}
	if OrderSideSell.String() != "SELL" {
		t.Errorf("Expected SELL, got %s", OrderSideSell.String())
	}
}

func TestOrderTypeString(t *testing.T) {
	tests := []struct {
		orderType OrderType
		expected  string
	}{
		{OrderTypeMarket, "MARKET"},
		{OrderTypeLimit, "LIMIT"},
		{OrderTypeStop, "STOP"},
		{OrderTypeStopLimit, "STOP_LIMIT"},
		{OrderTypeTrailingStop, "TRAILING_STOP"},
		{OrderTypeTrailingStopLimit, "TRAILING_STOP_LIMIT"},
	}

	for _, tt := range tests {
		if tt.orderType.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.orderType.String())
		}
	}
}

func TestNewLimitOrder(t *testing.T) {
	order := NewLimitOrder(1, 100, OrderSideBuy, 5000, 10)
	if order.ID != 1 {
		t.Errorf("Expected ID 1, got %d", order.ID)
	}
	if order.SymbolID != 100 {
		t.Errorf("Expected SymbolID 100, got %d", order.SymbolID)
	}
	if order.Type != OrderTypeLimit {
		t.Errorf("Expected type LIMIT, got %s", order.Type)
	}
	if order.Side != OrderSideBuy {
		t.Errorf("Expected side BUY, got %s", order.Side)
	}
	if order.Price != 5000 {
		t.Errorf("Expected price 5000, got %d", order.Price)
	}
	if order.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %d", order.Quantity)
	}
	if order.LeavesQuantity != 10 {
		t.Errorf("Expected leaves quantity 10, got %d", order.LeavesQuantity)
	}
}

func TestNewMarketOrder(t *testing.T) {
	order := NewMarketOrder(1, 100, OrderSideSell, 50)
	if order.Type != OrderTypeMarket {
		t.Errorf("Expected type MARKET, got %s", order.Type)
	}
	if order.Price != 0 {
		t.Errorf("Expected price 0, got %d", order.Price)
	}
}

func TestOrderIsHelpers(t *testing.T) {
	limitOrder := NewLimitOrder(1, 100, OrderSideBuy, 5000, 10)
	if !limitOrder.IsLimit() {
		t.Error("Expected IsLimit to be true")
	}
	if !limitOrder.IsBuy() {
		t.Error("Expected IsBuy to be true")
	}
	if limitOrder.IsSell() {
		t.Error("Expected IsSell to be false")
	}

	marketOrder := NewMarketOrder(2, 100, OrderSideSell, 50)
	if !marketOrder.IsMarket() {
		t.Error("Expected IsMarket to be true")
	}
	if !marketOrder.IsSell() {
		t.Error("Expected IsSell to be true")
	}
}

func TestOrderVisibleQuantity(t *testing.T) {
	order := NewLimitOrder(1, 100, OrderSideBuy, 5000, 100)
	
	// Default: full visibility
	if order.VisibleQuantity() != 100 {
		t.Errorf("Expected visible quantity 100, got %d", order.VisibleQuantity())
	}
	if order.HiddenQuantity() != 0 {
		t.Errorf("Expected hidden quantity 0, got %d", order.HiddenQuantity())
	}
	
	// Iceberg order
	order.MaxVisibleQuantity = 20
	if order.VisibleQuantity() != 20 {
		t.Errorf("Expected visible quantity 20, got %d", order.VisibleQuantity())
	}
	if order.HiddenQuantity() != 80 {
		t.Errorf("Expected hidden quantity 80, got %d", order.HiddenQuantity())
	}
	
	// Hidden order
	order.MaxVisibleQuantity = 0
	if order.VisibleQuantity() != 0 {
		t.Errorf("Expected visible quantity 0, got %d", order.VisibleQuantity())
	}
	if order.HiddenQuantity() != 100 {
		t.Errorf("Expected hidden quantity 100, got %d", order.HiddenQuantity())
	}
}

func TestLevelType(t *testing.T) {
	level := NewLevel(LevelTypeBid, 5000)
	if !level.IsBid() {
		t.Error("Expected IsBid to be true")
	}
	if level.IsAsk() {
		t.Error("Expected IsAsk to be false")
	}
}

func TestUpdateType(t *testing.T) {
	if UpdateNone.String() != "NONE" {
		t.Errorf("Expected NONE, got %s", UpdateNone.String())
	}
	if UpdateAdd.String() != "ADD" {
		t.Errorf("Expected ADD, got %s", UpdateAdd.String())
	}
	if UpdateUpdate.String() != "UPDATE" {
		t.Errorf("Expected UPDATE, got %s", UpdateUpdate.String())
	}
	if UpdateDelete.String() != "DELETE" {
		t.Errorf("Expected DELETE, got %s", UpdateDelete.String())
	}
}

func TestErrorCode(t *testing.T) {
	if ErrorOK.String() != "OK" {
		t.Errorf("Expected OK, got %s", ErrorOK.String())
	}
	if ErrorOK.Error() != nil {
		t.Error("Expected nil error for ErrorOK")
	}
	if ErrorSymbolNotFound.Error() == nil {
		t.Error("Expected non-nil error for ErrorSymbolNotFound")
	}
}

func TestOrderList(t *testing.T) {
	list := &OrderList{}
	
	order1 := NewOrderNode(Order{ID: 1})
	order2 := NewOrderNode(Order{ID: 2})
	order3 := NewOrderNode(Order{ID: 3})
	
	list.PushBack(order1)
	list.PushBack(order2)
	list.PushBack(order3)
	
	if list.Size != 3 {
		t.Errorf("Expected size 3, got %d", list.Size)
	}
	if list.Front().ID != 1 {
		t.Errorf("Expected front ID 1, got %d", list.Front().ID)
	}
	if list.Empty() {
		t.Error("Expected list to not be empty")
	}
	
	list.Remove(order2)
	if list.Size != 2 {
		t.Errorf("Expected size 2, got %d", list.Size)
	}
	
	list.Remove(order1)
	list.Remove(order3)
	if !list.Empty() {
		t.Error("Expected list to be empty")
	}
}

func TestAVLTree(t *testing.T) {
	tree := NewAVLTree(false) // Ascending order
	
	levels := []*LevelNode{
		NewLevelNode(LevelTypeBid, 100),
		NewLevelNode(LevelTypeBid, 200),
		NewLevelNode(LevelTypeBid, 50),
		NewLevelNode(LevelTypeBid, 150),
	}
	
	for _, level := range levels {
		tree.Insert(level)
	}
	
	if tree.Size() != 4 {
		t.Errorf("Expected size 4, got %d", tree.Size())
	}
	
	// First should be 50 (ascending)
	first := tree.First()
	if first.Price != 50 {
		t.Errorf("Expected first price 50, got %d", first.Price)
	}
	
	// Find 150
	found := tree.Find(150)
	if found == nil || found.Price != 150 {
		t.Error("Expected to find price 150")
	}
	
	// Find non-existent
	notFound := tree.Find(999)
	if notFound != nil {
		t.Error("Expected nil for non-existent price")
	}
}

func TestAVLTreeDescending(t *testing.T) {
	tree := NewAVLTree(true) // Descending order
	
	levels := []*LevelNode{
		NewLevelNode(LevelTypeBid, 100),
		NewLevelNode(LevelTypeBid, 200),
		NewLevelNode(LevelTypeBid, 50),
		NewLevelNode(LevelTypeBid, 150),
	}
	
	for _, level := range levels {
		tree.Insert(level)
	}
	
	// First should be 200 (descending)
	first := tree.First()
	if first.Price != 200 {
		t.Errorf("Expected first price 200, got %d", first.Price)
	}
}

func TestAVLTreeRemove(t *testing.T) {
	tree := NewAVLTree(false)
	
	level1 := NewLevelNode(LevelTypeBid, 100)
	level2 := NewLevelNode(LevelTypeBid, 200)
	level3 := NewLevelNode(LevelTypeBid, 50)
	
	tree.Insert(level1)
	tree.Insert(level2)
	tree.Insert(level3)
	
	tree.Remove(level1)
	
	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}
	
	if tree.Find(100) != nil {
		t.Error("Expected level 100 to be removed")
	}
}
