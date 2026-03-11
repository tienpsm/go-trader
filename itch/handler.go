// Package itch provides a handler for NASDAQ ITCH protocol messages.
// ITCH is a direct data-feed protocol that delivers real-time market data.
package itch

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// Message types for ITCH protocol
const (
	MessageTypeSystemEvent            = 'S'
	MessageTypeStockDirectory         = 'R'
	MessageTypeStockTradingAction     = 'H'
	MessageTypeRegSHO                 = 'Y'
	MessageTypeMarketParticipantPos   = 'L'
	MessageTypeMWCBDecline            = 'V'
	MessageTypeMWCBStatus             = 'W'
	MessageTypeIPOQuoting             = 'K'
	MessageTypeAddOrder               = 'A'
	MessageTypeAddOrderMPID           = 'F'
	MessageTypeOrderExecuted          = 'E'
	MessageTypeOrderExecutedWithPrice = 'C'
	MessageTypeOrderCancel            = 'X'
	MessageTypeOrderDelete            = 'D'
	MessageTypeOrderReplace           = 'U'
	MessageTypeTrade                  = 'P'
	MessageTypeCrossTrade             = 'Q'
	MessageTypeBrokenTrade            = 'B'
	MessageTypeNOII                   = 'I'
	MessageTypeRPII                   = 'N'
)

// Common errors
var (
	ErrInvalidMessage      = errors.New("invalid message")
	ErrUnknownMessageType  = errors.New("unknown message type")
	ErrInsufficientData    = errors.New("insufficient data")
)

// SystemEventMessage represents a system event message
type SystemEventMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	EventCode      byte
}

// StockDirectoryMessage represents a stock directory message
type StockDirectoryMessage struct {
	Type                       byte
	StockLocate                uint16
	TrackingNumber             uint16
	Timestamp                  uint64
	Stock                      [8]byte
	MarketCategory             byte
	FinancialStatusIndicator   byte
	RoundLotSize               uint32
	RoundLotsOnly              byte
	IssueClassification        byte
	IssueSubType               [2]byte
	Authenticity               byte
	ShortSaleThresholdIndicator byte
	IPOFlag                    byte
	LULDReferencePriceTier     byte
	ETPFlag                    byte
	ETPLeverageFactor          uint32
	InverseIndicator           byte
}

// StockTradingActionMessage represents a stock trading action message
type StockTradingActionMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	Stock          [8]byte
	TradingState   byte
	Reserved       byte
	Reason         byte
}

// RegSHOMessage represents a Reg SHO restriction message
type RegSHOMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	Stock          [8]byte
	RegSHOAction   byte
}

// MarketParticipantPositionMessage represents a market participant position message
type MarketParticipantPositionMessage struct {
	Type                   byte
	StockLocate            uint16
	TrackingNumber         uint16
	Timestamp              uint64
	MPID                   [4]byte
	Stock                  [8]byte
	PrimaryMarketMaker     byte
	MarketMakerMode        byte
	MarketParticipantState byte
}

// MWCBDeclineMessage represents a MWCB decline level message
type MWCBDeclineMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	Level1         uint64
	Level2         uint64
	Level3         uint64
}

// MWCBStatusMessage represents a MWCB status message
type MWCBStatusMessage struct {
	Type          byte
	StockLocate   uint16
	TrackingNumber uint16
	Timestamp     uint64
	BreachedLevel byte
}

// IPOQuotingMessage represents an IPO quoting period update message
type IPOQuotingMessage struct {
	Type               byte
	StockLocate        uint16
	TrackingNumber     uint16
	Timestamp          uint64
	Stock              [8]byte
	IPOReleaseTime     uint32
	IPOReleaseQualifier byte
	IPOPrice           uint32
}

// AddOrderMessage represents an add order message
type AddOrderMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
	BuySellIndicator     byte
	Shares               uint32
	Stock                [8]byte
	Price                uint32
}

