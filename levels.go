package dvotcWS

type LevelData struct {
	Levels     []Level `json:"levels"`
	LastUpdate int64   `json:"lastUpdate"`
	QuoteID    string  `json:"quoteId"`
	Market     string  `json:"market"`
}

type Level struct {
	SellPrice   float64 `json:"sellPrice"`
	BuyPrice    float64 `json:"buyPrice"`
	MaxQuantity float64 `json:"maxQuantity"`
}

type SubscribeLevelData = Subscription[LevelData]

func (dvotc *DVOTCClient) SubscribeLevels(symbol string) (*SubscribeLevelData, error) {
	sub := &SubscribeLevelData{
		Data:    make(chan LevelData),
		done:    make(chan struct{}),
		topic:   symbol,
		event:   "levels",
		chanIdx: 0,
		dvotc:   dvotc,
	}

	chanIdx, ok := checkConnExistAndReturnIdx(&dvotc.safeChanStore, &dvotc.chanMutex, sub.event, sub.topic, sub.Data)
	sub.chanIdx = chanIdx
	if ok {
		// just add a new channel to list to listen to subscriptions
		return sub, nil
	}

	conn, err := dvotc.getConnOrReuse()
	if err != nil {
		return nil, err
	}
	payload := Payload{
		Type:  MessageTypeSubscribe,
		Event: "levels",
		Topic: symbol,
	}

	err = conn.WriteJSON(payload)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
