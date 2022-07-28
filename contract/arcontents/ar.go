package main

// import

import (
	"encoding/json"
	"fmt"
	"time"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)
// 체인코드 구조체

type SimpleChaincode struct {
	contractapi.Contract
}

// WS ARContents 구조체

type ARContents struct {
	ObjectType string `json:"docType"`
	PID string `json:"pid"`		// Plug-in ID
	Owner string `json:"owner"`		// 소유자
	Price int `json:"price"`			// 판매 가격
	Status string `'json:"status"`	// Plug-in 상태 : 판매중, 판매완료, 등록승인, 등록거절
}

type HistoryQueryResult struct {
	Record    *ARContents    `json:"record"`
	TxId     string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// InitARC 함수

func (t *SimpleChaincode) InitARContents(ctx contractapi.TransactionContextInterface, pid string, owner string, price int, status string) error {
	fmt.Println("- start init ARContents")
	
	// 기등록 마블 검색

	ARContentsAsBytes, err := ctx.GetStub().GetState(pid)
	if err != nil {
		return fmt.Errorf("Failed to get AR Contents: " + err.Error())
	} else if ARContentsAsBytes != nil {
		return fmt.Errorf("This AR Contents already exists: " + pid)
	}

	// 구조체 생선 후 마샬링하고 PutState 처리
	ar_contents := &ARContents{"ar_contents", pid, owner, price, status}
	ARContentsJSONasBytes, err := json.Marshal(ar_contents)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(pid, ARContentsJSONasBytes)
	if err != nil {
		return err
	}
	return nil
}
// ReadMarble 함수
func (t *SimpleChaincode) ReadARContents(ctx contractapi.TransactionContextInterface, pid string) (*ARContents, error) {
	fmt.Println("- start read AR Contents")

	// 기등록 마블 검색
	ARContentsAsBytes, err := ctx.GetStub().GetState(pid)
	if err != nil {
		return nil, fmt.Errorf("Failed to get AR Contents: " + err.Error())
	}
	if ARContentsAsBytes == nil {
		return nil, fmt.Errorf("the asset does not exists: " + pid)
	}

	ar_contents := ARContents{}
	err = json.Unmarshal(ARContentsAsBytes, &ar_contents)
	if err != nil {
		return nil, err
	}

	return &ar_contents, nil
}

// TransferARContents 함수


func (t *SimpleChaincode) TransferARContents(ctx contractapi.TransactionContextInterface, pid string, newOwner string) error {
	fmt.Println("- start transfer AR Contents")
	
	// 기등록 마블 검색

	ARContentsAsBytes, err := ctx.GetStub().GetState(pid)
	if err != nil {
		return fmt.Errorf("Failed to get AR Contents: " + err.Error())
	} else if ARContentsAsBytes == nil {
		return fmt.Errorf("This marble does not exists: " + pid)
	}

	// 검증 해당 ARContents가 New User에게 Transfer Approve 되었나?
	// Unmarshal 시키는 것 먼저
	ar_contents := ARContents{}
	_ = json.Unmarshal(ARContentsAsBytes, &ar_contents)
	// 수정 => 오너 변경
	ar_contents.Owner = newOwner

	ARContentsJSONasBytes, err := json.Marshal(ar_contents)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(pid, ARContentsJSONasBytes)
	if err != nil {
		return err
	}
	return nil
}

// GetARContentsHistory 함수

func (t *SimpleChaincode) GetARContentsHistory(ctx contractapi.TransactionContextInterface, ARContentsID string) ([]HistoryQueryResult, error) {
	log.Printf("GetARContentsHistory: ID %v", ARContentsID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(ARContentsID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var ar_contents ARContents
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &ar_contents)
			if err != nil {
				return nil, err
			}
		} else {
			ar_contents = ARContents{
				PID: ARContentsID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &ar_contents,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}


// main 함수

func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating AR Contents chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting AR Contents chaincode: %v", err)
	}
}