// AddOrderMPIDMessage represents an add order with MPID attribution message
type AddOrderMPIDMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
	BuySellIndicator     byte
	Shares               uint32
	Stock                [8]byte
	Price                uint32
	Attribution          byte
}

// OrderExecutedMessage represents an order executed message
type OrderExecutedMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
	ExecutedShares       uint32
	MatchNumber          uint64
}

// OrderExecutedWithPriceMessage represents an order executed with price message
type OrderExecutedWithPriceMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
	ExecutedShares       uint32
	MatchNumber          uint64
	Printable            byte
	ExecutionPrice       uint32
}

// OrderCancelMessage represents an order cancel message
type OrderCancelMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
	CanceledShares       uint32
}

// OrderDeleteMessage represents an order delete message
type OrderDeleteMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
}

// OrderReplaceMessage represents an order replace message
type OrderReplaceMessage struct {
	Type                        byte
	StockLocate                 uint16
	TrackingNumber              uint16
	Timestamp                   uint64
	OriginalOrderReferenceNumber uint64
	NewOrderReferenceNumber     uint64
	Shares                      uint32
	Price                       uint32
}

// TradeMessage represents a trade message
type TradeMessage struct {
	Type                 byte
	StockLocate          uint16
	TrackingNumber       uint16
	Timestamp            uint64
	OrderReferenceNumber uint64
	BuySellIndicator     byte
	Shares               uint32
	Stock                [8]byte
	Price                uint32
	MatchNumber          uint64
}

// CrossTradeMessage represents a cross trade message
type CrossTradeMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	Shares         uint64
	Stock          [8]byte
	CrossPrice     uint32
	MatchNumber    uint64
	CrossType      byte
}

// BrokenTradeMessage represents a broken trade message
type BrokenTradeMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	MatchNumber    uint64
}

// NOIIMessage represents a Net Order Imbalance Indicator message
type NOIIMessage struct {
	Type               byte
	StockLocate        uint16
	TrackingNumber     uint16
	Timestamp          uint64
	PairedShares       uint64
	ImbalanceShares    uint64
	ImbalanceDirection byte
	Stock              [8]byte
	FarPrice           uint32
	NearPrice          uint32
	CurrentRefPrice    uint32
	CrossType          byte
	PriceVariationIndicator byte
}

// RPIIMessage represents a Retail Price Improvement Indicator message
type RPIIMessage struct {
	Type           byte
	StockLocate    uint16
	TrackingNumber uint16
	Timestamp      uint64
	Stock          [8]byte
	InterestFlag   byte
}

// Handler is the interface for handling ITCH messages
type Handler interface {
	OnSystemEvent(msg SystemEventMessage) error
	OnStockDirectory(msg StockDirectoryMessage) error
	OnStockTradingAction(msg StockTradingActionMessage) error
	OnRegSHO(msg RegSHOMessage) error
	OnMarketParticipantPosition(msg MarketParticipantPositionMessage) error
	OnMWCBDecline(msg MWCBDeclineMessage) error
	OnMWCBStatus(msg MWCBStatusMessage) error
	OnIPOQuoting(msg IPOQuotingMessage) error
	OnAddOrder(msg AddOrderMessage) error
	OnAddOrderMPID(msg AddOrderMPIDMessage) error
	OnOrderExecuted(msg OrderExecutedMessage) error
	OnOrderExecutedWithPrice(msg OrderExecutedWithPriceMessage) error
	OnOrderCancel(msg OrderCancelMessage) error
	OnOrderDelete(msg OrderDeleteMessage) error
	OnOrderReplace(msg OrderReplaceMessage) error
	OnTrade(msg TradeMessage) error
	OnCrossTrade(msg CrossTradeMessage) error
	OnBrokenTrade(msg BrokenTradeMessage) error
	OnNOII(msg NOIIMessage) error
	OnRPII(msg RPIIMessage) error
	OnUnknownMessage(msgType byte, data []byte) error
}

// DefaultHandler is a no-op implementation of Handler
type DefaultHandler struct{}

