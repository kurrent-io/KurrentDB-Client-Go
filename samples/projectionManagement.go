package samples

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func CreateClient(connectionString string) {
	// region createClient
	conf, err := kurrentdb.ParseConnectionString(connectionString)

	if err != nil {
		panic(err)
	}

	client, err := kurrentdb.NewProjectionClient(conf)

	if err != nil {
		panic(err)
	}
	// endregion createClient

	defer client.Close()
}

func Disable(client *kurrentdb.ProjectionClient) {
	// region Disable
	err := client.Disable(context.Background(), "$by_category", kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion Disable
}

func DisableNotFound(client *kurrentdb.ProjectionClient) {
	// region DisableNotFound
	err := client.Disable(context.Background(), "projection that doesn't exist", kurrentdb.GenericProjectionOptions{})

	if esdbError, ok := kurrentdb.FromError(err); !ok {
		if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceNotFound) {
			log.Printf("projection not found")
			return
		}
	}
	// endregion DisableNotFound
}

func Enable(client *kurrentdb.ProjectionClient) {
	// region Enable
	err := client.Enable(context.Background(), "$by_category", kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion Enable
}

func EnableNotFound(client *kurrentdb.ProjectionClient) {
	// region EnableNotFound
	err := client.Enable(context.Background(), "projection that doesn't exist", kurrentdb.GenericProjectionOptions{})

	if esdbError, ok := kurrentdb.FromError(err); !ok {
		if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceNotFound) {
			log.Printf("projection not found")
			return
		}
	}
	// endregion EnableNotFound
}

func Delete(client *kurrentdb.ProjectionClient) {
	// region Delete
	err := client.Delete(context.Background(), "$by_category", kurrentdb.DeleteProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion Delete
}

func DeleteNotFound(client *kurrentdb.ProjectionClient) {
	// region DeleteNotFound
	err := client.Delete(context.Background(), "projection that doesn't exist", kurrentdb.DeleteProjectionOptions{})

	if esdbError, ok := kurrentdb.FromError(err); !ok {
		if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceNotFound) {
			log.Printf("projection not found")
			return
		}
	}
	// endregion DeleteNotFound
}

func Abort(client *kurrentdb.ProjectionClient) {
	// region Abort
	err := client.Abort(context.Background(), "$by_category", kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion Abort
}

func AbortNotFound(client *kurrentdb.ProjectionClient) {
	// region Abort_NotFound
	err := client.Abort(context.Background(), "projection that doesn't exist", kurrentdb.GenericProjectionOptions{})

	if esdbError, ok := kurrentdb.FromError(err); !ok {
		if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceNotFound) {
			log.Printf("projection not found")
			return
		}
	}
	// endregion Abort_NotFound
}

func Reset(client *kurrentdb.ProjectionClient) {
	// region Reset
	err := client.Reset(context.Background(), "$by_category", kurrentdb.ResetProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion Reset
}

func ResetNotFound(client *kurrentdb.ProjectionClient) {
	// region Reset_NotFound
	err := client.Reset(context.Background(), "projection that doesn't exist", kurrentdb.ResetProjectionOptions{})

	if esdbError, ok := kurrentdb.FromError(err); !ok {
		if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceNotFound) {
			log.Printf("projection not found")
			return
		}
	}
	// endregion Reset_NotFound
}

func Create(client *kurrentdb.ProjectionClient) {
	// region CreateContinuous
	script := `
fromAll()
.when({
	$init:function(){
		return {
			count: 0
		}
	},
	myEventUpdatedType: function(state, event){
		state.count += 1;
	}
})
.transformBy(function(state){
	state.count = 10;
})
.outputState()
`
	name := fmt.Sprintf("countEvent_Create_%s", uuid.New())
	err := client.Create(context.Background(), name, script, kurrentdb.CreateProjectionOptions{})

	if err != nil {
		panic(err)
	}

	// endregion CreateContinuous
}

