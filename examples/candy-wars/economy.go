package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func (g *Game) CheckEndConditions() {
	day := g.Day.Get()
	debt := g.TotalDebt()
	totalWorth := g.TotalWorth()

	if day > g.MaxDays {
		if totalWorth >= debt {
			g.endGame(fmt.Sprintf(
				"Time's Up - But You Won!\n\nYou made $%d and paid your debt!\n\nCongratulations, Candy Trader!",
				totalWorth,
			), true, totalWorth)
		} else {
			g.endGame(fmt.Sprintf(
				"Game Over!\n\n%d days have passed.\nYou still owe $%d.\n\nThe candy mafia is not pleased...\n\nFinal worth: $%d",
				g.MaxDays, debt-totalWorth, totalWorth,
			), false, totalWorth)
		}
	}

	if debt <= 0 {
		g.endGame(fmt.Sprintf(
			"Congratulations!\n\nYou paid off your debt in %d days!\n\nFinal worth: $%d",
			day, totalWorth,
		), true, totalWorth)
	}
}

func (g *Game) Buy(candyName string, qty int) bool {
	if g.GameOver.Get() {
		return false
	}

	price := g.buyPrice(candyName)
	if price <= 0 {
		return false
	}
	totalCost := price * qty
	available := g.availableStock(candyName)
	if qty > available {
		g.Message.Set(fmt.Sprintf("Only %d in stock.", available))
		return false
	}
	cash := g.Cash.Get()
	if totalCost > cash {
		g.Message.Set("Not enough cash!")
		return false
	}

	inv := g.Inventory.Get()
	currentQty := 0
	for _, q := range inv {
		currentQty += q
	}
	if currentQty+qty > g.Capacity {
		g.Message.Set(fmt.Sprintf("Not enough space! (Capacity: %d/%d)", currentQty, g.Capacity))
		return false
	}

	g.Cash.Set(cash - totalCost)
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[candyName] += qty
	g.Inventory.Set(newInv)
	g.consumeStock(candyName, qty)
	g.consumeTradeBuff()
	g.recordTrade(qty, totalCost)
	g.recordTradeHistory("Buy", candyName, qty, price, totalCost, Locations[g.Location.Get()].Name)

	g.Message.Set(fmt.Sprintf("Bought %d %s for $%d", qty, candyName, totalCost))
	return true
}

func (g *Game) Sell(candyName string, qty int) bool {
	if g.GameOver.Get() {
		return false
	}

	inv := g.Inventory.Get()
	owned := inv[candyName]
	if qty > owned {
		g.Message.Set("You don't have that much!")
		return false
	}

	price := g.sellPrice(candyName)
	if price <= 0 {
		return false
	}
	totalValue := price * qty

	g.Cash.Update(func(c int) int { return c + totalValue })
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[candyName] -= qty
	if newInv[candyName] <= 0 {
		delete(newInv, candyName)
	}
	g.Inventory.Set(newInv)
	g.consumeTradeBuff()
	g.recordTrade(qty, totalValue)
	g.recordTradeHistory("Sell", candyName, qty, price, totalValue, Locations[g.Location.Get()].Name)

	g.Message.Set(fmt.Sprintf("Sold %d %s for $%d", qty, candyName, totalValue))
	return true
}

func (g *Game) PayDebt(amount int) bool {
	cash := g.Cash.Get()
	debt := g.Debt.Get()

	if amount > cash {
		amount = cash
	}
	if amount > debt {
		amount = debt
	}
	if amount <= 0 {
		return false
	}

	g.Cash.Set(cash - amount)
	remainingDebt := debt - amount
	g.Debt.Set(remainingDebt)
	g.Message.Set(fmt.Sprintf("Paid $%d towards debt. Remaining: $%d", amount, remainingDebt))
	g.CheckEndConditions()

	return true
}

func (g *Game) TotalLoanDebt() int {
	total := 0
	for _, loan := range g.Loans {
		total += loan.Balance
	}
	return total
}

func (g *Game) TotalDebt() int {
	return g.Debt.Get() + g.TotalLoanDebt()
}

func (g *Game) TakeLoan(tierIndex int) bool {
	if tierIndex < 0 || tierIndex >= len(LoanTiers) {
		return false
	}
	if !g.loanTierUnlocked(tierIndex) {
		g.Message.Set("That loan tier is still locked.")
		return false
	}
	tier := LoanTiers[tierIndex]
	g.Loans = append(g.Loans, Loan{
		Tier:    tierIndex,
		Balance: tier.Amount,
	})
	g.Cash.Update(func(c int) int { return c + tier.Amount })
	g.Message.Set(fmt.Sprintf("Took a %s loan for $%d.", tier.Name, tier.Amount))
	return true
}