func (h *DefaultHandler) OnSystemEvent(msg SystemEventMessage) error                     { return nil }
func (h *DefaultHandler) OnStockDirectory(msg StockDirectoryMessage) error               { return nil }
func (h *DefaultHandler) OnStockTradingAction(msg StockTradingActionMessage) error       { return nil }
func (h *DefaultHandler) OnRegSHO(msg RegSHOMessage) error                               { return nil }
func (h *DefaultHandler) OnMarketParticipantPosition(msg MarketParticipantPositionMessage) error { return nil }
func (h *DefaultHandler) OnMWCBDecline(msg MWCBDeclineMessage) error                     { return nil }
func (h *DefaultHandler) OnMWCBStatus(msg MWCBStatusMessage) error                       { return nil }
func (h *DefaultHandler) OnIPOQuoting(msg IPOQuotingMessage) error                       { return nil }
func (h *DefaultHandler) OnAddOrder(msg AddOrderMessage) error                           { return nil }
func (h *DefaultHandler) OnAddOrderMPID(msg AddOrderMPIDMessage) error                   { return nil }
func (h *DefaultHandler) OnOrderExecuted(msg OrderExecutedMessage) error                 { return nil }
func (h *DefaultHandler) OnOrderExecutedWithPrice(msg OrderExecutedWithPriceMessage) error { return nil }
func (h *DefaultHandler) OnOrderCancel(msg OrderCancelMessage) error                     { return nil }
func (h *DefaultHandler) OnOrderDelete(msg OrderDeleteMessage) error                     { return nil }
func (h *DefaultHandler) OnOrderReplace(msg OrderReplaceMessage) error                   { return nil }
func (h *DefaultHandler) OnTrade(msg TradeMessage) error                                 { return nil }
func (h *DefaultHandler) OnCrossTrade(msg CrossTradeMessage) error                       { return nil }
func (h *DefaultHandler) OnBrokenTrade(msg BrokenTradeMessage) error                     { return nil }
func (h *DefaultHandler) OnNOII(msg NOIIMessage) error                                   { return nil }
func (h *DefaultHandler) OnRPII(msg RPIIMessage) error                                   { return nil }
func (h *DefaultHandler) OnUnknownMessage(msgType byte, data []byte) error               { return nil }

// Parser parses ITCH protocol messages
type Parser struct {
	handler Handler
}

// NewParser creates a new ITCH parser
func NewParser(handler Handler) *Parser {
	return &Parser{handler: handler}
}

// Parse parses a single ITCH message
func (p *Parser) Parse(data []byte) (int, error) {
	if len(data) < 1 {
		return 0, ErrInsufficientData
	}

	msgType := data[0]
	var consumed int
	var err error

	switch msgType {
	case MessageTypeSystemEvent:
		consumed, err = p.parseSystemEvent(data)
	case MessageTypeStockDirectory:
		consumed, err = p.parseStockDirectory(data)
	case MessageTypeStockTradingAction:
		consumed, err = p.parseStockTradingAction(data)
	case MessageTypeRegSHO:
		consumed, err = p.parseRegSHO(data)
	case MessageTypeMarketParticipantPos:
		consumed, err = p.parseMarketParticipantPosition(data)
	case MessageTypeMWCBDecline:
		consumed, err = p.parseMWCBDecline(data)
	case MessageTypeMWCBStatus:
		consumed, err = p.parseMWCBStatus(data)
	case MessageTypeIPOQuoting:
		consumed, err = p.parseIPOQuoting(data)
	case MessageTypeAddOrder:
		consumed, err = p.parseAddOrder(data)
	case MessageTypeAddOrderMPID:
		consumed, err = p.parseAddOrderMPID(data)
	case MessageTypeOrderExecuted:
		consumed, err = p.parseOrderExecuted(data)
	case MessageTypeOrderExecutedWithPrice:
		consumed, err = p.parseOrderExecutedWithPrice(data)
	case MessageTypeOrderCancel:
		consumed, err = p.parseOrderCancel(data)
	case MessageTypeOrderDelete:
		consumed, err = p.parseOrderDelete(data)
	case MessageTypeOrderReplace:
		consumed, err = p.parseOrderReplace(data)
	case MessageTypeTrade:
		consumed, err = p.parseTrade(data)
	case MessageTypeCrossTrade:
		consumed, err = p.parseCrossTrade(data)
	case MessageTypeBrokenTrade:
		consumed, err = p.parseBrokenTrade(data)
	case MessageTypeNOII:
		consumed, err = p.parseNOII(data)
	case MessageTypeRPII:
		consumed, err = p.parseRPII(data)
	default:
		err = p.handler.OnUnknownMessage(msgType, data)
		consumed = len(data)
	}

	return consumed, err
}

