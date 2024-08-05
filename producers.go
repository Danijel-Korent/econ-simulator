package main

import (
	"fmt"
	"math/rand"
)

type ProductionCost struct {
	ProducerName string
	PerUnits     int
	Amount       int
}

type Producer struct {
	BankBalance int
	Product     string
	MonthSalary int
	MonthHires  int
	Employees   []*Person
	Price       int
	Stock       int
	//Number of units since the producer last bought necessary materials (gas and coffee)
	UnpaidUnits       int
	MonthlyProduction int
	ProductionLimit   int
	PosX              int
	PosY              int

	UnitsSold int

	MaxHires               int
	ProductionChangeAmount float64
	PriceChangeAmount      float64

	ProductionCosts []ProductionCost
}

// Initialises a new producer and returns it
func initProducer(config SimConfig, pConfig ProducerConfig, r *rand.Rand) Producer {
	return Producer{BankBalance: pConfig.InitBalance, Product: pConfig.ProductName, Price: pConfig.InitPrice, MaxHires: pConfig.MaxHires, Stock: pConfig.InitStock, MonthSalary: pConfig.InitSalary, Employees: []*Person{}, MonthlyProduction: pConfig.InitMonthlyProduction, PosX: randIntInRange(config.PositionMin, config.PositionMax, r), PosY: randIntInRange(config.PositionMin, config.PositionMax, r), ProductionChangeAmount: pConfig.ProductionChangeAmount, PriceChangeAmount: pConfig.PriceChangeAmount, ProductionCosts: pConfig.ProductionCosts, UnpaidUnits: 0, ProductionLimit: pConfig.ProductionLimit}
}

func (p *Producer) simulationStep(producers []Producer) {
	p.UnitsSold = 0
	p.MonthHires = 0
	p.adjustVariables()
	p.produceProducts()
	p.payProductionCost(producers)
	p.payEmployees()
}

func (p *Producer) setBankBalance(amount int) {
	if amount < 0 {
		panic(fmt.Sprintf("Attempted to set bank balance of producer %v to %v", p.Product, amount))
	}

	p.BankBalance = amount
}

// Adds as much product to the producer as they have money to make
func (p *Producer) produceProducts() {
	p.Stock += p.MonthlyProduction
	p.UnpaidUnits += p.MonthlyProduction
}

func (p *Producer) payProductionCost(producers []Producer) {
	for _, cost := range p.ProductionCosts {
		desiredPurchases := int(float64(p.UnpaidUnits) / float64(cost.PerUnits))
		producerIdx := findProducerIdx(cost.ProducerName, producers)
		purchasableUnits := producers[producerIdx].getMaxUnits(p.BankBalance)
		units := desiredPurchases
		if desiredPurchases > purchasableUnits {
			units = purchasableUnits
		}

		purchaseCost := producers[producerIdx].registerPurchase(units)
		p.setBankBalance(p.BankBalance - purchaseCost)
		p.UnpaidUnits -= units * cost.PerUnits
	}
}

// Removes the given employee from the producer
func (p *Producer) removeEmployee(person *Person) {
	for i := range p.Employees {
		if p.Employees[i].IdNumber == person.IdNumber {
			p.Employees = append(p.Employees[:i], p.Employees[i+1:]...)
			return
		}
	}
}

func (p *Producer) payEmployees() {
	if len(p.Employees) > 0 {
		p.MonthSalary = p.BankBalance / len(p.Employees)
	} else {
		p.MonthSalary = p.BankBalance
	}

	for i := range p.Employees {
		p.Employees[i].setWalletAmount(p.Employees[i].WalletAmount + p.MonthSalary)
		p.Employees[i].Salary = p.MonthSalary
		p.setBankBalance(p.BankBalance - p.MonthSalary)
	}
}

// Checks if a new employee can be hired. Employs them and returns true if so, returns false otherwise.
func (p *Producer) addEmployee(person *Person) bool {
	if p.MonthHires < p.MaxHires {
		p.Employees = append(p.Employees, person)
		p.MonthHires += 1
		return true
	}

	return false
}

// Subtracts from stock, adding to bank balance. Returns the cost of purchase.
func (p *Producer) registerPurchase(amount int) int {
	if p.Stock >= amount {
		p.Stock -= amount
		p.UnitsSold += amount
		newBalance := p.BankBalance + (amount * p.Price)
		p.setBankBalance(newBalance)
		return amount * p.Price
	} else {
		price := p.Stock * p.Price
		p.UnitsSold += p.Stock
		p.Stock = 0
		p.setBankBalance(p.BankBalance + price)
		return price
	}
}

// Returns the maximum number of units one can buy with a certain amount of money
func (p *Producer) getMaxUnits(money int) int {
	if money < 0 {
		return 0
	}
	//Cast instead of rounding to truncate (prevents overspends)
	amount := int(float64(money) / float64(p.Price))
	if amount > p.Stock {
		return p.Stock
	}
	return amount
}

// Adjusts the price and salary of employees based on the stock
func (p *Producer) adjustVariables() {
	newPrice := 0.0
	newProduction := 0.0
	if p.Stock == 0 {
		newProduction = float64(p.MonthlyProduction) * (1.0 + p.ProductionChangeAmount)
		newPrice = float64(p.Price) * (1.0 + p.PriceChangeAmount)
	} else {
		newProduction = float64(p.MonthlyProduction) * (1.0 - p.ProductionChangeAmount)
		newPrice = float64(p.Price) * (1.0 - p.PriceChangeAmount)
	}

	if newProduction > float64(p.ProductionLimit) {
		newProduction = float64(p.ProductionLimit)
	}

	p.MonthlyProduction = int(newProduction + 0.5)
	p.Price = int(newPrice + 0.5)
}

// Looks through the array of producers and finds one with the matching product name and returns its index in the array
func findProducerIdx(name string, producers []Producer) int {
	for i, p := range producers {
		if p.Product == name {
			return i
		}
	}
	return -1
}
