package event_types

const (
	// investment
	EventInvestmentCreated eventTypeVal = "investment.created"
	EventInvestmentUpdated eventTypeVal = "investment.updated"
	EventInvestmentBought  eventTypeVal = "investment.bought"
	EventInvestmentSold    eventTypeVal = "investment.sold"
	EventDividendReceived  eventTypeVal = "investment.dividend_received"
	EventStockSplit        eventTypeVal = "investment.stock_split"
	EventUnrealizedMarked  eventTypeVal = "investment.unrealized_marked"
	EventRateUpdated       eventTypeVal = "investment.fx_rate_updated"
)