// ParseAll parses all ITCH messages in the data
func (p *Parser) ParseAll(data []byte) (int, int, error) {
	totalConsumed := 0
	messageCount := 0

	for len(data) > 0 {
		consumed, err := p.Parse(data)
		if err != nil {
			if err == ErrInsufficientData {
				break
			}
			return totalConsumed, messageCount, err
		}
		if consumed == 0 {
			break
		}
		totalConsumed += consumed
		messageCount++
		data = data[consumed:]
	}

	return totalConsumed, messageCount, nil
}

// Helper functions for parsing

func readUint16BE(data []byte) uint16 {
	return binary.BigEndian.Uint16(data)
}

func readUint32BE(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func readUint48BE(data []byte) uint64 {
	return uint64(data[0])<<40 | uint64(data[1])<<32 | uint64(data[2])<<24 |
		uint64(data[3])<<16 | uint64(data[4])<<8 | uint64(data[5])
}

func readUint64BE(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

func (p *Parser) parseSystemEvent(data []byte) (int, error) {
	const size = 12
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := SystemEventMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		EventCode:      data[11],
	}

	return size, p.handler.OnSystemEvent(msg)
}

func (p *Parser) parseStockDirectory(data []byte) (int, error) {
	const size = 39
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := StockDirectoryMessage{
		Type:                        data[0],
		StockLocate:                 readUint16BE(data[1:3]),
		TrackingNumber:              readUint16BE(data[3:5]),
		Timestamp:                   readUint48BE(data[5:11]),
		MarketCategory:              data[19],
		FinancialStatusIndicator:    data[20],
		RoundLotSize:                readUint32BE(data[21:25]),
		RoundLotsOnly:               data[25],
		IssueClassification:         data[26],
		Authenticity:                data[29],
		ShortSaleThresholdIndicator: data[30],
		IPOFlag:                     data[31],
		LULDReferencePriceTier:      data[32],
		ETPFlag:                     data[33],
		ETPLeverageFactor:           readUint32BE(data[34:38]),
		InverseIndicator:            data[38],
	}
	copy(msg.Stock[:], data[11:19])
	copy(msg.IssueSubType[:], data[27:29])

	return size, p.handler.OnStockDirectory(msg)
}

func (p *Parser) parseStockTradingAction(data []byte) (int, error) {
	const size = 25
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := StockTradingActionMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		TradingState:   data[19],
		Reserved:       data[20],
		Reason:         data[21],
	}
	copy(msg.Stock[:], data[11:19])

	return size, p.handler.OnStockTradingAction(msg)
}

func (p *Parser) parseRegSHO(data []byte) (int, error) {
	const size = 20
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := RegSHOMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		RegSHOAction:   data[19],
	}
	copy(msg.Stock[:], data[11:19])

	return size, p.handler.OnRegSHO(msg)
}

func (p *Parser) parseMarketParticipantPosition(data []byte) (int, error) {
	const size = 26
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := MarketParticipantPositionMessage{
		Type:                   data[0],
		StockLocate:            readUint16BE(data[1:3]),
		TrackingNumber:         readUint16BE(data[3:5]),
		Timestamp:              readUint48BE(data[5:11]),
		PrimaryMarketMaker:     data[23],
		MarketMakerMode:        data[24],
		MarketParticipantState: data[25],
	}
	copy(msg.MPID[:], data[11:15])
	copy(msg.Stock[:], data[15:23])

	return size, p.handler.OnMarketParticipantPosition(msg)
}

