package main

import (
	"fmt"
	"html/template"
	"os"
)

type BasicMonthTable struct {
	Month         int
	AverageWallet int
	FoodPrice     int
	GasPrice      int
	CoffeePrice   int
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

// Fills a detailed month struct, which is used for the detailed view in the HTML template
func fillDetailedMonth(people []Person, producers []Producer, month int) DetailedMonth {
	return DetailedMonth{Month: month + 1, People: people, Producers: producers}
}

// Fills a basic month struct, used for the overview in the HTML template
func fillBasicMonth(people []Person, producers []Producer, month int) BasicMonthTable {
	averageWallet := 0
	for _, person := range people {
		averageWallet += person.WalletAmount
	}
	averageWallet /= len(people)

	return BasicMonthTable{
		Month:         month + 1,
		AverageWallet: averageWallet,
		FoodPrice:     producers[FoodIdx].Price,
		GasPrice:      producers[GasolineIdx].Price,
		CoffeePrice:   producers[CoffeeIdx].Price,
	}
}

// Outputs the simulation HTML by creating a file and writing to it according to the template
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

// Prints the simulation state using the basic month table
func printSimulationState(tables []BasicMonthTable) {
	for _, table := range tables {
		output := fmt.Sprintf("Month: %v | Average walletAmount: %v | Food price: %v | Coffee price: %v | Gasoline price: %v", table.Month, table.AverageWallet, table.FoodPrice, table.CoffeePrice, table.GasPrice)
		fmt.Println(output)
	}
}
