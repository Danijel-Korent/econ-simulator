package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const MAX_MONTHS = 100

// 0 based, payout happens when month = 49 instead of 50
const PAYOUT_MONTH = 49
const NUM_PEOPLE = 20

const SALARY_MIN = 1000
const SALARY_MAX = 10000
const FOOD_INTAKE_MIN = 30
const FOOD_INTAKE_MAX = 60
const GAS_INTAKE_MIN = 100
const GAS_INTAKE_MAX = 200

const FOOD_PRODUCTION_COST = 5
const GAS_PRODUCTION_COST = 5
const COFFEE_PRODUCTION_COST = 5

const INIT_PRICE = 10
const MONTHLY_PRODUCTION = 1000

// Controls if producers start with one month of stock on month 0
const INIT_WITH_STOCK = true

type Person struct {
	IdNumber     int
	WalletAmount int
	Salary       int
	//You originally specified these as ints, but I'm assuming you don't want people to be able to buy fractional amounts
	MonthlyFoodIntake int
	MonthlyGasIntake  int
}

// Simulates the purchase of goods, adjusting variables on the person and alerting the producer
func (p *Person) buyGoods(producers []Producer) {
	foodCost := producers[FoodIdx].registerPurchase(p.MonthlyFoodIntake)
	gasCost := producers[GasolineIdx].registerPurchase(p.MonthlyGasIntake)
	p.WalletAmount -= (foodCost + gasCost)
	if p.WalletAmount > foodCost {
		p.WalletAmount -= producers[FoodIdx].registerPurchase(p.MonthlyFoodIntake)
	}

	maxCoffee := producers[CoffeeIdx].getMaxUnits(p.WalletAmount)
	p.WalletAmount -= producers[CoffeeIdx].registerPurchase(maxCoffee)
}

type Producer struct {
	BankBalance       int
	Product           string
	Price             int
	Stock             int
	MonthlyProduction int
	ProductionCost    int
}

// Enum equivalent constants
const (
	FoodIdx     int = 0
	GasolineIdx int = 1
	CoffeeIdx   int = 2
)

// Adjusts the price of goods based on the stock
func (p *Producer) adjustVariables() {
	newPrice := 0.0
	if p.Stock == 0 {
		newPrice = float64(p.Price) * 1.1
	} else {
		newPrice = float64(p.Price) * 1.1
	}
	p.Price = int(newPrice + 0.5)
}

// Adds as much product to the producer as they have money to make
func (p *Producer) produceProducts() {
	amount := int(p.BankBalance / p.ProductionCost)
	p.MonthlyProduction = amount
	p.BankBalance = amount * p.ProductionCost
}

// Returns the amount purchased and subtracts it from the stock, adding to the bank balance.
func (p *Producer) registerPurchase(amount int) int {

	if p.Stock > amount {
		p.Stock -= amount
		p.BankBalance += amount * p.Price
		return amount * p.Price
	} else {
		amount := p.Stock * p.Price
		p.Stock = 0
		return amount
	}
}

// Returns the maximum number of units one can buy with a certain amount of money
func (p *Producer) getMaxUnits(money int) int {
	//Cast instead of rounding to truncate (prevents overspends)
	amount := int(money / p.Price)
	if amount > p.Stock {
		return p.Stock
	}
	return amount
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	producers := make([]Producer, 3)
	producers[FoodIdx] = initProducer("food")
	producers[GasolineIdx] = initProducer("gasoline")
	producers[CoffeeIdx] = initProducer("coffee")

	people := []Person{}
	for i := 0; i < NUM_PEOPLE; i++ {
		people = append(people, initPerson(r, i))
	}

	detailedMonths := []DetailedMonth{}
	basicMonths := []BasicMonthTable{}
	for month := 0; month < MAX_MONTHS; month++ {
		producers, people = simulationStep(producers, people, month)
		detailedMonths = append(detailedMonths, fillDetailedMonth(people, producers, month))
		basicMonths = append(basicMonths, fillBasicMonth(people, producers, month))
	}

	printSimulationState(basicMonths)
	err := outputSimulationHTML(basicMonths, detailedMonths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Println("Simulation exiting")

}

// Steps through one month of the simulation, adjusting variables as needed
func simulationStep(producers []Producer, people []Person, month int) ([]Producer, []Person) {
	for i, _ := range producers {
		producers[i].adjustVariables()
		if month != 0 {
			producers[i].produceProducts()
		}
	}

	for i, p := range people {
		people[i].WalletAmount += p.Salary
		if month == PAYOUT_MONTH {
			people[i].WalletAmount *= 2
		}
		people[i].buyGoods(producers)
	}

	return producers, people

}

// Initialises a new producer and returns it
func initProducer(product string) Producer {
	stock := 0
	if INIT_WITH_STOCK {
		stock = 1000
	}
	productionCost := 0
	switch product {
	case "food":
		productionCost = FOOD_PRODUCTION_COST
		break
	case "gasoline":
		productionCost = GAS_PRODUCTION_COST
		break
	case "coffee":
		productionCost = COFFEE_PRODUCTION_COST
		break
	}

	return Producer{BankBalance: 0, ProductionCost: productionCost, Product: product, Price: 10, Stock: stock}
}

// Creates a new person, generating random variables
func initPerson(r *rand.Rand, ID int) Person {
	return Person{
		IdNumber:          ID,
		Salary:            randIntInRange(SALARY_MIN, SALARY_MAX, r),
		MonthlyFoodIntake: randIntInRange(FOOD_INTAKE_MIN, FOOD_INTAKE_MAX, r),
		MonthlyGasIntake:  randIntInRange(GAS_INTAKE_MIN, GAS_INTAKE_MAX, r),
	}
}

// Generates a random integer between min and max
func randIntInRange(min int, max int, r *rand.Rand) int {
	return r.Intn(max-min) + min

}