func (p *Parser) parseMWCBDecline(data []byte) (int, error) {
	const size = 35
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := MWCBDeclineMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		Level1:         readUint64BE(data[11:19]),
		Level2:         readUint64BE(data[19:27]),
		Level3:         readUint64BE(data[27:35]),
	}

	return size, p.handler.OnMWCBDecline(msg)
}

func (p *Parser) parseMWCBStatus(data []byte) (int, error) {
	const size = 12
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := MWCBStatusMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		BreachedLevel:  data[11],
	}

	return size, p.handler.OnMWCBStatus(msg)
}

func (p *Parser) parseIPOQuoting(data []byte) (int, error) {
	const size = 28
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := IPOQuotingMessage{
		Type:                data[0],
		StockLocate:         readUint16BE(data[1:3]),
		TrackingNumber:      readUint16BE(data[3:5]),
		Timestamp:           readUint48BE(data[5:11]),
		IPOReleaseTime:      readUint32BE(data[19:23]),
		IPOReleaseQualifier: data[23],
		IPOPrice:            readUint32BE(data[24:28]),
	}
	copy(msg.Stock[:], data[11:19])

	return size, p.handler.OnIPOQuoting(msg)
}

func (p *Parser) parseAddOrder(data []byte) (int, error) {
	const size = 36
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := AddOrderMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
		BuySellIndicator:     data[19],
		Shares:               readUint32BE(data[20:24]),
		Price:                readUint32BE(data[32:36]),
	}
	copy(msg.Stock[:], data[24:32])

	return size, p.handler.OnAddOrder(msg)
}

func (p *Parser) parseAddOrderMPID(data []byte) (int, error) {
	const size = 40
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := AddOrderMPIDMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
		BuySellIndicator:     data[19],
		Shares:               readUint32BE(data[20:24]),
		Price:                readUint32BE(data[32:36]),
		Attribution:          data[36],
	}
	copy(msg.Stock[:], data[24:32])

	return size, p.handler.OnAddOrderMPID(msg)
}

func (p *Parser) parseOrderExecuted(data []byte) (int, error) {
	const size = 31
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := OrderExecutedMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
		ExecutedShares:       readUint32BE(data[19:23]),
		MatchNumber:          readUint64BE(data[23:31]),
	}

	return size, p.handler.OnOrderExecuted(msg)
}

func (p *Parser) parseOrderExecutedWithPrice(data []byte) (int, error) {
	const size = 36
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := OrderExecutedWithPriceMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
		ExecutedShares:       readUint32BE(data[19:23]),
		MatchNumber:          readUint64BE(data[23:31]),
		Printable:            data[31],
		ExecutionPrice:       readUint32BE(data[32:36]),
	}

	return size, p.handler.OnOrderExecutedWithPrice(msg)
}

func (p *Parser) parseOrderCancel(data []byte) (int, error) {
	const size = 23
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := OrderCancelMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
		CanceledShares:       readUint32BE(data[19:23]),
	}

	return size, p.handler.OnOrderCancel(msg)
}

func (p *Parser) parseOrderDelete(data []byte) (int, error) {
	const size = 19
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := OrderDeleteMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
	}

	return size, p.handler.OnOrderDelete(msg)
}

func (p *Parser) parseOrderReplace(data []byte) (int, error) {
	const size = 35
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := OrderReplaceMessage{
		Type:                         data[0],
		StockLocate:                  readUint16BE(data[1:3]),
		TrackingNumber:               readUint16BE(data[3:5]),
		Timestamp:                    readUint48BE(data[5:11]),
		OriginalOrderReferenceNumber: readUint64BE(data[11:19]),
		NewOrderReferenceNumber:      readUint64BE(data[19:27]),
		Shares:                       readUint32BE(data[27:31]),
		Price:                        readUint32BE(data[31:35]),
	}

	return size, p.handler.OnOrderReplace(msg)
}