func (g *Game) RepayLoan(amount int) bool {
	if amount <= 0 || len(g.Loans) == 0 {
		return false
	}
	cash := g.Cash.Get()
	if amount > cash {
		amount = cash
	}
	if amount <= 0 {
		return false
	}

	idx := g.highestInterestLoanIndex()
	if idx < 0 {
		return false
	}
	loan := g.Loans[idx]
	payment := amount
	if payment > loan.Balance {
		payment = loan.Balance
	}
	loan.Balance -= payment
	g.Cash.Set(cash - payment)
	g.Message.Set(fmt.Sprintf("Repaid $%d on %s loan.", payment, LoanTiers[loan.Tier].Name))
	g.recordLoanPaid(payment)

	if loan.Balance <= 0 {
		g.Loans = append(g.Loans[:idx], g.Loans[idx+1:]...)
	} else {
		g.Loans[idx] = loan
	}
	g.CheckEndConditions()
	return true
}

func (g *Game) InventoryCount() int {
	inv := g.Inventory.Get()
	count := 0
	for _, q := range inv {
		count += q
	}
	return count
}

func (g *Game) Deposit(amount int) bool {
	if amount <= 0 {
		return false
	}
	cash := g.Cash.Get()
	balance := g.Bank.Get()
	space := g.BankLimit - balance
	if space <= 0 {
		g.Message.Set("Bank is full.")
		return false
	}
	if amount > cash {
		amount = cash
	}
	if amount > space {
		amount = space
	}
	if amount <= 0 {
		return false
	}
	g.Cash.Set(cash - amount)
	g.Bank.Set(balance + amount)
	g.Message.Set(fmt.Sprintf("Deposited $%d.", amount))
	return true
}

func (g *Game) Withdraw(amount int) bool {
	if amount <= 0 {
		return false
	}
	balance := g.Bank.Get()
	if amount > balance {
		amount = balance
	}
	if amount <= 0 {
		return false
	}
	g.Bank.Set(balance - amount)
	g.Cash.Update(func(c int) int { return c + amount })
	g.Message.Set(fmt.Sprintf("Withdrew $%d.", amount))
	return true
}

func (g *Game) loseRandomCandy(count int) int {
	if count <= 0 {
		return 0
	}
	inv := g.Inventory.Get()
	if len(inv) == 0 {
		return 0
	}
	names := make([]string, 0, len(inv))
	for name, qty := range inv {
		if qty > 0 {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return 0
	}

	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}

	lost := 0
	for i := 0; i < count; i++ {
		if len(names) == 0 {
			break
		}
		name := names[rand.Intn(len(names))]
		if newInv[name] > 0 {
			newInv[name]--
			lost++
		}
		if newInv[name] <= 0 {
			delete(newInv, name)
			for idx, candidate := range names {
				if candidate == name {
					names = append(names[:idx], names[idx+1:]...)
					break
				}
			}
		}
	}
	g.Inventory.Set(newInv)
	return lost
}

func (g *Game) TotalWorth() int {
	cash := g.Cash.Get()
	inv := g.Inventory.Get()
	prices := g.Prices.Get()
	totalWorth := cash + g.Bank.Get()
	for name, qty := range inv {
		if price, ok := prices[name]; ok {
			totalWorth += qty * price
		}
	}
	if g.hasStash {
		for name, qty := range g.stash {
			if price, ok := prices[name]; ok {
				totalWorth += qty * price
			}
		}
	}
	return totalWorth
}

func (g *Game) endGame(msg string, win bool, totalWorth int) {
	if g.GameOver.Get() {
		return
	}
	if win {
		g.Wins++
	} else {
		g.Losses++
	}
	if totalWorth > g.BestWorth {
		g.BestWorth = totalWorth
	}
	if g.meta != nil {
		g.meta.Wins = g.Wins
		g.meta.Losses = g.Losses
		g.meta.BestWorth = g.BestWorth
		if win && g.Difficulty() == DifficultyHell {
			g.meta.HellWins++
		}
		g.recordPlaytime(win)
		g.recordRunHistory(win, totalWorth, g.TotalDebt(), g.Day.Get())
		unlocked := g.checkAchievements()
		if len(unlocked) > 0 {
			msg += "\n\nNew achievements: " + strings.Join(unlocked, ", ")
		}
		g.saveMeta()
	}
	g.EventTitle.Set("")
	g.EventMessage.Set("")
	g.ShowEvent.Set(false)
	g.GameOverMsg.Set(msg)
	g.GameOver.Set(true)
}

