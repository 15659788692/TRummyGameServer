package game

import (
	"TeenPatti/TRummyGameServer/Poker"
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
	this.QuickSortLV(&cards1st, 0, len(cards1st)-1)
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
	this.QuickSortCV(&cards1st, 0, len(cards1st)-1)
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
	tCard := append([]GCard{}, cards2st...)
	wildCards := []GCard{}
	for i, v := range tCard {
		if v.GetCardColor() == Poker.CARD_COLOR_King || v.GetCardValue() == WildCard.GetCardValue() {
			wildCards = append(wildCards, v)
			cards2st = append(cards2st[:i], cards2st[i+1:]...)
		}
	}
	//如果除万能牌外不超过一张则成立
	if len(cards2st) <= 1 {
		return true
	}
	//如果花色不同就不是
	for _, v := range cards2st {
		if v.GetCardColor() != cards2st[0].GetCardColor() {
			return false
		}
	}
	//检测同花顺
	wildNum := int32(len(wildCards))
	//按A最大2最小排序
	this.QuickSortLV(&cards2st, 0, len(cards2st)-1)
	isTrue := true
	for k, v := range cards2st {
		if k == 0 {
			continue
		}
		x := cards2st[k-1].GetLogicValue() - v.GetLogicValue()
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
	this.QuickSortCV(&cards2st, 0, len(cards2st)-1)
	for k, v := range cards2st {
		if k == 0 {
			continue
		}
		x := cards2st[k-1].GetCardValue() - v.GetCardValue()
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
	tCard := append([]GCard{}, cardsSet...)
	wildCards := []GCard{}
	for i, v := range tCard {
		if v.GetCardColor() == Poker.CARD_COLOR_King || v.GetCardValue() == WildCard.GetCardValue() {
			wildCards = append(wildCards, v)
			cardsSet = append(cardsSet[:i], cardsSet[i+1:]...)
		}
	}
	//如果除万能牌外不超过一张则成立
	if len(cardsSet) <= 1 {
		return true
	}
	//检测set
	this.QuickSortCLV(&cardsSet, 0, len(cardsSet)-1)
	for k, v := range cardsSet {
		if k == 0 {
			continue
		}
		if v.GetCardValue() != cardsSet[0].GetCardValue() || v.Card == cardsSet[k-1].Card {
			return false
		}
	}
	//
	return true
}

//快速排序（根据逻辑值从大到小）A最大2最小
func (this *GMgrCard) QuickSortLV(cards *[]GCard, begin int, end int) {
	if len(*cards) <= 1 {
		return
	}
	if begin < 0 {
		begin = 0
	}
	if end >= len(*cards) {
		end = len(*cards) - 1
	}
	if begin < end {
		temp := (*cards)[begin]
		i := begin
		j := end
		for i < j {
			for i < j && (*cards)[j].GetLogicValue() < temp.GetLogicValue() {
				j--
			}
			(*cards)[i] = (*cards)[j]
			for i < j && (*cards)[i].GetLogicValue() >= temp.GetLogicValue() {
				i++
			}
			(*cards)[j] = (*cards)[i]
		}
		(*cards)[i] = temp
		this.QuickSortLV(cards, begin, i-1)
		this.QuickSortLV(cards, i+1, end)
	}
}

//快速排序（根据牌值从大到小）A最小K最大
func (this *GMgrCard) QuickSortCV(cards *[]GCard, begin int, end int) {
	if len(*cards) <= 1 {
		return
	}
	if begin < 0 {
		begin = 0
	}
	if end >= len(*cards) {
		end = len(*cards) - 1
	}
	if begin < end {
		temp := (*cards)[begin]
		i := begin
		j := end
		for i < j {
			for i < j && (*cards)[j].GetCardValue() < temp.GetCardValue() {
				j--
			}
			(*cards)[i] = (*cards)[j]
			for i < j && (*cards)[i].GetCardValue() >= temp.GetCardValue() {
				i++
			}
			(*cards)[j] = (*cards)[i]
		}
		(*cards)[i] = temp
		this.QuickSortCV(cards, begin, i-1)
		this.QuickSortCV(cards, i+1, end)
	}
}

//快速排序 （根据花色排序）黑桃A最大 方块2最小
func (this *GMgrCard) QuickSortCLV(cards *[]GCard, begin int, end int) {
	if len(*cards) <= 1 {
		return
	}
	if begin < 0 {
		begin = 0
	}
	if end >= len(*cards) {
		end = len(*cards) - 1
	}
	if begin < end {
		temp := (*cards)[begin]
		i := begin
		j := end
		for i < j {
			for i < j && ((*cards)[j].GetCardColor() < temp.GetCardColor() ||
				((*cards)[j].GetCardColor() == temp.GetCardColor() && (*cards)[j].GetLogicValue() < temp.GetLogicValue())) {
				j--
			}
			(*cards)[i] = (*cards)[j]
			for i < j && ((*cards)[j].GetCardColor() > temp.GetCardColor() ||
				((*cards)[j].GetCardColor() == temp.GetCardColor() && (*cards)[j].GetLogicValue() >= temp.GetLogicValue())) {
				i++
			}
			(*cards)[j] = (*cards)[i]
		}
		(*cards)[i] = temp
		this.QuickSortCV(cards, begin, i-1)
		this.QuickSortCV(cards, i+1, end)
	}
}