func (p *Parser) parseTrade(data []byte) (int, error) {
	const size = 44
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := TradeMessage{
		Type:                 data[0],
		StockLocate:          readUint16BE(data[1:3]),
		TrackingNumber:       readUint16BE(data[3:5]),
		Timestamp:            readUint48BE(data[5:11]),
		OrderReferenceNumber: readUint64BE(data[11:19]),
		BuySellIndicator:     data[19],
		Shares:               readUint32BE(data[20:24]),
		Price:                readUint32BE(data[32:36]),
		MatchNumber:          readUint64BE(data[36:44]),
	}
	copy(msg.Stock[:], data[24:32])

	return size, p.handler.OnTrade(msg)
}

func (p *Parser) parseCrossTrade(data []byte) (int, error) {
	const size = 40
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := CrossTradeMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		Shares:         readUint64BE(data[11:19]),
		CrossPrice:     readUint32BE(data[27:31]),
		MatchNumber:    readUint64BE(data[31:39]),
		CrossType:      data[39],
	}
	copy(msg.Stock[:], data[19:27])

	return size, p.handler.OnCrossTrade(msg)
}

func (p *Parser) parseBrokenTrade(data []byte) (int, error) {
	const size = 19
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := BrokenTradeMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		MatchNumber:    readUint64BE(data[11:19]),
	}

	return size, p.handler.OnBrokenTrade(msg)
}

func (p *Parser) parseNOII(data []byte) (int, error) {
	const size = 50
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := NOIIMessage{
		Type:                    data[0],
		StockLocate:             readUint16BE(data[1:3]),
		TrackingNumber:          readUint16BE(data[3:5]),
		Timestamp:               readUint48BE(data[5:11]),
		PairedShares:            readUint64BE(data[11:19]),
		ImbalanceShares:         readUint64BE(data[19:27]),
		ImbalanceDirection:      data[27],
		FarPrice:                readUint32BE(data[36:40]),
		NearPrice:               readUint32BE(data[40:44]),
		CurrentRefPrice:         readUint32BE(data[44:48]),
		CrossType:               data[48],
		PriceVariationIndicator: data[49],
	}
	copy(msg.Stock[:], data[28:36])

	return size, p.handler.OnNOII(msg)
}

func (p *Parser) parseRPII(data []byte) (int, error) {
	const size = 20
	if len(data) < size {
		return 0, ErrInsufficientData
	}

	msg := RPIIMessage{
		Type:           data[0],
		StockLocate:    readUint16BE(data[1:3]),
		TrackingNumber: readUint16BE(data[3:5]),
		Timestamp:      readUint48BE(data[5:11]),
		InterestFlag:   data[19],
	}
	copy(msg.Stock[:], data[11:19])

	return size, p.handler.OnRPII(msg)
}

// String returns a string representation of the message
func (msg SystemEventMessage) String() string {
	return fmt.Sprintf("SystemEvent{EventCode: %c, Timestamp: %d}", msg.EventCode, msg.Timestamp)
}

// String returns a string representation of the message
func (msg AddOrderMessage) String() string {
	stock := string(msg.Stock[:])
	side := "BUY"
	if msg.BuySellIndicator == 'S' {
		side = "SELL"
	}
	return fmt.Sprintf("AddOrder{Ref: %d, Side: %s, Shares: %d, Stock: %s, Price: %d}",
		msg.OrderReferenceNumber, side, msg.Shares, stock, msg.Price)
}

// MessageStats tracks statistics about parsed messages
type MessageStats struct {
	TotalMessages         int
	SystemEvents          int
	StockDirectories      int
	StockTradingActions   int
	RegSHO                int
	MarketParticipantPos  int
	MWCBDecline           int
	MWCBStatus            int
	IPOQuoting            int
	AddOrders             int
	AddOrderMPID          int
	OrderExecuted         int
	OrderExecutedPrice    int
	OrderCancels          int
	OrderDeletes          int
	OrderReplaces         int
	Trades                int
	CrossTrades           int
	BrokenTrades          int
	NOII                  int
	RPII                  int
	UnknownMessages       int
}

