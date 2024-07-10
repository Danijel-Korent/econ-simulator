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

const FOOD_PRODUCTION_COST = 50
const GAS_PRODUCTION_COST = 50
const COFFEE_PRODUCTION_COST = 50

const INIT_MAX_PRODUCTION = 1000
const INIT_PRODUCER_BALANCE = 1000
const INIT_PRICE = 10

// Controls if producers start with one month of stock on month 0
const INIT_WITH_STOCK = true

type Person struct {
	IdNumber     int
	Employer     int
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

func (p *Person) receiveSalary(producers []Producer) {
	p.Salary = producers[p.Employer].MonthSalary
	p.WalletAmount += p.Salary
}

func (p *Person) checkNewJobs(producers []Producer) {
	for i, producer := range producers {
		if float64(producer.MonthSalary)/float64(p.Salary) >= 1.5 && i != p.Employer {
			if producers[i].addEmployee(p) {
				producers[p.Employer].removeEmployee(p)
				p.Employer = i
				return
			}
		}
	}
}

type Producer struct {
	BankBalance       int
	Product           string
	MonthSalary       int
	MonthHires        int
	Employees         []*Person
	NumEmployees      int
	Price             int
	Stock             int
	MonthlyProduction int
	MaximumProduction int
	ProductionCost    int
}

// Enum equivalent constants
const (
	FoodIdx     int = 0
	GasolineIdx int = 1
	CoffeeIdx   int = 2
)

// Adjusts the price and maximum production of goods based on the stock
func (p *Producer) adjustPriceAndProduction() {
	newPrice := 0.0
	if p.Stock == 0 {
		newPrice = float64(p.Price) * 1.1
		p.MaximumProduction = int(float64(p.MaximumProduction)*1.1 + 0.5)
	} else {
		newPrice = float64(p.Price) * 0.9
		p.MaximumProduction = int(float64(p.MaximumProduction)*0.9 + 0.5)
	}
	p.Price = int(newPrice + 0.5)
}

// Works out the monthly salary based on the remaining money from production
func (p *Producer) calculateSalary() {
	if len(p.Employees) > 0 {
		p.MonthSalary = int(float64(p.BankBalance) / float64(len(p.Employees)))
		return
	}

}

// Adds as much product to the producer as they have money to make
func (p *Producer) produceProducts() {
	amount := int(float64(p.BankBalance) / float64(p.ProductionCost))
	if amount > p.MaximumProduction {
		amount = p.MaximumProduction
	}
	p.MonthlyProduction = amount
	p.BankBalance -= amount * p.ProductionCost
}

// Removes the given employee from the producer
func (p *Producer) removeEmployee(person *Person) {
	for i := range p.Employees {
		if p.Employees[i].IdNumber == person.IdNumber {
			p.Employees = append(p.Employees[:i], p.Employees[i+1:]...)
			p.NumEmployees -= 1
			return
		}
	}
}

// Checks if a new employee can be hired. Employs them and returns true if so, returns false otherwise.
func (p *Producer) addEmployee(person *Person) bool {
	if p.MonthHires < 2 {
		p.Employees = append(p.Employees, person)
		p.NumEmployees += 1
		return true
	}

	return false
}

// Subtracts from stock, adding to bank balance. Returns the cost of purchase.
func (p *Producer) registerPurchase(amount int) int {
	if p.Stock >= amount {
		p.Stock -= amount
		p.BankBalance += amount * p.Price
		return amount * p.Price
	} else {
		price := p.Stock * p.Price
		p.Stock = 0
		return price
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
		person := initPerson(r, i)

		producers[person.Employer].NumEmployees += 1
		producers[person.Employer].Employees = append(producers[person.Employer].Employees, &person)
		people = append(people, person)

	}

	detailedMonths := []DetailedMonth{}
	basicMonths := []BasicMonthTable{}
	for month := 0; month < MAX_MONTHS; month++ {
		producers, people = simulationStep(producers, people, month)
		detailedMonths = append(detailedMonths, fillDetailedMonth(people, producers, month))
		basicMonths = append(basicMonths, fillBasicMonth(people, producers, month))

		printProducer(detailedMonths[month].Producers[0])

		if month == MAX_MONTHS-1 {
			printSimulationState(basicMonths)
			err := outputSimulationHTML(basicMonths, detailedMonths)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	fmt.Println("Simulation exiting")
}

// Steps through one month of the simulation, adjusting variables as needed
func simulationStep(producers []Producer, people []Person, month int) ([]Producer, []Person) {
	for i := range producers {
		producers[i].adjustPriceAndProduction()
		producers[i].produceProducts()
		producers[i].calculateSalary()
	}

	for i := range people {
		people[i].checkNewJobs(producers)
		people[i].receiveSalary(producers)
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

	return Producer{BankBalance: INIT_PRODUCER_BALANCE, ProductionCost: productionCost, Product: product, Price: 10, Stock: stock, MaximumProduction: INIT_MAX_PRODUCTION, Employees: []*Person{}, NumEmployees: 0}
}

// Creates a new person, generating random variables. Returns the person and the producer they are employed by
func initPerson(r *rand.Rand, ID int) Person {
	randomEmployer := randIntInRange(0, 3, r)

	return Person{
		IdNumber:          ID,
		Employer:          randomEmployer,
		WalletAmount:      0,
		Salary:            0,
		MonthlyFoodIntake: randIntInRange(FOOD_INTAKE_MIN, FOOD_INTAKE_MAX, r),
		MonthlyGasIntake:  randIntInRange(GAS_INTAKE_MIN, GAS_INTAKE_MAX, r),
	}
}

// Generates a random integer between min and max
func randIntInRange(min int, max int, r *rand.Rand) int {
	return r.Intn(max-min) + min
}

func calculateTotalMoneyInSimulation(people []Person, producers []Producer) int {
	total := 0
	for _, p := range people {
		total += p.WalletAmount
	}
	for _, p := range producers {
		total += p.BankBalance
		total += p.Stock * p.ProductionCost
	}

	return total
}
