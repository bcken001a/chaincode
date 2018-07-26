cckage main

import (
    "encoding/json"
    "fmt"
    "bytes"
    "strconv"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    sc "github.com/hyperledger/fabric/protos/peer"
)

type Blockchain_bank struct {
}

type Owner struct {
  ObjectType string `json:"docType"`
  Id string `json:"id"`
  Quantity float64 `json:"quantity"`
}

type Bank struct {
  ObjectType string `json:"docType"`
  BankCode string `json:"bankCode"`
  Accounts []Account `json:"accounts"`
}

type Account struct {
  ObjectType string `json:"docType"`
  AccountNumber string `json:"accountNumber"`
  Id string `json:"id"`
  BankCode string `json:"bankCode"`
  Balance float64 `json:"balance"`
}

type Transfer struct {
  ObjectType string `json:"docType"`
  TxId string `json:"txid"`
  FromAccount string `json:"fromAccount"`
  ToAccount string `json:"toAccount"`
  Quantity float64 `json:"quantity"`
  Fee float64  `json:"fee"`
}

func (t *Blockchain_bank) Init(stub shim.ChaincodeStubInterface) sc.Response {
    //初期データを登録する
    owner := Owner{
      ObjectType:"Owner",
      Id:"owner",
      Quantity:0}

    account1 := Account{
        ObjectType:"Account",
        AccountNumber:"0011001",
        Id:"user1001",
        BankCode:"001",
        Balance:100}
    account2 := Account{
        ObjectType:"Account",
        AccountNumber:"0011002",
        Id:"user1002",
        BankCode:"001",
        Balance:500}
    bank :=Bank{
      ObjectType:"Bank",
      BankCode:"001",
      Accounts:[]Account{account1, account2}}

    stub.PutState(owner.Id, OtoB(owner))
    stub.PutState(bank.BankCode, OtoB(bank))
    return shim.Success(nil)
}

func (t *Blockchain_bank) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    functionName, args := stub.GetFunctionAndParameters()
    switch functionName{
    case "createBank":
        return t.createBank(stub, args)
    case "createAccount":
        return t.createAccount(stub, args)
    case "transfer":
        return t.transfer(stub, args)
    case "query":
        return t.query(stub, args)
    default:
        return shim.Error(fmt.Sprintf("Invalid Smart Contract function name. args[0]=> %s", functionName))
    }
}

//参加する銀行を登録する
func (t *Blockchain_bank) createBank(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 1 {
        return shim.Error("Incorrect arguments. <createBank one argument>")
    }
    if len(args[0]) != 3 {
        return shim.Error(fmt.Sprintf("銀行コードは3桁ではない！　bankCode: %s ", args[0]))
    }
    _, errorMessage := newBank(stub, args[0])
    if errorMessage != ""{
      return shim.Error(errorMessage)
    }else{
      return shim.Success(nil)
    }
}

//銀行口座も持つ一般ユーザを登録する(銀行登録しない場合、自動で銀行を登録する)
func (t *Blockchain_bank) createAccount(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 4 {
        return shim.Error("Incorrect arguments. <createAccount four arguments>")
    }
    //stringから、floatへ変更する
    balance, balanceErr := strconv.ParseFloat(args[3], 10)
    if balanceErr != nil {
        return shim.Error(fmt.Sprintf("残高は数字ではない！ balance: %s ", args[3]))
    }
    if len(args[0]) != 7 {
        return shim.Error(fmt.Sprintf("口座番号は７桁ではない！　AccountNumber: %s ", args[0]))
    }
    if len(args[2]) != 3 {
        return shim.Error(fmt.Sprintf("銀行コードは3桁ではない！　AccountNumber: %s ", args[2]))
    }
    if args[0][0:3] != args[2] {
        return shim.Error(fmt.Sprintf("口座番号または銀行コードが間違い！(口座番号 = 銀行コード + 4桁数字)　AccountNumber: %s, bankCode : %s ", args[0], args[2]))
    }
    account := Account{ObjectType:"Account", AccountNumber:args[0], Id:args[1], BankCode:args[2], Balance:balance}
    message:= changeAccount(stub, account, true)
    if message != ""{
      return shim.Error(message)
    }else{
      return shim.Success(nil)
    }
}

func newBank(stub shim.ChaincodeStubInterface, bankCode string)  (Bank, string){
  var bank Bank
  value, err := stub.GetState(bankCode)
  if err != nil {
      return bank, fmt.Sprintf("システム異常! bankCode: %s with error: %s", bankCode, err)
  }
  //銀行コードの存在チェック
  if value != nil {
      return bank, fmt.Sprintf("銀行コードが存在しました！　BankCode: %s", bankCode)
  }
  bank = Bank{ObjectType:"Bank", BankCode:bankCode}
  bank.Accounts =[]Account{}
  stub.PutState(bank.BankCode, OtoB(bank))
  return bank, ""
}