// StatsHandler is a handler that collects message statistics
type StatsHandler struct {
	DefaultHandler
	Stats MessageStats
}

func (h *StatsHandler) OnSystemEvent(msg SystemEventMessage) error {
	h.Stats.SystemEvents++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnStockDirectory(msg StockDirectoryMessage) error {
	h.Stats.StockDirectories++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnStockTradingAction(msg StockTradingActionMessage) error {
	h.Stats.StockTradingActions++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnRegSHO(msg RegSHOMessage) error {
	h.Stats.RegSHO++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnMarketParticipantPosition(msg MarketParticipantPositionMessage) error {
	h.Stats.MarketParticipantPos++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnMWCBDecline(msg MWCBDeclineMessage) error {
	h.Stats.MWCBDecline++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnMWCBStatus(msg MWCBStatusMessage) error {
	h.Stats.MWCBStatus++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnIPOQuoting(msg IPOQuotingMessage) error {
	h.Stats.IPOQuoting++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnAddOrder(msg AddOrderMessage) error {
	h.Stats.AddOrders++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnAddOrderMPID(msg AddOrderMPIDMessage) error {
	h.Stats.AddOrderMPID++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnOrderExecuted(msg OrderExecutedMessage) error {
	h.Stats.OrderExecuted++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnOrderExecutedWithPrice(msg OrderExecutedWithPriceMessage) error {
	h.Stats.OrderExecutedPrice++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnOrderCancel(msg OrderCancelMessage) error {
	h.Stats.OrderCancels++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnOrderDelete(msg OrderDeleteMessage) error {
	h.Stats.OrderDeletes++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnOrderReplace(msg OrderReplaceMessage) error {
	h.Stats.OrderReplaces++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnTrade(msg TradeMessage) error {
	h.Stats.Trades++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnCrossTrade(msg CrossTradeMessage) error {
	h.Stats.CrossTrades++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnBrokenTrade(msg BrokenTradeMessage) error {
	h.Stats.BrokenTrades++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnNOII(msg NOIIMessage) error {
	h.Stats.NOII++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnRPII(msg RPIIMessage) error {
	h.Stats.RPII++
	h.Stats.TotalMessages++
	return nil
}

func (h *StatsHandler) OnUnknownMessage(msgType byte, data []byte) error {
	h.Stats.UnknownMessages++
	h.Stats.TotalMessages++
	return nil
}

// ParseFile parses an ITCH file from the given path
func ParseFile(filename string, handler Handler) (int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return ParseReader(file, handler)
}

// ParseReader parses ITCH messages from an io.Reader
// The format expects each message to be prefixed with a 2-byte big-endian length
func ParseReader(reader io.Reader, handler Handler) (int64, error) {
	parser := NewParser(handler)
	var totalBytes int64
	
	// Use buffered reader for efficient reading
	bufReader := bufio.NewReaderSize(reader, 64*1024) // 64KB buffer
	lengthBuf := make([]byte, 2)

	for {
		// Read 2-byte length prefix
		_, err := io.ReadFull(bufReader, lengthBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return totalBytes, err
		}
		
		// Parse message length
		msgLen := binary.BigEndian.Uint16(lengthBuf)
		if msgLen == 0 {
			continue
		}
		
		// Read the message
		msgBuf := make([]byte, msgLen)
		_, err = io.ReadFull(bufReader, msgBuf)
		if err != nil {
			return totalBytes, err
		}
		
		// Parse the message
		_, parseErr := parser.Parse(msgBuf)
		if parseErr != nil {
			return totalBytes, parseErr
		}
		
		// Count total bytes including length prefix
		totalBytes += int64(2 + msgLen)
	}

	return totalBytes, nil
}
