package client

import (
	"context"
	"sync"

	"github.com/go-playground/log/v7"
	"github.com/google/uuid"
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const DefaultTaskRef = "task"
const DefaultPlanRef = "cp"
const DefaultAlgoRef = "algo"
const DefaultObjectiveRef = "objective"
const DefaultDataManagerRef = "dm"

// Taskable represent the ability to create a new task
type Taskable interface {
	GetNewTask(ks *KeyStore) *asset.NewComputeTask
}

// KeyStore is the component holding matching between key references and their UUID.
type KeyStore struct {
	keys map[string]string
	lock *sync.RWMutex
}

func NewKeyStore() *KeyStore {
	return &KeyStore{
		keys: make(map[string]string),
		lock: new(sync.RWMutex),
	}
}

// GetKey will create a UUID or return the previously generated one.
// This is useful when building relationships between entities.
func (ks *KeyStore) GetKey(id string) string {
	ks.lock.RLock()
	k, ok := ks.keys[id]
	ks.lock.RUnlock()

	if !ok {
		k = uuid.New().String()
		ks.lock.Lock()
		ks.keys[id] = k
		ks.lock.Unlock()
	}

	return k
}

// TestClient is a client for the tested app
type TestClient struct {
	ctx                context.Context
	ks                 *KeyStore
	nodeService        asset.NodeServiceClient
	objectiveService   asset.ObjectiveServiceClient
	algoService        asset.AlgoServiceClient
	dataManagerService asset.DataManagerServiceClient
	dataSampleService  asset.DataSampleServiceClient
	modelService       asset.ModelServiceClient
	computeTaskService asset.ComputeTaskServiceClient
	computePlanService asset.ComputePlanServiceClient
	performanceService asset.PerformanceServiceClient
	datasetService     asset.DatasetServiceClient
	eventService       asset.EventServiceClient
}

func NewTestClient(conn *grpc.ClientConn, mspid, channel, chaincode string) (*TestClient, error) {
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", mspid, "channel", channel, "chaincode", chaincode)

	return &TestClient{
		ctx:                ctx,
		ks:                 NewKeyStore(),
		nodeService:        asset.NewNodeServiceClient(conn),
		algoService:        asset.NewAlgoServiceClient(conn),
		objectiveService:   asset.NewObjectiveServiceClient(conn),
		dataManagerService: asset.NewDataManagerServiceClient(conn),
		dataSampleService:  asset.NewDataSampleServiceClient(conn),
		modelService:       asset.NewModelServiceClient(conn),
		computeTaskService: asset.NewComputeTaskServiceClient(conn),
		computePlanService: asset.NewComputePlanServiceClient(conn),
		performanceService: asset.NewPerformanceServiceClient(conn),
		datasetService:     asset.NewDatasetServiceClient(conn),
		eventService:       asset.NewEventServiceClient(conn),
	}, nil
}

func (c *TestClient) WithKeyStore(ks *KeyStore) *TestClient {
	c.ks = ks
	return c
}

func (c *TestClient) GetKeyStore() *KeyStore {
	return c.ks
}

// EnsureNode attempts to register the node but won't fail on existing node
func (c *TestClient) EnsureNode() {
	_, err := c.nodeService.RegisterNode(c.ctx, &asset.RegisterNodeParam{})
	if status.Code(err) == codes.AlreadyExists {
		log.Debug("node already exists")
		// expected error
		return
	}
	if err != nil {
		log.WithError(err).Fatal("RegisterNode failed")
	}
}

func (c *TestClient) RegisterAlgo(o *AlgoOptions) {
	newAlgo := &asset.NewAlgo{
		Key:      c.ks.GetKey(o.KeyRef),
		Name:     "Algo test",
		Category: o.Category,
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc",
		},
		Algorithm: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/algo",
		},
		NewPermissions: &asset.NewPermissions{Public: true},
	}
	log.WithField("algo", newAlgo).Debug("registering algo")
	_, err := c.algoService.RegisterAlgo(c.ctx, newAlgo)
	if err != nil {
		log.WithError(err).Fatal("RegisterAlgo failed")
	}

}

func (c *TestClient) RegisterDataManager() {
	newDm := &asset.NewDataManager{
		Key:            c.ks.GetKey(DefaultDataManagerRef),
		Name:           "Test datamanager",
		NewPermissions: &asset.NewPermissions{Public: true},
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc",
		},
		Opener: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/opener",
		},
		Type: "test",
	}
	log.WithField("datamanager", newDm).Debug("registering datamanager")
	_, err := c.dataManagerService.RegisterDataManager(c.ctx, newDm)
	if err != nil {
		log.WithError(err).Fatal("RegisterDataManager failed")
	}

}