func CreateConflict(client *kurrentdb.ProjectionClient) {
	script := ""
	name := ""

	// region CreateContinuous_Conflict
	err := client.Create(context.Background(), name, script, kurrentdb.CreateProjectionOptions{})

	if esdbErr, ok := kurrentdb.FromError(err); !ok {
		if esdbErr.IsErrorCode(kurrentdb.ErrorCodeUnknown) && strings.Contains(esdbErr.Err().Error(), "Conflict") {
			log.Printf("projection %s already exists", name)
			return
		}
	}
	// endregion CreateContinuous_Conflict
}

func Update(client *kurrentdb.ProjectionClient) {
	script := ""
	newScript := ""
	name := ""

	// region Update
	err := client.Create(context.Background(), name, script, kurrentdb.CreateProjectionOptions{})

	if err != nil {
		panic(err)
	}

	err = client.Update(context.Background(), name, newScript, kurrentdb.UpdateProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion Update

}

func UpdateNotFound(client *kurrentdb.ProjectionClient) {
	script := ""

	// region Update_NotFound
	err := client.Update(context.Background(), "projection that doesn't exist", script, kurrentdb.UpdateProjectionOptions{})

	if esdbError, ok := kurrentdb.FromError(err); !ok {
		if esdbError.IsErrorCode(kurrentdb.ErrorCodeResourceNotFound) {
			log.Printf("projection not found")
			return
		}
	}
	// endregion Update_NotFound
}

func ListAll(client *kurrentdb.ProjectionClient) {
	// region ListAll
	projections, err := client.ListAll(context.Background(), kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}

	for i := range projections {
		projection := projections[i]

		log.Printf(
			"%s, %s, %s, %s, %f",
			projection.Name,
			projection.Status,
			projection.CheckpointStatus,
			projection.Mode,
			projection.Progress,
		)
	}
	// endregion ListAll
}

func List(client *kurrentdb.ProjectionClient) {
	// region ListContinuous
	projections, err := client.ListContinuous(context.Background(), kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}

	for i := range projections {
		projection := projections[i]

		log.Printf(
			"%s, %s, %s, %s, %f",
			projection.Name,
			projection.Status,
			projection.CheckpointStatus,
			projection.Mode,
			projection.Progress,
		)
	}
	// endregion ListContinuous
}

func GetStatus(client *kurrentdb.ProjectionClient) {
	// region GetStatus
	projection, err := client.GetStatus(context.Background(), "$by_category", kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}

	log.Printf(
		"%s, %s, %s, %s, %f",
		projection.Name,
		projection.Status,
		projection.CheckpointStatus,
		projection.Mode,
		projection.Progress,
	)
	// endregion GetStatus
}

func GetState(client *kurrentdb.ProjectionClient) {
	projectionName := ""
	// region GetState
	type Foobar struct {
		Count int64
	}

	value, err := client.GetState(context.Background(), projectionName, kurrentdb.GetStateProjectionOptions{})

	if err != nil {
		panic(err)
	}

	jsonContent, err := value.MarshalJSON()

	if err != nil {
		panic(err)
	}

	var foobar Foobar

	if err = json.Unmarshal(jsonContent, &foobar); err != nil {
		panic(err)
	}

	log.Printf("count %d", foobar.Count)
	// endregion GetState
}

func GetResult(client *kurrentdb.ProjectionClient) {
	projectionName := ""
	// region GetResult
	type Baz struct {
		Result int64
	}

	value, err := client.GetResult(context.Background(), projectionName, kurrentdb.GetResultProjectionOptions{})

	if err != nil {
		panic(err)
	}

	jsonContent, err := value.MarshalJSON()

	if err != nil {
		panic(err)
	}

	var baz Baz

	if err = json.Unmarshal(jsonContent, &baz); err != nil {
		panic(err)
	}

	log.Printf("result %d", baz.Result)
	// endregion GetResult
}

func RestartSubSystem(client *kurrentdb.ProjectionClient) {
	// region RestartSubSystem
	err := client.RestartSubsystem(context.Background(), kurrentdb.GenericProjectionOptions{})

	if err != nil {
		panic(err)
	}
	// endregion RestartSubSystem
}