func (g *Game) reduceLoanDebt(amount int) int {
	if amount <= 0 || len(g.Loans) == 0 {
		return 0
	}
	reduced := 0
	for amount > 0 && len(g.Loans) > 0 {
		idx := g.highestInterestLoanIndex()
		if idx < 0 {
			break
		}
		loan := g.Loans[idx]
		payment := amount
		if payment > loan.Balance {
			payment = loan.Balance
		}
		loan.Balance -= payment
		reduced += payment
		amount -= payment
		if loan.Balance <= 0 {
			g.Loans = append(g.Loans[:idx], g.Loans[idx+1:]...)
		} else {
			g.Loans[idx] = loan
		}
	}
	return reduced
}

func (g *Game) buyPrice(candyName string) int {
	prices := g.Prices.Get()
	price, ok := prices[candyName]
	if !ok {
		return 0
	}
	bonus := g.BuyBonuses[g.Location.Get()]
	if bonus <= 0 {
		bonus = 100
	}
	price = applyPercent(price, bonus)
	buyMult, _ := g.schedulePriceMultipliers()
	price = applyPercent(price, buyMult)
	if g.tradeBuffUses > 0 {
		price = applyPercent(price, g.tradeBuffBuyMult)
	}
	return price
}

func (g *Game) sellPrice(candyName string) int {
	prices := g.Prices.Get()
	price, ok := prices[candyName]
	if !ok {
		return 0
	}
	bonus := g.SellBonuses[g.Location.Get()]
	if bonus <= 0 {
		bonus = 100
	}
	_, sellMult := g.schedulePriceMultipliers()
	price = applyPercent(price, bonus)
	price = applyPercent(price, sellMult)
	if g.tradeBuffUses > 0 {
		price = applyPercent(price, g.tradeBuffSellMult)
	}
	if g.Difficulty() == DifficultyHell && g.difficultySettings().HellBonusesApply {
		rank := g.GetHellRank()
		if rank.Rank == HellRankBronze || rank.Rank == HellRankSilver || rank.Rank == HellRankGold || rank.Rank == HellRankPlatinum || rank.Rank == HellRankDiamond {
			price = applyPercent(price, 105)
		}
	}
	return price
}

func (g *Game) buyPriceAtLocation(loc int, candyName string) int {
	if g.isShortage(loc, candyName) {
		return 0
	}
	base := g.Markets[loc]
	if base == nil {
		base = g.rollBasePrices()
		g.Markets[loc] = base
	}
	price, ok := base[candyName]
	if !ok {
		return 0
	}
	price = applyPercent(price, timeOfDayPriceMultiplier(g.Hour.Get()))
	bonus := g.BuyBonuses[loc]
	if bonus <= 0 {
		bonus = 100
	}
	price = applyPercent(price, bonus)
	buyMult, _ := g.schedulePriceMultipliersAt(loc)
	price = applyPercent(price, buyMult)
	return price
}

func (g *Game) sellPriceAtLocation(loc int, candyName string) int {
	base := g.Markets[loc]
	if base == nil {
		base = g.rollBasePrices()
		g.Markets[loc] = base
	}
	price, ok := base[candyName]
	if !ok {
		return 0
	}
	price = applyPercent(price, timeOfDayPriceMultiplier(g.Hour.Get()))
	bonus := g.SellBonuses[loc]
	if bonus <= 0 {
		bonus = 100
	}
	price = applyPercent(price, bonus)
	_, sellMult := g.schedulePriceMultipliersAt(loc)
	price = applyPercent(price, sellMult)
	if g.Difficulty() == DifficultyHell && g.difficultySettings().HellBonusesApply {
		rank := g.GetHellRank()
		if rank.Rank == HellRankBronze || rank.Rank == HellRankSilver || rank.Rank == HellRankGold || rank.Rank == HellRankPlatinum || rank.Rank == HellRankDiamond {
			price = applyPercent(price, 105)
		}
	}
	return price
}

func (g *Game) highestInterestLoanIndex() int {
	bestIdx := -1
	bestRate := -1
	for i, loan := range g.Loans {
		rate := LoanTiers[loan.Tier].InterestPercent
		if rate > bestRate {
			bestRate = rate
			bestIdx = i
		}
	}
	return bestIdx
}