func (c *TestClient) RegisterDataSample(o *DataSampleOptions) {
	newDs := &asset.NewDataSample{
		Key:             c.ks.GetKey(o.KeyRef),
		DataManagerKeys: []string{c.ks.GetKey(DefaultDataManagerRef)},
		TestOnly:        o.TestOnly,
		Checksum:        "7e87a07aeb05e0e66918ce1c93155acf54649eec453060b75caf494bc0bc0b9c",
	}
	log.WithField("datasample", newDs).Debug("registering datasample")
	input := &asset.RegisterDataSamplesParam{
		Samples: []*asset.NewDataSample{newDs},
	}
	_, err := c.dataSampleService.RegisterDataSamples(c.ctx, input)
	if err != nil {
		log.WithError(err).Fatal("RegisterDataSample failed")
	}
}

func (c *TestClient) RegisterObjective(o *ObjectiveOptions) {
	newObj := &asset.NewObjective{
		Key:            c.ks.GetKey(o.KeyRef),
		Name:           "test objective",
		DataManagerKey: c.ks.GetKey(o.DataManagerRef),
		DataSampleKeys: []string{c.ks.GetKey(o.DataSampleRef)},
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc",
		},
		MetricsName: "test metrics",
		Metrics: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/metrics",
		},
		NewPermissions: &asset.NewPermissions{Public: true},
	}

	log.WithField("objective", newObj).Debug("registering objective")
	_, err := c.objectiveService.RegisterObjective(c.ctx, newObj)
	if err != nil {
		log.WithError(err).Fatal("RegisterObjective failed")
	}
}

func (c *TestClient) RegisterTasks(optList ...Taskable) {
	newTasks := make([]*asset.NewComputeTask, len(optList))
	for i, o := range optList {
		newTasks[i] = o.GetNewTask(c.ks)
	}
	log.WithField("nbTasks", len(newTasks)).Debug("registering tasks")
	_, err := c.computeTaskService.RegisterTasks(c.ctx, &asset.RegisterTasksParam{Tasks: newTasks})
	if err != nil {
		log.WithError(err).Fatal("RegisterTasks failed")
	}
}

func (c *TestClient) StartTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_DOING)
}

func (c *TestClient) CancelTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_CANCELED)
}

func (c *TestClient) FailTask(keyRef string) {
	c.applyTaskAction(keyRef, asset.ComputeTaskAction_TASK_ACTION_FAILED)
}

func (c *TestClient) applyTaskAction(keyRef string, action asset.ComputeTaskAction) {
	taskKey := c.ks.GetKey(keyRef)
	log.WithField("taskKey", taskKey).WithField("action", action).Debug("applying task action")
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: taskKey,
		Action:         action,
	})
	if err != nil {
		log.WithError(err).Fatalf("failed to mark task as %v", action)
	}
}

func (c *TestClient) RegisterModel(o *ModelOptions) {
	newModel := &asset.NewModel{
		ComputeTaskKey: c.ks.GetKey(o.TaskRef),
		Key:            c.ks.GetKey(o.KeyRef),
		Category:       o.Category,
		Address: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.online/model",
		},
	}
	log.WithField("model", newModel).Debug("registering model")
	_, err := c.modelService.RegisterModel(c.ctx, newModel)
	if err != nil {
		log.WithError(err).Fatal("RegisterModel failed")
	}
}

func (c *TestClient) GetTaskOutputModels(taskRef string) []*asset.Model {
	resp, err := c.modelService.GetComputeTaskOutputModels(c.ctx, &asset.GetComputeTaskModelsParam{ComputeTaskKey: c.ks.GetKey(taskRef)})
	if err != nil {
		log.WithError(err).Fatal("GetComputeTaskOutputModels failed")
	}

	return resp.Models
}

func (c *TestClient) CanDisableModel(modelRef string) bool {
	resp, err := c.modelService.CanDisableModel(c.ctx, &asset.CanDisableModelParam{ModelKey: c.ks.GetKey(modelRef)})
	if err != nil {
		log.WithError(err).Fatal("CanDisableModel failed")
	}

	return resp.CanDisable
}

func (c *TestClient) DisableModel(modelRef string) {
	modelKey := c.ks.GetKey(modelRef)
	log.WithField("modelKey", modelKey).Debug("disabling model")
	_, err := c.modelService.DisableModel(c.ctx, &asset.DisableModelParam{ModelKey: modelKey})
	if err != nil {
		log.WithError(err).Fatal("DisableModel failed")
	}
}

