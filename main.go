package main

import (
	"fmt"
	"html/template"
	"math"
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

const INIT_PRICE = 10
const MONTHLY_PRODUCTION = 1000

// Controls if producers start with one month of stock on month 0
const INIT_WITH_STOCK = true

type BasicMonthTable struct {
	Month         int
	AverageWallet float64
	FoodPrice     float64
	GasPrice      float64
	CoffeePrice   float64
}

type DetailedMonth struct {
	Month     int
	People    []Person
	Producers []Producer
}

type CombinedMonth struct {
	Basic    []BasicMonthTable
	Detailed []DetailedMonth
}

type Person struct {
	IdNumber     int
	WalletAmount float64
	Salary       float64
	//You originally specified these as ints, but I'm assuming you don't want people to be able to buy fractional amounts
	MonthlyFoodIntake int
	MonthlyGasIntake  int
}

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
	Product           string
	Price             float64
	Stock             int
	MonthlyProduction int
}

// Enum equivalent constants
const (
	FoodIdx     int = 0
	GasolineIdx int = 1
	CoffeeIdx   int = 2
)

func (p *Producer) adjustVariables() {
	if p.Stock == 0 {
		p.MonthlyProduction = int(math.Round(float64(p.MonthlyProduction) * 1.1))
		p.Price *= 1.1
	} else {
		p.MonthlyProduction = int(math.Round(float64(p.MonthlyProduction) * 0.9))
		p.Price *= 0.9
	}
}

func (p *Producer) registerPurchase(amount int) float64 {

	if p.Stock > amount {
		p.Stock -= amount
		return float64(amount) * p.Price
	} else {
		amount := float64(p.Stock) * p.Price
		p.Stock = 0
		return amount
	}
}

func (p *Producer) getMaxUnits(money float64) int {
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
		printSimulationState(month, producers, people)

	}
	err := outputSimulationHTML(basicMonths, detailedMonths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Println("Simulation exiting")

}

func simulationStep(producers []Producer, people []Person, month int) ([]Producer, []Person) {
	for i, p := range producers {
		producers[i].adjustVariables()
		if month != 0 {
			producers[i].Stock += p.MonthlyProduction
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

func outputSimulationHTML(months []BasicMonthTable, detailedMonths []DetailedMonth) error {
	template, err := template.ParseFiles("output.tmpl")
	if err != nil {
		return err
	}

	file, err := os.Create("output.html")
	if err != nil {
		return err
	}

	combined := CombinedMonth{Basic: months, Detailed: detailedMonths}

	err = template.Execute(file, combined)
	if err != nil {
		return err
	}

	return nil
}

func fillDetailedMonth(people []Person, producers []Producer, month int) DetailedMonth {
	newPeople := []Person{}
	for _, p := range people {
		newPeople = append(newPeople, Person{MonthlyFoodIntake: p.MonthlyFoodIntake, MonthlyGasIntake: p.MonthlyGasIntake, Salary: roundFloatToPrecision(p.Salary, 2), WalletAmount: roundFloatToPrecision(p.WalletAmount, 2), IdNumber: p.IdNumber})

	}

	newProducers := []Producer{}
	for _, p := range producers {
		newProducers = append(newProducers, Producer{Product: p.Product, Stock: p.Stock, MonthlyProduction: p.MonthlyProduction, Price: roundFloatToPrecision(p.Price, 2)})
	}

	return DetailedMonth{Month: month + 1, People: newPeople, Producers: newProducers}
}

func fillBasicMonth(people []Person, producers []Producer, month int) BasicMonthTable {
	averageWallet := 0.0
	for _, person := range people {
		averageWallet += person.WalletAmount
	}
	averageWallet /= float64(len(people))

	return BasicMonthTable{
		Month:         month + 1,
		AverageWallet: roundFloatToPrecision(averageWallet, 2),
		FoodPrice:     roundFloatToPrecision(producers[FoodIdx].Price, 2),
		GasPrice:      roundFloatToPrecision(producers[GasolineIdx].Price, 2),
		CoffeePrice:   roundFloatToPrecision(producers[CoffeeIdx].Price, 2),
	}
}

func printSimulationState(month int, producers []Producer, people []Person) {
	averageWallet := 0.0
	for _, person := range people {
		averageWallet += person.WalletAmount
	}
	averageWallet /= float64(len(people))

	output := fmt.Sprintf("Month: %v | Average walletAmount: %v | Food price: %v | Coffee price: %v | Gasoline price: %v", month+1, roundFloatToPrecision(averageWallet, 2), roundFloatToPrecision(producers[FoodIdx].Price, 2), roundFloatToPrecision(producers[CoffeeIdx].Price, 2), roundFloatToPrecision(producers[GasolineIdx].Price, 2))

	fmt.Println(output)
}

func initProducer(product string) Producer {
	stock := 0
	if INIT_WITH_STOCK {
		stock = 1000
	}
	return Producer{Product: product, Price: 10, Stock: stock, MonthlyProduction: 1000}
}

func initPerson(r *rand.Rand, ID int) Person {
	return Person{
		IdNumber:          ID,
		Salary:            randFloatInRange(SALARY_MIN, SALARY_MAX, r),
		MonthlyFoodIntake: randIntInRange(FOOD_INTAKE_MIN, FOOD_INTAKE_MAX, r),
		MonthlyGasIntake:  randIntInRange(GAS_INTAKE_MIN, GAS_INTAKE_MAX, r),
	}
}

func randIntInRange(min int, max int, r *rand.Rand) int {
	return r.Intn(max-min) + min

}
func randFloatInRange(min float64, max float64, r *rand.Rand) float64 {
	return min + r.Float64()*(max-min)
}

func roundFloatToPrecision(value float64, digits uint) float64 {
	ratio := math.Pow(10, float64(digits))
	return math.Round(value*ratio) / ratio
}
