package main

import (
    "encoding/json"
    "fmt"
    "strconv"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/protos/peer"
)

type blockchain_bank struct {
}

type Owner struct {
  ObjectType string `json:"doctype"`
  Id string `json:"id"`
  Quantity float64 `json:"quantity"`
}

type Bank struct {
  ObjectType string `json:"doctype"`
  BankCode string `json:"bank_code"`
  Accounts []Account `json:"accounts"`
}

type Account struct {
  ObjectType string `json:"doctype"`
  AccountNumber string `json:"account_number"`
  Id string `json:"id"`
  BankCode string `json:"bank_code"`
  Balance float64 `json:"balance"`
}

type Transfer struct {
  ObjectType string `json:"doctype"`
  TxId string `json:"transaction_id"`
  FromAccount string `json:"from_account"`
  ToAccount string `json:"to_account"`
  Quantity float64 `json:"quantity"`
  Fee float64 `json:"fee"`
}

func main() {
  err := shim.Start(new(blockchain_bank));
  if err != nil {
    fmt.Printf("Error starting blockchain_bank chaincode: %s", err)
  }
}

func (b *blockchain_bank) Init(APIstub shim.ChaincodeStubInterface) peer.Response {

  var owner = Owner{ObjectType:"Owner",Id:"owner",Quantity:0.0}
  ownerByte, _ := json.Marshal(owner)

  APIstub.PutState("Owner",ownerByte)

  return shim.Success(nil)
}

func (b *blockchain_bank) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

  functionName, args := APIstub.GetFunctionAndParameters()

  switch functionName{
  case "createBank":
    return b.createBank(APIstub, args)
  case "createAccount":
    return b.createAccount(APIstub, args)
  case "transfer":
    return b.transfer(APIstub, args)
  case "query":
    return b.query(APIstub, args)
  }

  return shim.Error("Invalid Smart Contract function name.")

}

func (b *blockchain_bank) query(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
  //  0
  // key

  // 引数が足りているかチェック
  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
  }
  // 空文字チェック
  if args[0] == "" {
    return shim.Error("key is empty.")
  }
  result,_ := APIstub.GetState(args[0])

  return shim.Success(result)

}

func (b *blockchain_bank) createBank(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
  //    0
  // BankCode

  // 引数が足りているかチェック
  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting 1")
  }
  // 空文字チェック
  if args[0] == "" {
    return shim.Error("BankCode is empty.")
  }
  accounts := []Account{}
  var bank = Bank{ObjectType:"Bank", BankCode:args[0], Accounts:accounts}
  bankByte, _ := json.Marshal(bank)

  APIstub.PutState(args[0], bankByte)

  return shim.Success(nil)

}

func (b *blockchain_bank) createAccount(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
  //       0        1     2       3
  // AccountNumber Id BankCode Balance

  // 引数が足りているかチェック
  if len(args) != 4 {
    return shim.Error("Incorrect number of arguments. Expecting 4")
  }
  // 空文字チェック
  if args[0] == "" {
    return shim.Error("AccountNumber is empty.")
  }
  if args[1] == "" {
    return shim.Error("Id is empty.")
  }
  if args[2] == "" {
    return shim.Error("BankCode is empty.")
  }
  if args[3] == "" {
    return shim.Error("Balance is empty.")
  }
  balance, check := AtoF(args[3])
  // 数値チェック
  if !check {
    return shim.Error("Balance is not a number value.")
  }

  bankByte, _ := APIstub.GetState(args[2])
  // 指定された銀行が存在するかチェック
  if bankByte == nil {
    return shim.Error("This bank does not exist.")
  }
  bank := Bank{}

  json.Unmarshal(bankByte, &bank)

  account := Account{ObjectType:"Account", AccountNumber:args[0],
    Id:args[1], BankCode:args[2], Balance:balance}

  bank.Accounts = append(bank.Accounts, account)

  accountByte, _ := json.Marshal(account)
  bankByte, _ = json.Marshal(bank)

  APIstub.PutState(args[0], accountByte)
  APIstub.PutState(args[2], bankByte)

  return shim.Success(nil)

}

func (b *blockchain_bank) transfer(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
  //       0              1           2         3      4
  // TransactionId FromAccountNo ToAccountNo Quantity Fee

  // 引数が足りているかチェック
  if len(args) != 5 {
    return shim.Error("Incorrect number of arguments. Expecting 5")
  }
  // 空文字チェック
  if args[0] == "" {
    return shim.Error("TransactionId is empty.")
  }
  if args[1] == "" {
    return shim.Error("FromAccountNo is empty.")
  }
  if args[2] == "" {
    return shim.Error("ToAccountNo is empty.")
  }
  if args[3] == "" {
    return shim.Error("Quantity is empty.")
  }
  quantity, checkQ := AtoF(args[3])
  // 数値チェック
  if !checkQ {
    return shim.Error("Quantity is not a number value.")
  }
  if args[4] == "" {
    return shim.Error("Fee is empty.")
  }
  fee, checkF := AtoF(args[4])
  // 数値チェック
  if !checkF {
    return shim.Error("Fee is not a number value.")
  }
  fromAccountByte,_ := APIstub.GetState(args[1])
  // 指定された口座(from)が存在するかチェック
  if fromAccountByte == nil {
    return shim.Error("This FromAccount does not exist.")
  }
  toAccountByte,_ := APIstub.GetState(args[1])
  // 指定された口座(to)が存在するかチェック
  if toAccountByte == nil {
    return shim.Error("This ToAccount does not exist.")
  }

  fromAccount := new(Account)
  toAccount := new(Account)
  owner := Owner{}
  ownerByte,_ := APIstub.GetState("Owner")
  json.Unmarshal(fromAccountByte, fromAccount)
  json.Unmarshal(toAccountByte, toAccount)
  json.Unmarshal(ownerByte, &owner)

  fromBalance := fromAccount.Balance - quantity
  fromBalance = fromBalance - fee
  fromAccount.Balance = fromBalance
  toBalance := toAccount.Balance + quantity
  toAccount.Balance = toBalance
  owner.Quantity += fee

  transfer := Transfer{ObjectType:"Transfer", TxId:args[0],
    FromAccount:args[1], ToAccount:args[2], Quantity:quantity, Fee:fee}

  ownerByte, _ = json.Marshal(owner)
  transferByte, _ := json.Marshal(transfer)
  fromAccountByte, _ = json.Marshal(fromAccount)
  toAccountByte, _ = json.Marshal(toAccount)

  APIstub.PutState(args[0], transferByte)
  APIstub.PutState(fromAccount.AccountNumber, fromAccountByte)
  APIstub.PutState(toAccount.AccountNumber, toAccountByte)
  APIstub.PutState("Owner", ownerByte)

  return shim.Success(nil)

}

func AtoF(s string)(float64, bool){
  result,err := strconv.ParseFloat(s, 64)
  if err != nil {
    return result, false
  }
  return result, true
}

