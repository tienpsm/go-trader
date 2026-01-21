// Package matching provides a high-performance order matching engine.
// This is a Go port of the CppTrader C++ library.
package matching

import "errors"

// ErrorCode represents the possible error codes from the matching engine.
type ErrorCode uint8

const (
	// ErrorOK indicates success
	ErrorOK ErrorCode = iota
	// ErrorSymbolDuplicate indicates the symbol already exists
	ErrorSymbolDuplicate
	// ErrorSymbolNotFound indicates the symbol was not found
	ErrorSymbolNotFound
	// ErrorOrderBookDuplicate indicates the order book already exists
	ErrorOrderBookDuplicate
	// ErrorOrderBookNotFound indicates the order book was not found
	ErrorOrderBookNotFound
	// ErrorOrderDuplicate indicates the order already exists
	ErrorOrderDuplicate
	// ErrorOrderNotFound indicates the order was not found
	ErrorOrderNotFound
	// ErrorOrderIDInvalid indicates the order ID is invalid
	ErrorOrderIDInvalid
	// ErrorOrderTypeInvalid indicates the order type is invalid
	ErrorOrderTypeInvalid
	// ErrorOrderParameterInvalid indicates an order parameter is invalid
	ErrorOrderParameterInvalid
	// ErrorOrderQuantityInvalid indicates the order quantity is invalid
	ErrorOrderQuantityInvalid
)

// Error messages for matching engine errors
var (
	ErrSymbolDuplicate       = errors.New("symbol duplicate")
	ErrSymbolNotFound        = errors.New("symbol not found")
	ErrOrderBookDuplicate    = errors.New("order book duplicate")
	ErrOrderBookNotFound     = errors.New("order book not found")
	ErrOrderDuplicate        = errors.New("order duplicate")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderIDInvalid        = errors.New("order ID invalid")
	ErrOrderTypeInvalid      = errors.New("order type invalid")
	ErrOrderParameterInvalid = errors.New("order parameter invalid")
	ErrOrderQuantityInvalid  = errors.New("order quantity invalid")
)

// String returns the string representation of an ErrorCode
func (e ErrorCode) String() string {
	switch e {
	case ErrorOK:
		return "OK"
	case ErrorSymbolDuplicate:
		return "SYMBOL_DUPLICATE"
	case ErrorSymbolNotFound:
		return "SYMBOL_NOT_FOUND"
	case ErrorOrderBookDuplicate:
		return "ORDER_BOOK_DUPLICATE"
	case ErrorOrderBookNotFound:
		return "ORDER_BOOK_NOT_FOUND"
	case ErrorOrderDuplicate:
		return "ORDER_DUPLICATE"
	case ErrorOrderNotFound:
		return "ORDER_NOT_FOUND"
	case ErrorOrderIDInvalid:
		return "ORDER_ID_INVALID"
	case ErrorOrderTypeInvalid:
		return "ORDER_TYPE_INVALID"
	case ErrorOrderParameterInvalid:
		return "ORDER_PARAMETER_INVALID"
	case ErrorOrderQuantityInvalid:
		return "ORDER_QUANTITY_INVALID"
	default:
		return "UNKNOWN"
	}
}

// Error returns the error for this ErrorCode, or nil if OK
func (e ErrorCode) Error() error {
	switch e {
	case ErrorOK:
		return nil
	case ErrorSymbolDuplicate:
		return ErrSymbolDuplicate
	case ErrorSymbolNotFound:
		return ErrSymbolNotFound
	case ErrorOrderBookDuplicate:
		return ErrOrderBookDuplicate
	case ErrorOrderBookNotFound:
		return ErrOrderBookNotFound
	case ErrorOrderDuplicate:
		return ErrOrderDuplicate
	case ErrorOrderNotFound:
		return ErrOrderNotFound
	case ErrorOrderIDInvalid:
		return ErrOrderIDInvalid
	case ErrorOrderTypeInvalid:
		return ErrOrderTypeInvalid
	case ErrorOrderParameterInvalid:
		return ErrOrderParameterInvalid
	case ErrorOrderQuantityInvalid:
		return ErrOrderQuantityInvalid
	default:
		return errors.New("unknown error")
	}
}
