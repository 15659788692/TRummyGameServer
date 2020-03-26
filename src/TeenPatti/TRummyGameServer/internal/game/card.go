package game

import (
	"TeenPatti/TRummyGameServer/Poker"
)

//组牌的类型
const (
	Type1stLife     = 7 //第一生命
	Type2stLife     = 6 //第二生命
	TypeNeed1stLife = 5 //没有第一生命
	TypeSequence    = 4 //组合
	TypeSet         = 3 //集
	TypeNeed2stLife = 2 //没有第二生命
	TypeOther       = 1 //杂牌类型
)

type GCard struct {
	Poker.CardBase
}

//获取卡牌逻辑值得大小
func (this *GCard) GetLogicValue() int32 {
	d := this.GetCardValue()
	if this.Card == 0x41 {
		return 16
	}
	if this.Card == 0x42 {
		return 17
	}
	if d <= 1 {
		return d + 13
	}
	return d
}

type GMgrCard struct {
	Poker.MgrCard
}

//两幅牌
func (this *GMgrCard) InitCards() {
	//两副52张牌加两张小王
	this.MVCard = []Poker.CardBase{}
	this.InitNoKingCards()
	this.InitNoKingCards()
	this.MVCard = append(this.MVCard, Poker.CardBase{Card: Poker.Card_King_1}, Poker.CardBase{Card: Poker.Card_King_1})
}

func (this *GMgrCard) SendCard(num int) (sendCards []GCard) {
	if this.MSendId+num > len(this.MVCard) {
		return
	}
	c := this.SendHandCard(num)
	for _, v := range c {
		sendCards = append(sendCards, GCard{CardBase: v})
	}
	return
}

//检测1st Life不算万能牌的
func (this *GMgrCard) Is1stLife(cards1st []GCard) bool {
	//排数小于3就不是
	if len(cards1st) < 3 {
		return false
	}
	//如果有王牌或者花色不同就不是
	for _, v := range cards1st {
		if v.GetCardColor() >= Poker.CARD_COLOR_King || v.GetCardColor() != cards1st[0].GetCardColor() {
			return false
		}
	}
	//按A最大2最小排序
	this.QuickSortLV(cards1st)
	//检测同花顺子
	isTrue := true
	for k, v := range cards1st {
		if k == 0 {
			continue
		}
		if cards1st[k-1].GetLogicValue() != v.GetLogicValue()+1 {
			isTrue = false
			break
		}
	}
	if isTrue {
		return isTrue
	}
	//按A最小K最大排序
	this.QuickSortCV(cards1st)
	for k, v := range cards1st {
		if k == 0 {
			continue
		}
		if cards1st[k-1].GetCardValue() != v.GetCardValue()+1 {
			isTrue = false
			break
		}
		isTrue = true
	}
	return isTrue
}

//检测2st Life可以有万能牌的
func (this *GMgrCard) Is2stLife(cards2st []GCard, WildCard GCard) bool {
	//排数小于3就不是
	if len(cards2st) < 3 {
		return false
	}
	//挑出万能牌
	//如果万能牌是王，那么A也是万能牌
	if WildCard.GetCardColor() == Poker.CARD_COLOR_King {
		WildCard = GCard{Poker.CardBase{Card: Poker.Card_Fang_1}}
	}
	var tCard []GCard
	var wildCards []GCard
	for _, v := range cards2st {
		if v.GetCardColor() == Poker.CARD_COLOR_King || v.GetCardValue() == WildCard.GetCardValue() {
			wildCards = append(wildCards, v)
		} else {
			tCard = append(tCard, v)
		}
	}
	//如果除万能牌外不超过一张则成立
	if len(tCard) <= 1 {
		return true
	}
	//如果花色不同就不是
	for _, v := range tCard {
		if v.GetCardColor() != tCard[0].GetCardColor() {
			return false
		}
	}
	//检测同花顺
	wildNum := int32(len(wildCards))
	//按A最大2最小排序
	this.QuickSortLV(tCard)
	isTrue := true
	for k, v := range tCard {
		if k == 0 {
			continue
		}
		x := tCard[k-1].GetLogicValue() - v.GetLogicValue()
		if x > wildNum {
			isTrue = false
			break
		}
		wildNum -= x
	}
	if isTrue {
		return isTrue
	}
	//按A最小K最大排序
	wildNum = int32(len(wildCards))
	this.QuickSortCV(tCard)
	for k, v := range tCard {
		if k == 0 {
			continue
		}
		x := tCard[k-1].GetCardValue() - v.GetCardValue()
		if x > wildNum {
			isTrue = false
			break
		}
		isTrue = true
		wildNum -= x
	}

	return isTrue
}

