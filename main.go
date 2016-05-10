package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

type BankHistRow struct {
	ID   int
	Bank string  `json:bank`
	Tm   int64   `json:tm`
	Sum  float64 `json:sum`
	Rest float64 `json:rest`
	Op   string  `json:op`
	Desc string  `json:desc`
}

type UserBalanceRow struct {
	ID, UID int64
	Tm      int64   `json:tm`
	Sum     float64 `json:sum`
	Bal     float64 `json:bal`
	Acc     string  `json:acc`
	Op      string  `json:op`
}

func totalBuff(bank map[string]float64) float64 {
	total := 0.0
	increment := func(bankName string) float64 {
		if val, ok := bank[bankName]; ok {
			return val
		}
		return 0.0
	}
	total += increment("current")
	total += increment("gold")
	total += increment("coupon")
	return total
}

func floatEquals(a, b float64) bool {
	EPSILON := 0.0001
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}

func checkBankBuff() {
	initTotalBuff := 0.0
	fmt.Printf("chenk in bagining: %.2f\n", initTotalBuff)
	fail := false
	raw, err := readJSONFile("./bhd.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		return
	}
	var bhd []BankHistRow
	err = json.Unmarshal(raw, &bhd)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
	}
	fmt.Printf("rows in file: %d\n", len(bhd))
	bankInitValue := make(map[string]float64)
	for _, bhr := range bhd {
		fail = false
		if _, ok := bankInitValue[bhr.Bank]; !ok {
			bankInitValue[bhr.Bank] = bhr.Rest
		}
		if !floatEquals(bhr.Rest, bankInitValue[bhr.Bank]) {

			tm := time.Unix(bhr.Tm, 0)
			fmt.Printf("log <-> bank (diff): %.2f <-> %.2f (%.2f); bank: %s, id: %d, time: %v, total: %.2f\n", bhr.Rest, bankInitValue[bhr.Bank], bhr.Rest-bankInitValue[bhr.Bank], bhr.Bank, bhr.ID, tm, totalBuff(bankInitValue))
			fail = true
		}
		//create backlog
		switch bhr.Op {
		case "add":
			bankInitValue[bhr.Bank] -= bhr.Sum
		case "remove":
			bankInitValue[bhr.Bank] += bhr.Sum
		case "set":
			bankInitValue[bhr.Bank] = bhr.Sum
		default:
			fmt.Printf("unexpected operation:%+v\n", bhr)
		}
		if !fail {
			fmt.Printf("log: %d - OK, total: %.2f\n", bhr.ID, totalBuff(bankInitValue))
		}
	}
	fmt.Printf("delta: %.2f\n", initTotalBuff-totalBuff(bankInitValue))
	//3646026.19 - 3646992.84 = - 966.65
}

func contains(s []int64, e int64) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

func removeFrom(s []int64, n int) []int64 {
	result := []int64{}
	// Append part before the removed element.
	// ... Three periods (ellipsis) are needed.
	result = append(result, s[0:n]...)
	// Append part after the removed element.
	result = append(result, s[n+1:]...)
	return result
}

func readJSONFile(fileName string) ([]byte, error) {
	file, e := ioutil.ReadFile(fileName)
	if e != nil {
		return nil, e
	}
	return file, nil
}

func checkUserBalance() {

	raw, err := readJSONFile("./ubh.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		return
	}
	var userBalanceHist []UserBalanceRow
	err = json.Unmarshal(raw, &userBalanceHist)
	if err != nil {
		fmt.Printf("Decode error: %v\n", err)
	}
	fmt.Printf("rows in file: %d\n", len(userBalanceHist))
	sum := 0.0
	uBalSet := make(map[string]float64)
	for _, ubh := range userBalanceHist {
		mapKey := strconv.FormatInt(ubh.UID, 10) + "_" + ubh.Acc
		_, ok := uBalSet[mapKey]
		if ok == false {
			fmt.Printf("%d: %.2f\n", ubh.UID, ubh.Bal)
			uBalSet[mapKey] = ubh.Bal
		}
		if !floatEquals(uBalSet[mapKey], ubh.Bal) {
			tm := time.Unix(ubh.Tm, 0)
			fmt.Printf("log <-> user balance (%s) (diff): %.2f <-> %.2f (%.2f); id: %d, uid: %d, time: %v\n", ubh.Acc, uBalSet[mapKey], ubh.Bal, uBalSet[mapKey]-ubh.Bal, ubh.ID, ubh.UID, tm)
			uBalSet[mapKey] = ubh.Bal //fix diffirence
			//fmt.Println("broken balance: ", uBalSet[ubh.UID], ubh)
		}
		switch ubh.Op {
		case "add":
			sum += ubh.Sum
			uBalSet[mapKey] -= ubh.Sum
		case "removal":
			sum -= ubh.Sum
			uBalSet[mapKey] += ubh.Sum
		default:
			fmt.Printf("unexpected operation:%+v\n", ubh)
		}
	}
	fmt.Printf("user balance chenged on: %.2f\n", sum)
}

func main() {
	checkBankBuff()
	checkUserBalance()
}
