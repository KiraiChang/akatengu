package event_types

import "akatengu/internal/pkg/enumx"

const (
	// investment
	eventInvestmentCreated eventTypeVal = "investment.created"
	eventInvestmentUpdated eventTypeVal = "investment.updated"
	eventInvestmentBought  eventTypeVal = "investment.bought"
	eventInvestmentSold    eventTypeVal = "investment.sold"
	eventDividendReceived  eventTypeVal = "investment.dividend_received"
	eventStockSplit        eventTypeVal = "investment.stock_split"
	eventFxBought          eventTypeVal = "investment.fx_bought"
	eventFxSold            eventTypeVal = "investment.fx_sold"
	eventUnrealizedMarked  eventTypeVal = "investment.unrealized_marked"
	eventRateUpdated       eventTypeVal = "investment.fx_rate_updated"
)

var validInvestmentEventType = []eventTypeVal{
	eventInvestmentCreated,
	eventInvestmentUpdated,
	eventInvestmentBought,
	eventInvestmentSold,
	eventDividendReceived,
	eventStockSplit,
	eventFxBought,
	eventFxSold,
	eventUnrealizedMarked,
	eventRateUpdated,
}

var (
	// investment
	EventInvestmentCreated = enumx.Must(string(eventInvestmentCreated), validEventTypes)
	EventInvestmentUpdated = enumx.Must(string(eventInvestmentUpdated), validEventTypes)
	EventInvestmentBought  = enumx.Must(string(eventInvestmentBought), validEventTypes)
	EventInvestmentSold    = enumx.Must(string(eventInvestmentSold), validEventTypes)
	EventDividendReceived  = enumx.Must(string(eventDividendReceived), validEventTypes)
	EventStockSplit        = enumx.Must(string(eventStockSplit), validEventTypes)
	EventFxBought          = enumx.Must(string(eventFxBought), validEventTypes)
	EventFxSold            = enumx.Must(string(eventFxSold), validEventTypes)
	EventUnrealizedMarked  = enumx.Must(string(eventUnrealizedMarked), validEventTypes)
	EventRateUpdated       = enumx.Must(string(eventRateUpdated), validEventTypes)
)