//ユーザ情報を登録する
func changeAccount(stub shim.ChaincodeStubInterface, account Account, isCreate bool) string {
    //銀行コードの存在チェック
    bank, errorMessage := getBank(stub, account.BankCode, isCreate)
    if errorMessage != "" {
        return errorMessage
    }
    for index, savedAccount := range bank.Accounts {
        if savedAccount.AccountNumber == account.AccountNumber {
            if isCreate {
                return fmt.Sprintf("口座番号は存在しました！　AccountNumber : %s", account.AccountNumber)
            }else{
                bank.Accounts[index] = account
            }
        }
    }
    if isCreate{
        bank.Accounts = append(bank.Accounts, account)
    }
    stub.PutState(account.BankCode, OtoB(bank))
    return ""
}

func getBank(stub shim.ChaincodeStubInterface, bankCode string, isCreate bool) (Bank, string) {
    //銀行コードの存在チェック
    var bank Bank
    bytes, accountErr := stub.GetState(bankCode)
    if accountErr != nil {
        return bank, fmt.Sprintf("システム異常. bankCode: %s with error: %s", bankCode, accountErr)
    }
    if bytes == nil {
        if isCreate {
           return newBank(stub, bankCode)
        }else{
          return bank, fmt.Sprintf("指定銀行が存在しない！ BankCode : %s", bankCode)
        }

    }

    if err := BtoO(bytes, &bank); err != nil {
        return bank, fmt.Sprintf("システム異常: %s", err)
    }
    return bank, ""
}

func commitAccount(stub shim.ChaincodeStubInterface, account1 Account, account2 Account) string {
    if account1.BankCode != account2.BankCode{
        var errorMessage string
        errorMessage = changeAccount(stub, account1, false)
        if errorMessage == ""{
            return changeAccount(stub, account2, false)
        }else{
            return errorMessage;
        }
    }else{
        bank, errorMessage := getBank(stub, account1.BankCode, false)
        if errorMessage != "" {
            return errorMessage
        }

        for index, savedAccount := range bank.Accounts {
            if savedAccount.AccountNumber == account1.AccountNumber {
                bank.Accounts[index] = account1
            }
            if savedAccount.AccountNumber == account2.AccountNumber {
                bank.Accounts[index] = account2
            }
        }
        stub.PutState(account1.BankCode, OtoB(bank))
        return ""
    }
}

//ユーザ情報を取得する
func getAccount(stub shim.ChaincodeStubInterface, accountNumber string) (Account, string) {
    var savedAccount Account
    bank, errorMessage := getBank(stub, accountNumber[0:3], false)
    if errorMessage != ""{
        return savedAccount, fmt.Sprintf("指定した口座番号の銀行コードが存在しない! BankCode : %s", accountNumber[0:3])
    }

    for _, savedAccount = range bank.Accounts {
        if savedAccount.AccountNumber == accountNumber {
            return savedAccount, ""
        }
    }

    return savedAccount, fmt.Sprintf("指定した口座番号が存在しない! AccountNumber : %s", accountNumber)
}