//检测set可以有万能牌
func (this *GMgrCard) IsSetLife(cardsSet []GCard, WildCard GCard) bool {
	//排数小于3就不是
	if len(cardsSet) < 3 {
		return false
	}
	//挑出万能牌
	//如果万能牌是王，那么A也是万能牌
	if WildCard.GetCardColor() == Poker.CARD_COLOR_King {
		WildCard = GCard{Poker.CardBase{Card: Poker.Card_Fang_1}}
	}
	var tCard []GCard
	var wildCards []GCard
	for _, v := range cardsSet {
		if v.GetCardColor() == Poker.CARD_COLOR_King || v.GetCardValue() == WildCard.GetCardValue() {
			wildCards = append(wildCards, v)
		} else {
			tCard = append(tCard, v)
		}
	}
	//如果除万能牌外不超过一张则成立
	if len(tCard) <= 1 {
		return true
	}
	//检测set
	this.QuickSortCLV(tCard)
	for k, v := range tCard {
		if k == 0 {
			continue
		}
		if v.GetCardValue() != tCard[0].GetCardValue() || v.Card == cardsSet[k-1].Card {
			return false
		}
	}
	//
	return true
}

//快速排序（根据逻辑值从大到小）A最大2最小
func (this *GMgrCard) QuickSortLV(values []GCard) {
	if len(values) <= 1 {
		return
	}
	mid, i := values[0], 1
	head, tail := 0, len(values)-1
	for head < tail {
		if values[i].GetLogicValue() < mid.GetLogicValue() ||
			(values[i].GetCardColor() < mid.GetCardColor() && values[i].GetLogicValue() == mid.GetLogicValue()) {
			values[i], values[tail] = values[tail], values[i]
			tail--
		} else {
			values[i], values[head] = values[head], values[i]
			head++
			i++
		}
	}
	values[head] = mid
	this.QuickSortLV(values[:head])
	this.QuickSortLV(values[head+1:])
}

//快速排序（根据牌值从大到小）A最小K最大
func (this *GMgrCard) QuickSortCV(values []GCard) {
	if len(values) <= 1 {
		return
	}
	mid, i := values[0], 1
	head, tail := 0, len(values)-1
	for head < tail {
		if values[i].GetCardValue() < mid.GetCardValue() ||
			(values[i].GetCardColor() < mid.GetCardColor() && values[i].GetCardValue() == mid.GetCardValue()) {
			values[i], values[tail] = values[tail], values[i]
			tail--
		} else {
			values[i], values[head] = values[head], values[i]
			head++
			i++
		}
	}
	values[head] = mid
	this.QuickSortCV(values[:head])
	this.QuickSortCV(values[head+1:])
}

//快速排序 （根据花色排序）黑桃A最大 方块2最小
func (this *GMgrCard) QuickSortCLV(values []GCard) {
	if len(values) <= 1 {
		return
	}
	mid, i := values[0], 1
	head, tail := 0, len(values)-1
	for head < tail {
		if values[i].GetCardColor() > mid.GetCardColor() ||
			(values[i].GetCardColor() == mid.GetCardColor() && values[i].GetLogicValue() > mid.GetLogicValue()) {
			values[i], values[tail] = values[tail], values[i]
			tail--
		} else {
			values[i], values[head] = values[head], values[i]
			head++
			i++
		}
	}
	values[head] = mid
	this.QuickSortCLV(values[:head])
	this.QuickSortCLV(values[head+1:])
}

//检测是否胡牌
func (this *GMgrCard) CheckoutHu(cards [][]GCard, WildCard GCard) (bool, int64) {
	tCards := append([][]GCard{}, cards...)
	//检测1st life
	Is1stlife := false
	for k, v := range tCards {
		if this.Is1stLife(v) {
			tCards = append(tCards[:k], tCards[k+1:]...)
			Is1stlife = true
			break
		}
	}
	if !Is1stlife {
		return false, this.ComputePoint(tCards, WildCard)
	}
	//检测2st life
	Is2stlife := false
	for k, v := range tCards {
		if this.Is2stLife(v, WildCard) {
			tCards = append(tCards[:k], tCards[k+1:]...)
			Is2stlife = true
			break
		}
	}
	if !Is2stlife {
		return false, this.ComputePoint(tCards, WildCard)
	}
	//检测其他牌组是否成型
	var falsecard [][]GCard
	for _, v := range tCards {
		if !this.Is2stLife(v, WildCard) && !this.IsSetLife(v, WildCard) {
			falsecard = append(falsecard, v)

		}
	}
	if len(falsecard) > 0 {
		return false, this.ComputePoint(falsecard, WildCard)
	}
	return true, 0
}

//计算点数
func (this *GMgrCard) ComputePoint(cards [][]GCard, wildCard GCard) int64 {
	//挑出万能牌
	//如果万能牌是王，那么A也是万能牌
	if wildCard.GetCardColor() == Poker.CARD_COLOR_King {
		wildCard = GCard{Poker.CardBase{Card: Poker.Card_Fang_1}}
	}
	var wildCards []GCard //万能牌
	var tCard []GCard     //杂牌
	for _, v := range cards {
		for _, v1 := range v {
			if v1.GetCardColor() == Poker.CARD_COLOR_King || v1.GetCardValue() == wildCard.GetCardValue() {
				wildCards = append(wildCards, v1)
			} else {
				tCard = append(tCard, v1)
			}
		}
	}
	var coins int64
	for _, v := range tCard {
		c := int64(v.GetLogicValue())
		if c > 10 {
			c = 10
		}
		coins += c
	}
	return coins
}