func (c *TestClient) RegisterComputePlan(o *ComputePlanOptions) {
	newCp := &asset.NewComputePlan{
		Key:                      c.ks.GetKey(o.KeyRef),
		DeleteIntermediaryModels: o.DeleteIntermediaryModels,
	}
	log.WithField("plan", newCp).Debug("registering compute plan")
	_, err := c.computePlanService.RegisterPlan(c.ctx, newCp)
	if err != nil {
		log.WithError(err).Fatal("RegisterPlan failed")
	}
}

func (c *TestClient) GetComputePlan(keyRef string) *asset.ComputePlan {
	plan, err := c.computePlanService.GetPlan(c.ctx, &asset.GetComputePlanParam{Key: c.ks.GetKey(keyRef)})
	if err != nil {
		log.WithError(err).Fatal("GetPlan failed")
	}

	return plan
}

func (c *TestClient) GetComputeTask(keyRef string) *asset.ComputeTask {
	task, err := c.computeTaskService.GetTask(c.ctx, &asset.GetTaskParam{Key: c.ks.GetKey(keyRef)})
	if err != nil {
		log.WithError(err).Fatal("GetTask failed")
	}

	return task
}

func (c *TestClient) QueryTasks(filter *asset.TaskQueryFilter, pageToken string, pageSize int) *asset.QueryTasksResponse {
	resp, err := c.computeTaskService.QueryTasks(c.ctx, &asset.QueryTasksParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		log.WithError(err).Fatal("QueryTasks failed")
	}

	return resp
}

func (c *TestClient) RegisterPerformance(o *PerformanceOptions) {
	newPerf := &asset.NewPerformance{
		ComputeTaskKey:   c.ks.GetKey(o.KeyRef),
		PerformanceValue: o.PerformanceValue,
	}

	log.WithField("performance", newPerf).Debug("registering performance")
	_, err := c.performanceService.RegisterPerformance(c.ctx, newPerf)
	if err != nil {
		log.WithError(err).Fatal("RegisterPerformance failed")
	}
}

func (c *TestClient) GetTaskPerformance(taskRef string) *asset.Performance {
	param := &asset.GetComputeTaskPerformanceParam{ComputeTaskKey: c.ks.GetKey(taskRef)}
	perf, err := c.performanceService.GetComputeTaskPerformance(c.ctx, param)
	if err != nil {
		log.WithError(err).Fatal("GetTaskPerformance failed")
	}

	return perf
}

func (c *TestClient) GetLeaderboard(key string) *asset.Leaderboard {
	param := &asset.LeaderboardQueryParam{ObjectiveKey: c.ks.GetKey(key), SortOrder: asset.SortOrder_ASCENDING}
	log.WithField("objective key", c.ks.GetKey(key)).Debug("GetLeaderboard")
	resp, err := c.objectiveService.GetLeaderboard(c.ctx, param)
	if err != nil {
		log.WithError(err).Fatal("Query Leaderboard failed")
	}
	return resp
}

func (c *TestClient) GetInputModels(taskRef string) []*asset.Model {
	param := &asset.GetComputeTaskModelsParam{
		ComputeTaskKey: c.ks.GetKey(taskRef),
	}
	log.WithField("task key", c.ks.GetKey(taskRef)).Debug("GetComputeTaskInputModels")
	resp, err := c.modelService.GetComputeTaskInputModels(c.ctx, param)
	if err != nil {
		log.WithError(err).Fatal("Task input model retrieval failed")
	}
	return resp.Models
}

func (c *TestClient) GetDataset(dataManagerRef string) *asset.Dataset {
	param := &asset.GetDatasetParam{
		Key: c.ks.GetKey(dataManagerRef),
	}
	log.WithField("data manager key", c.ks.GetKey(dataManagerRef)).Debug("GetDataset")
	resp, err := c.datasetService.GetDataset(c.ctx, param)
	if err != nil {
		log.WithError(err).Fatal("GetDataset failed")
	}
	return resp
}

func (c *TestClient) QueryAlgos(filter *asset.AlgoQueryFilter, pageToken string, pageSize int) *asset.QueryAlgosResponse {
	resp, err := c.algoService.QueryAlgos(c.ctx, &asset.QueryAlgosParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		log.WithError(err).Fatal("QueryAlgos failed")
	}

	return resp
}

func (c *TestClient) QueryEvents(filter *asset.EventQueryFilter, pageToken string, pageSize int) *asset.QueryEventsResponse {
	resp, err := c.eventService.QueryEvents(c.ctx, &asset.QueryEventsParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		log.WithError(err).Fatal("QueryEvents failed")
	}

	return resp
}

func (c *TestClient) QueryPlans(pageToken string, pageSize int) *asset.QueryPlansResponse {
	resp, err := c.computePlanService.QueryPlans(c.ctx, &asset.QueryPlansParam{PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		log.WithError(err).Fatal("QueryPlans failed")
	}

	return resp
}
