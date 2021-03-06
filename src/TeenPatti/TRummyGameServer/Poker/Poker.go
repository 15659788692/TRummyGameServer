package Poker

import (
	"math/rand"
	"strconv"
	"time"
)

/////////////////////////////////////////////////////////
//卡牌管理器，负责做牌
type MgrCard struct {
	MVCard       []CardBase
	MVSourceCard []CardBase
	MSendId      int
}

//含大小王的54张牌
func (this *MgrCard) InitNormalCards() {
	begaincard := []int{Card_Fang_1, Card_Mei_1, Card_Hong_1, Card_Hei_1}
	for _, v := range begaincard {
		for j := 0; j < 13; j++ {
			card := CardBase{Card: int32(v + j)}
			switch card.GetCardColor() {
			case 0:
				card.Name += "方块"
			case 1:
				card.Name += "梅花"
			case 2:
				card.Name += "红桃"
			case 3:
				card.Name += "黑桃"
			}
			card.Name += strconv.Itoa(int(card.GetCardValue()))
			this.MVCard = append(this.MVCard, card)
		}
	}

	// 添加大小王
	this.MVCard = append(this.MVCard, CardBase{Card: Card_King_1, Name: "鬼牌"}, CardBase{Card: Card_King_2, Name: "鬼牌"})
}

//不含大小王的52张牌
func (this *MgrCard) InitNoKingCards() {
	begaincard := []int{Card_Fang_1, Card_Mei_1, Card_Hong_1, Card_Hei_1}
	for _, v := range begaincard {
		for j := 0; j < 13; j++ {
			card := CardBase{Card: int32(v + j)}
			card.GetCardName()
			this.MVCard = append(this.MVCard, card)
		}
	}
}

//洗牌
func (this *MgrCard) Shuffle() {

	this.MSendId = 0
	this.MVSourceCard = append([]CardBase{}, this.MVCard...)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	perm := r.Perm(len(this.MVCard))
	for i, randIndex := range perm {
		this.MVSourceCard[i] = this.MVCard[randIndex]
	}
}

//发手牌
func (this *MgrCard) SendHandCard(num int) []CardBase {
	this.MSendId += num
	return append([]CardBase{}, this.MVSourceCard[this.MSendId-num:this.MSendId]...)
}

//获取剩余牌数，超过返回0
func (this *MgrCard) GetLeftCardCount() int {
	if this.MSendId > len(this.MVSourceCard) {
		return 0
	}
	return len(this.MVSourceCard) - this.MSendId
}

//获取已发牌数
func (this *MgrCard) GetSendCardCount() int {
	return this.MSendId
}
