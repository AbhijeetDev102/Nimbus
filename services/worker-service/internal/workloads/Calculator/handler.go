package calculator

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/AbhijeetDev102/Nimbus/pkg/nimbus"
	"gorm.io/datatypes"
)

type values struct {
	Num1 int64 `json:"num1"`
	Num2 int64 `json:"num2"`
}

type calculatorHandler struct {
}

func NewCalculator() *calculatorHandler {
	return &calculatorHandler{}
}

func (c *calculatorHandler) Execute(ctx nimbus.Context, job *nimbus.Job) (*nimbus.ExecutionResult, error) {
	var values values

	if err := json.Unmarshal(job.Parameters, &values); err != nil {
		return nil, fmt.Errorf("failed to unmarsher job parameters : %v", err)
	}

	ans := values.Num1 + values.Num2
	resultPayload := map[string]any{
		"num1":   values.Num1,
		"num2":   values.Num2,
		"result": ans,
	}
	log.Printf("[Calculator] Executing job %s: %d + %d", job.ID, values.Num1, values.Num2)
	ctx.ReportProgress(100.0)

	metadataJSON, _ := json.Marshal(resultPayload)
	return &nimbus.ExecutionResult{
		Metadata: datatypes.JSON(metadataJSON),
	}, nil
}
