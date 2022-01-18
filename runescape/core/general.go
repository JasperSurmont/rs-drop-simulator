package core

import (
	"fmt"
	"math/rand"
	"rs-drop-emulator/runescape/util"
	"sort"
	"strings"
)

const (
	CommonRateWithoutRare float32 = 2.0 / 3.0
	CommonRateWithRare    float32 = 6.0 / 10.0
	UncommonRateWithRare  float32 = 3.0 / 10.0
)

type Drop struct {
	Rate        float32
	Id          string
	Name        string
	AmountRange [2]int
	Amount      int
	Bold        bool // Whether the drop should be put in bold or not
}

func (d *Drop) SetAmount() {
	if d.AmountRange == [2]int{0, 0} {
		d.Amount = 1
	} else {
		diff := d.AmountRange[1] - d.AmountRange[0]
		d.Amount = d.AmountRange[0] + rand.Intn(diff+1) // We do +1 because it works with an open interval
	}
}

// Returns an array of structs with all the drops and their price
func AmountToPrice(drops map[string]*Drop) (res []util.NamedRSPrice, total util.RSPrice, ok bool) {
	ok = true
	total = util.RSPrice("0")

	ch := make(chan util.NamedRSPrice)

	for _, d := range drops {
		go util.GetItemPrice(d.Name, ch)
	}

	// We get same amount of values out of the channel, but continue if there's an error
	for range drops {
		namedPrice := <-ch
		if namedPrice.Error != nil {
			ok = false
			continue
		}

		price := namedPrice.Price
		d := drops[namedPrice.Name]

		err := price.Multiply(d.Amount)
		if err != nil {
			ok = false
			continue
		}
		res = append(res, util.NamedRSPrice{
			Name:  d.Name,
			Price: price,
		})
		total.Add(price)
	}
	return
}

// Sort drops in decreasing value
func SortDrops(m *[]util.NamedRSPrice) {
	sort.Slice((*m)[:], func(i, j int) bool {
		comp, _ := (*m)[i].Price.Compare((*m)[j].Price)
		return comp == 1
	})
}

func AddAlwaysDroptable(amount int64, drops *map[string]*Drop, alwaysDroptable []Drop) {
	for _, d := range alwaysDroptable {
		_, ok := (*drops)[d.Name]
		if ok {
			(*drops)[d.Name].Amount += d.Amount * int(amount)
		} else {
			d.Amount *= int(amount)
			(*drops)[d.Name] = &d
		}
	}
}

// Given the drops with values and drops with prices, make the drop list to be printed
func MakeDropList(n []util.NamedRSPrice, m map[string]*Drop, total util.RSPrice, ok bool) string {
	var sb strings.Builder
	for _, d := range n {
		if m[d.Name].Bold {
			sb.WriteString(fmt.Sprintf("**%v %v:** %v gp\n", m[d.Name].Amount, d.Name, d.Price))
		} else {
			sb.WriteString(fmt.Sprintf("%v %v: %v gp\n", m[d.Name].Amount, d.Name, d.Price))
		}
	}
	sb.WriteString(fmt.Sprintf("\n**Total GE value: %v**", total))
	if !ok {
		sb.WriteString("\nSomething went wrong; not all items were processed correctly")
	}
	return sb.String()
}