//ユーザから別のユーザへ送金する
func (t *Blockchain_bank) transfer(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) != 5 {
        return shim.Error("Incorrect arguments. <transfer five arguments>")
    }
    value, transferErr := stub.GetState("Transfer" + args[0])
    if transferErr != nil {
        return shim.Error(fmt.Sprintf("システム異常！　TxId: %s with error: %s", args[0], transferErr))
    }
    if value != nil {
        return shim.Error(fmt.Sprintf("指定したトランザクションIDが存在してので、送金できない！　TxId: %s", args[0]))
    }
    quantity, quantityErr := strconv.ParseFloat(args[3], 10)
    if quantityErr != nil {
        return shim.Error(fmt.Sprintf("金額が数字ではない！　Quantity: %s ", args[3]))
    }
    fee, feeErr := strconv.ParseFloat(args[4], 10)
    if feeErr != nil {
        return shim.Error(fmt.Sprintf("送金手数料が数字ではない！ Fee: %s", args[4]))
    }
    transfer := Transfer{ObjectType:"Transfer", TxId:args[0], FromAccount:args[1], ToAccount:args[2], Quantity:quantity, Fee:fee}

    on, errw := stub.GetState("owner")
    if errw != nil {
        return shim.Error("システム異常！")
    }
    if on == nil {
        return shim.Error("運営情報がない！")
    }
    var owner Owner
    if err := BtoO(on, &owner); err != nil {
        return shim.Error(fmt.Sprintf("システム異常 : %s", err))
    }

    var fromAccount, toAccount Account
    var errorMessage string
    fromAccount, errorMessage = getAccount(stub, transfer.FromAccount )
    if errorMessage != "" {
        return shim.Error(errorMessage)
    }

    toAccount, errorMessage = getAccount(stub, transfer.ToAccount)
    if errorMessage != "" {
        return shim.Error(errorMessage)
    }

    if fromAccount.Balance < transfer.Quantity * (1 + transfer.Fee / 100) {
      return shim.Error("送金元の残高は不足です！")
    }

    fromAccount.Balance  = fromAccount.Balance - transfer.Quantity * (1 + transfer.Fee / 100)
    toAccount.Balance  = toAccount.Balance + transfer.Quantity
    owner.Quantity = owner.Quantity + transfer.Quantity * transfer.Fee / 100

    errorMessage = commitAccount(stub, fromAccount, toAccount)
    if errorMessage != ""{
      return shim.Error(errorMessage)
    }

    stub.PutState("owner", OtoB(owner))
    stub.PutState("Transfer" + args[0], OtoB(transfer))

    return shim.Success(nil)
}

//特定の情報を取得する
func (t *Blockchain_bank) query(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    if len(args) == 0 {
        return shim.Error("Incorrect arguments. <query one argument>")
    }

    buffer := new(bytes.Buffer)
    if  args[0] == "owner"{
      value, err := stub.GetState(args[0])
      if err != nil {
          return shim.Error(fmt.Sprintf("システム異常: %s with error: %s", args[0], err))
      }
      if value == nil {
          return shim.Error(fmt.Sprintf("運営情報がない！ Id: %s", args[0]))
      }

      var owner Owner
      if err = BtoO(value, &owner); err != nil {
          return shim.Error(fmt.Sprintf("システム異常 : %s", err))
      }
      buffer.WriteString(fmt.Sprintf("ObjectType:%s, Id:%s Quantity:%f",owner.ObjectType, owner.Id, owner.Quantity))
    }else if len(args[0]) == 3 {
      value, err := stub.GetState(args[0])
      if err != nil {
          return shim.Error(fmt.Sprintf("システム異常！ query: %s with error: %s", args[0], err))
      }
      if value == nil {
          return shim.Error(fmt.Sprintf("指定した銀行情報がない！　BankCOde: %s", args[0]))
      }
      var bank Bank
      if err = BtoO(value, &bank); err != nil {
          return shim.Error(fmt.Sprintf("システム異常 : %s", err))
      }
      buffer.WriteString(fmt.Sprintf("ObjectType:%s, BankCode:%s, Accounts:%s", bank.ObjectType, bank.BankCode, string(OtoB(bank.Accounts))))

   }else if len(args[0]) == 7{
      var account Account

      bankBytes, err := stub.GetState(args[0][0:3])
      if err != nil {
          return shim.Error(fmt.Sprintf("システム異常！ AccountNumber: %s with error: %s", args[0], err))
      }
      if bankBytes == nil {
          return shim.Error(fmt.Sprintf("指定したユーザ情報がない！　AccountNumber : %s", args[0]))
      }
      var bank Bank
      if err = BtoO(bankBytes, &bank); err != nil {
          return shim.Error(fmt.Sprintf("システム異常 : %s", err))
      }

      for _, a := range bank.Accounts {
           if a.AccountNumber == args[0] {
             account = a
             buffer.WriteString(fmt.Sprintf("ObjectType:%s, AccountNumber:%s, Id:%s, BankCode:%s, Balance:%f", account.ObjectType, account.AccountNumber, account.Id, account.BankCode, account.Balance))
             break
           }
      }
      if account.Id == "" {
          return shim.Error(fmt.Sprintf("指定したユーザ情報がない！　AccountNumber : %s", args[0]))
      }
  }else{
      return shim.Error("Invalid Keyword!")
  }
  return shim.Success(buffer.Bytes())
}

func BtoO(bytes []byte, v interface{}) error {
    return json.Unmarshal(bytes, &v)
}

func OtoB(v interface{}) []byte {
    value, _ := json.Marshal(v)
    return value
}

func main() {
    err := shim.Start(new(Blockchain_bank));
    if err != nil {
        fmt.Printf("Error starting Blockchain_bank chaincode: %s", err)
    }
}


