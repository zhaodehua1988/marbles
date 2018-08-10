/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"fmt"
	"strconv"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

const (
	StepNum = 8
)
//申请所处的各个阶段
const (
	New = iota     //供应商新建marble并提交申请   0
	CompanyCheck   //核心企业审核  1
	BankCheck      //银行审核     2
	SuppRecv       //供应商收款   3
	CompanyRePayMent //核心企业还款  4
	SuppRepayment  //供应商还款      5
	BankRecv       //银行确认收款    6
	EndOf            //包括成功和失败两种情况 7
)
//确认阶段
const(
	Disable = iota //0
	Wait
	Success
	Failure
)
//var Step_Company[StepNum]string

//{                    "enrollId": "core-enterprise",                    "enrollSecret": "cepw"                },
//{                    "enrollId": "bank",                    "enrollSecret": "bankpw"                },
//{                    "enrollId": "auditor",                    "enrollSecret": "auditor"                }
//                                     0                 1          2       3              4            5          6      7
var Step_company  =[StepNum]string {"supplier","core-enterprise","bank","supplier","core-enterprise","supplier","bank","bank"}
// ============================================================================================================================
// Asset Definitions - The ledger will store marbles and owners
// ============================================================================================================================
//
// ----- Marbles ----- //
type Marble struct {
	ObjectType string             `json:"docType"`  //field for couchdb
	Id         string             `json:"id"`       //the fieldtags are needed to keep case from bouncing around
	Contact    string             `json:"contact"` //contract num
	Balance    int                `json:"balance"`  //the balance of contract
	Title      string             `json:"title"`
	User       UserRelation       `json:"user"`  //User
	Check      [StepNum]CheckInfo `json:"check"` //申请审核进度 0生成 1供应商 2 核心企业 3 银行 4 银行放款 5供应商收款 6供应商还款  7完成
}

// ----- User ----- //               User
type User struct {
	ObjectType string `json:"docType"`     //field for couchdb
	Id         string `json:"id"`
	Username   string `json:"username"`
	Company    string `json:"company"`
	Enabled    bool   `json:"enabled"`     //disabled owners will not be visible to the application
}
// ----- Owners ----- //
type UserRelation struct {
	Id         string `json:"id"`
	Username   string `json:"username"`    //this is mostly cosmetic/handy, the real relation is by UserID not Username
	Company    string `json:"company"`     //this is mostly cosmetic/handy, the real relation is by UserID not Company
}

type CheckInfo struct{
	UserID  string `json:"userid"`    //id
	Company    string `json:"company"`      //name
	Date    string `json:"date"`      //操作的日期
	Review  int    `json:"review"`    //确认阶段{ 0:不需要确认 1:待确认 2:成功 3:失败 }
	Comment string `json:"comment"`   //备注
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}


// ============================================================================================================================
// Init - initialize the chaincode 
//
// Marbles does not require initialization, so let's run a simple test instead.
//
// Shows off PutState() and how to pass an input argument to chaincode.
// Shows off GetFunctionAndParameters() and GetStringArgs()
// Shows off GetTxID() to get the transaction ID of the proposal
//
// Inputs - Array of strings
//  ["314"]
// 
// Returns - shim.Success or error
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Marbles Is Starting Up")
	funcName, args := stub.GetFunctionAndParameters()
	var number int
	var err error
	txId := stub.GetTxID()
	
	fmt.Println("Init() is running")
	fmt.Println("Transaction ID:", txId)
	fmt.Println("  GetFunctionAndParameters() function:", funcName)
	fmt.Println("  GetFunctionAndParameters() args count:", len(args))
	fmt.Println("  GetFunctionAndParameters() args found:", args)

	// expecting 1 arg for instantiate or upgrade
	if len(args) == 1 {
		fmt.Println("  GetFunctionAndParameters() arg[0] length", len(args[0]))

		// expecting arg[0] to be length 0 for upgrade
		if len(args[0]) == 0 {
			fmt.Println("  Uh oh, args[0] is empty...")
		} else {
			fmt.Println("  Great news everyone, args[0] is not empty")

			// convert numeric string to integer
			number, err = strconv.Atoi(args[0])
			if err != nil {
				return shim.Error("Expecting a numeric string argument to Init() for instantiate")
			}

			// this is a very simple test. let's write to the ledger and error out on any errors
			// it's handy to read this right away to verify network is healthy if it wrote the correct value
			err = stub.PutState("selftest", []byte(strconv.Itoa(number)))
			if err != nil {
				return shim.Error(err.Error())                  //self-test fail
			}
		}
	}

	// showing the alternative argument shim function
	alt := stub.GetStringArgs()
	fmt.Println("  GetStringArgs() args count:", len(alt))
	fmt.Println("  GetStringArgs() args found:", alt)

	// store compatible marbles application version
	err = stub.PutState("marbles_ui", []byte("4.0.1"))
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("Ready for action")                          //self-test pass
	return shim.Success(nil)
}


// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	if function == "init" {                    //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "read" {             //generic read ledger
		return read(stub, args)
	} else if function == "write" {            //generic writes to ledger
		return write(stub, args)
	} else if function == "delete_marble" {    //deletes a marble from state
		return delete_marble(stub, args)
	} else if function == "init_marble" {      //create a new marble
		return init_marble(stub, args)
	}else if function == "init_owner"{        //create a new marble owner
		return init_owner(stub, args)
	} else if function == "read_everything"{   //read everything, (owners + marbles + companies)
		return read_everything(stub,args)
	} else if function == "getHistory"{        //read history of a marble (audit)
		return getHistory(stub, args)
	} else if function == "getMarblesByRange"{ //read a bunch of marbles by start and stop id
		return getMarblesByRange(stub, args)
	} else if function == "disable_owner"{     //disable a marble owner from appearing on the UI
		return disable_owner(stub, args)
	} else if function == "review_marble"{
		return review_marble(stub,args)        //对marble的审核，或者放款，还款等操作
	}else if function == "read_allmarble"{
		return read_allmarble(stub,args)       //通过userID查询所有的相关的marble
	}else if function == "read_allstate"{      //
		return read_allstate(stub,args)
	}else if function == "tx_marble"{
		return tx_marble(stub,args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}


// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}
