package Poker

const (
	CARD_COLOR   = 0xF0 //花色掩码
	CARD_VALUE   = 0x0F //数值掩码
	Card_Invalid = 0x00
	Card_Rear    = 0xFF
)

const (
	//方
	Card_Fang_1 = iota + 0x01
	Card_Fang_2
	Card_Fang_3
	Card_Fang_4
	Card_Fang_5
	Card_Fang_6
	Card_Fang_7
	Card_Fang_8
	Card_Fang_9
	Card_Fang_10
	Card_Fang_J
	Card_Fang_Q
	Card_Fang_K
)

const (
	//梅
	Card_Mei_1 = iota + 0x11
	Card_Mei_2
	Card_Mei_3
	Card_Mei_4
	Card_Mei_5
	Card_Mei_6
	Card_Mei_7
	Card_Mei_8
	Card_Mei_9
	Card_Mei_10
	Card_Mei_J
	Card_Mei_Q
	Card_Mei_K
)

const (
	//红
	Card_Hong_1 = iota + 0x21
	Card_Hong_2
	Card_Hong_3
	Card_Hong_4
	Card_Hong_5
	Card_Hong_6
	Card_Hong_7
	Card_Hong_8
	Card_Hong_9
	Card_Hong_10
	Card_Hong_J
	Card_Hong_Q
	Card_Hong_K
)

const (
	//黑
	Card_Hei_1 = iota + 0x31
	Card_Hei_2
	Card_Hei_3
	Card_Hei_4
	Card_Hei_5
	Card_Hei_6
	Card_Hei_7
	Card_Hei_8
	Card_Hei_9
	Card_Hei_10
	Card_Hei_J
	Card_Hei_Q
	Card_Hei_K
)

const (
	// 王
	Card_King_1 = iota + 0x41
	Card_King_2
)

//花色值
const (
	CARD_COLOR_Fang = iota
	CARD_COLOR_Mei
	CARD_COLOR_Hong
	CARD_COLOR_Hei
	CARD_COLOR_King
	CARD_COLOR_Invalid
)

//牌
type CardBase struct {
	GetCard
	Card int32
	Name string
}

type GetCard interface {
	GetCardColor() int32
	GetCardValue() int32
}

/////////////////////////////////////////////////////////
// 牌值获取函数
func (this *CardBase) GetCardColor() int32 {
	return (this.Card & CARD_COLOR) >> 4
}

func (this *CardBase) GetCardValue() int32 {
	return (this.Card & CARD_VALUE)
}
