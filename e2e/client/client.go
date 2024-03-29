//go:build e2e
// +build e2e

package client

import (
	"context"
	"runtime"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/substra/orchestrator/lib/asset"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const DefaultCompositeTaskRef = "composite task"
const DefaultAggregateTaskRef = "aggregate task"
const DefaultPredictTaskRef = "predict task"
const DefaultTestTaskRef = "test task"
const DefaultTrainTaskRef = "task"

const DefaultSimpleFunctionRef = "simple function"
const DefaultCompositeFunctionRef = "composite function"
const DefaultAggregateFunctionRef = "aggregate function"
const DefaultPredictFunctionRef = "predict function"
const DefaultMetricFunctionRef = "metric function"

const DefaultPlanRef = "cp"
const DefaultDataManagerRef = "dm"
const DefaultDataSampleRef = "ds"
const DefaultModelRef = "model"

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
	MSPID                string
	Channel              string
	ctx                  context.Context
	ks                   *KeyStore
	logger               zerolog.Logger
	organizationService  asset.OrganizationServiceClient
	functionService      asset.FunctionServiceClient
	dataManagerService   asset.DataManagerServiceClient
	dataSampleService    asset.DataSampleServiceClient
	modelService         asset.ModelServiceClient
	computeTaskService   asset.ComputeTaskServiceClient
	computePlanService   asset.ComputePlanServiceClient
	performanceService   asset.PerformanceServiceClient
	datasetService       asset.DatasetServiceClient
	eventService         asset.EventServiceClient
	failureReportService asset.FailureReportServiceClient
}

type TestClientFactory struct {
	conn    *grpc.ClientConn
	mspid   string
	channel string
}

func NewTestClientFactory(conn *grpc.ClientConn, mspid, channel string) *TestClientFactory {
	return &TestClientFactory{
		conn: conn, mspid: mspid, channel: channel,
	}
}

func (f *TestClientFactory) NewTestClient() *TestClient {
	logger := log.With().
		Str("mspid", f.mspid).
		Str("channel", f.channel).
		Logger()

	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			logger = logger.With().Str("caller", fn.Name()).Logger()
		}
	}

	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "mspid", f.mspid, "channel", f.channel)

	client := &TestClient{
		MSPID:                f.mspid,
		Channel:              f.channel,
		ctx:                  ctx,
		ks:                   NewKeyStore(),
		logger:               logger,
		organizationService:  asset.NewOrganizationServiceClient(f.conn),
		functionService:      asset.NewFunctionServiceClient(f.conn),
		dataManagerService:   asset.NewDataManagerServiceClient(f.conn),
		dataSampleService:    asset.NewDataSampleServiceClient(f.conn),
		modelService:         asset.NewModelServiceClient(f.conn),
		computeTaskService:   asset.NewComputeTaskServiceClient(f.conn),
		computePlanService:   asset.NewComputePlanServiceClient(f.conn),
		performanceService:   asset.NewPerformanceServiceClient(f.conn),
		datasetService:       asset.NewDatasetServiceClient(f.conn),
		eventService:         asset.NewEventServiceClient(f.conn),
		failureReportService: asset.NewFailureReportServiceClient(f.conn),
	}

	client.EnsureOrganization()
	return client
}

func (f *TestClientFactory) WithMSPID(mspid string) *TestClientFactory {
	return &TestClientFactory{
		conn:    f.conn,
		mspid:   mspid,
		channel: f.channel,
	}
}

func (f *TestClientFactory) WithChannel(channel string) *TestClientFactory {
	return &TestClientFactory{
		conn:    f.conn,
		mspid:   f.mspid,
		channel: channel,
	}
}

func (c *TestClient) WithKeyStore(ks *KeyStore) *TestClient {
	c.ks = ks
	return c
}

func (c *TestClient) GetKeyStore() *KeyStore {
	return c.ks
}

// EnsureOrganization attempts to register the organization but won't fail on existing organization
func (c *TestClient) EnsureOrganization() {
	_, err := c.organizationService.RegisterOrganization(c.ctx, &asset.RegisterOrganizationParam{})
	if status.Code(err) == codes.AlreadyExists {
		c.logger.Debug().Msg("organization already exists")
		// expected error
		return
	}
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterOrganization failed")
	}
}

func (c *TestClient) RegisterFunction(o *FunctionOptions) *asset.Function {
	newFunction := &asset.NewFunction{
		Key:  c.ks.GetKey(o.KeyRef),
		Name: "Function test",
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc/" + uuid.NewString(),
		},
		Archive: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/function/" + uuid.NewString(),
		},
		NewPermissions: &asset.NewPermissions{Public: true},
		Inputs:         o.Inputs,
		Outputs:        o.Outputs,
	}
	c.logger.Debug().Interface("function", newFunction).Msg("registering function")
	function, err := c.functionService.RegisterFunction(c.ctx, newFunction)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterFunction failed")
	}
	return function
}

func (c *TestClient) RegisterDataManager(o *DataManagerOptions) *asset.DataManager {
	newDm := &asset.NewDataManager{
		Key:            c.ks.GetKey(DefaultDataManagerRef),
		Name:           "Test datamanager",
		NewPermissions: &asset.NewPermissions{Public: true},
		Description: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/desc" + uuid.NewString(),
		},
		Opener: &asset.Addressable{
			Checksum:       "1d55e9c55fa7ad6b6a49ad79da897d58be7ce8b76f92ced4c20f361ba3a0af6e",
			StorageAddress: "http://somewhere.local/opener" + uuid.NewString(),
		},
		LogsPermission: o.LogsPermission,
	}
	c.logger.Debug().Interface("datamanager", newDm).Msg("registering datamanager")
	dataManager, err := c.dataManagerService.RegisterDataManager(c.ctx, newDm)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterDataManager failed")
	}
	return dataManager
}

func (c *TestClient) GetDataManager(dataManagerRef string) *asset.DataManager {
	param := &asset.GetDataManagerParam{
		Key: c.ks.GetKey(dataManagerRef),
	}
	c.logger.Debug().Str("datamanager key", c.ks.GetKey(dataManagerRef)).Msg("GetDataManager")
	resp, err := c.dataManagerService.GetDataManager(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetDataManager failed")
	}
	return resp
}

func (c *TestClient) UpdateDataManager(dataManagerRef string, name string) *asset.UpdateDataManagerResponse {
	param := &asset.UpdateDataManagerParam{
		Key:  c.ks.GetKey(dataManagerRef),
		Name: name,
	}
	c.logger.Debug().Str("data manager key", c.ks.GetKey(dataManagerRef)).Msg("UpdateDataManager")
	resp, err := c.dataManagerService.UpdateDataManager(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("UpdateDataManager failed")
	}
	return resp
}

func (c *TestClient) RegisterDataSample(o *DataSampleOptions) *asset.DataSample {
	newDs := &asset.NewDataSample{
		Key:             c.ks.GetKey(o.KeyRef),
		DataManagerKeys: []string{c.ks.GetKey(DefaultDataManagerRef)},
		Checksum:        "7e87a07aeb05e0e66918ce1c93155acf54649eec453060b75caf494bc0bc0b9c",
	}
	c.logger.Debug().Interface("datasample", newDs).Msg("registering datasample")
	input := &asset.RegisterDataSamplesParam{
		Samples: []*asset.NewDataSample{newDs},
	}
	res, err := c.dataSampleService.RegisterDataSamples(c.ctx, input)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterDataSample failed")
	}
	return res.DataSamples[0]
}

func (c *TestClient) GetDataSample(dataSampleRef string) *asset.DataSample {
	param := &asset.GetDataSampleParam{
		Key: c.ks.GetKey(dataSampleRef),
	}
	c.logger.Debug().Interface("data sample key", c.ks.GetKey(dataSampleRef)).Msg("GetDataSample")
	resp, err := c.dataSampleService.GetDataSample(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetDataSample failed")
	}
	return resp
}

func (c *TestClient) QueryDataSamples(pageToken string, pageSize uint32, filter *asset.DataSampleQueryFilter) *asset.QueryDataSamplesResponse {
	param := &asset.QueryDataSamplesParam{
		PageToken: pageToken,
		PageSize:  pageSize,
		Filter:    filter,
	}
	c.logger.Debug().
		Str("pageToken", pageToken).
		Uint32("pageSize", pageSize).
		Interface("filter", filter).
		Msg("QueryDataSamples")

	resp, err := c.dataSampleService.QueryDataSamples(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("QueryDataSamples failed")
	}
	return resp
}

func (c *TestClient) RegisterTasks(optList ...Taskable) []*asset.ComputeTask {
	res, err := c.FailableRegisterTasks(optList...)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterTasks failed")
	}
	return res.Tasks
}

func (c *TestClient) FailableRegisterTasks(optList ...Taskable) (*asset.RegisterTasksResponse, error) {
	newTasks := make([]*asset.NewComputeTask, len(optList))
	for i, o := range optList {
		newTasks[i] = o.GetNewTask(c.ks)
	}
	c.logger.Debug().Int("nbTasks", len(newTasks)).Msg("registering tasks")
	return c.computeTaskService.RegisterTasks(c.ctx, &asset.RegisterTasksParam{Tasks: newTasks})
}

func (c *TestClient) StartTask(keyRef string) {
	c.applyTaskTransition(keyRef, asset.ComputeTaskAction_TASK_ACTION_EXECUTING)
}

func (c *TestClient) CancelTask(keyRef string) {
	c.applyTaskTransition(keyRef, asset.ComputeTaskAction_TASK_ACTION_CANCELED)
}

func (c *TestClient) FailTask(keyRef string) {
	c.applyTaskTransition(keyRef, asset.ComputeTaskAction_TASK_ACTION_FAILED)
}

func (c *TestClient) DoneTask(keyRef string) {
	c.applyTaskTransition(keyRef, asset.ComputeTaskAction_TASK_ACTION_DONE)
}

func (c *TestClient) applyTaskTransition(keyRef string, action asset.ComputeTaskAction) {
	taskKey := c.ks.GetKey(keyRef)
	c.logger.Debug().Str("taskKey", taskKey).Str("action", action.String()).Msg("applying task action")
	_, err := c.computeTaskService.ApplyTaskAction(c.ctx, &asset.ApplyTaskActionParam{
		ComputeTaskKey: taskKey,
		Action:         action,
	})
	if err != nil {
		c.logger.Fatal().Err(err).Msgf("failed to mark task as %v", action)
	}
}

func (c *TestClient) RegisterModel(o *ModelOptions) *asset.Model {
	newModel := c.makeNewModel(o)
	c.logger.Debug().Interface("model", newModel).Msg("registering model")
	//nolint: staticcheck //This method is deprecated but still needs to be tested
	model, err := c.modelService.RegisterModel(c.ctx, newModel)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterModel failed")
	}
	return model
}

func (c *TestClient) GetModel(modelRef string) *asset.Model {
	param := &asset.GetModelParam{
		Key: c.ks.GetKey(modelRef),
	}
	c.logger.Debug().Str("model key", c.ks.GetKey(modelRef)).Msg("GetModel")
	resp, err := c.modelService.GetModel(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetModel failed")
	}
	return resp
}

func (c *TestClient) FailableRegisterModels(o ...*ModelOptions) (*asset.RegisterModelsResponse, error) {
	newModels := make([]*asset.NewModel, len(o))
	for i, modelOpt := range o {
		newModel := c.makeNewModel(modelOpt)
		c.logger.Debug().Interface("model", newModel).Msg("registering model")
		newModels[i] = newModel
	}
	return c.modelService.RegisterModels(c.ctx, &asset.RegisterModelsParam{Models: newModels})
}

func (c *TestClient) RegisterModels(o ...*ModelOptions) []*asset.Model {
	res, err := c.FailableRegisterModels(o...)
	if err != nil {
		log.Fatal().Err(err).Msg("RegisterModels failed")
	}
	return res.Models
}

func (c *TestClient) GetTaskOutputModels(taskRef string) []*asset.Model {
	resp, err := c.modelService.GetComputeTaskOutputModels(c.ctx, &asset.GetComputeTaskModelsParam{ComputeTaskKey: c.ks.GetKey(taskRef)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetComputeTaskOutputModels failed")
	}

	return resp.Models
}

func (c *TestClient) DisableOutput(taskRef string, identifier string) {
	taskKey := c.ks.GetKey(taskRef)
	c.logger.Debug().Str("taskKey", taskKey).Str("identifier", identifier).Msg("disabling output")
	_, err := c.computeTaskService.DisableOutput(c.ctx, &asset.DisableOutputParam{ComputeTaskKey: taskKey, Identifier: identifier})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("DisableOutput failed")
	}
}

func (c *TestClient) RegisterComputePlan(o *ComputePlanOptions) *asset.ComputePlan {
	newCp := &asset.NewComputePlan{
		Key:  c.ks.GetKey(o.KeyRef),
		Name: "Compute plan test",
	}
	c.logger.Debug().Interface("plan", newCp).Msg("registering compute plan")
	plan, err := c.computePlanService.RegisterPlan(c.ctx, newCp)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterPlan failed")
	}
	return plan
}

func (c *TestClient) GetComputePlan(keyRef string) *asset.ComputePlan {
	plan, err := c.computePlanService.GetPlan(c.ctx, &asset.GetComputePlanParam{Key: c.ks.GetKey(keyRef)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetPlan failed")
	}

	return plan
}

func (c *TestClient) GetComputeTask(keyRef string) *asset.ComputeTask {
	task, err := c.computeTaskService.GetTask(c.ctx, &asset.GetTaskParam{Key: c.ks.GetKey(keyRef)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetTask failed")
	}

	return task
}

func (c *TestClient) QueryTasks(filter *asset.TaskQueryFilter, pageToken string, pageSize int) *asset.QueryTasksResponse {
	resp, err := c.computeTaskService.QueryTasks(c.ctx, &asset.QueryTasksParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("QueryTasks failed")
	}

	return resp
}

func (c *TestClient) RegisterPerformance(o *PerformanceOptions) (*asset.Performance, error) {
	newPerf := &asset.NewPerformance{
		ComputeTaskKey:              c.ks.GetKey(o.ComputeTaskKeyRef),
		ComputeTaskOutputIdentifier: o.ComputeTaskOutput,
		PerformanceValue:            o.PerformanceValue,
	}

	c.logger.Debug().Interface("performance", newPerf).Msg("registering performance")
	return c.performanceService.RegisterPerformance(c.ctx, newPerf)
}

func (c *TestClient) QueryPerformances(filter *asset.PerformanceQueryFilter, pageToken string, pageSize int) *asset.QueryPerformancesResponse {
	resp, err := c.performanceService.QueryPerformances(c.ctx, &asset.QueryPerformancesParam{
		Filter:    filter,
		PageToken: pageToken,
		PageSize:  uint32(pageSize),
	})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("QueryPerformances failed")
	}
	return resp
}

func (c *TestClient) GetTaskInputAssets(taskRef string) []*asset.ComputeTaskInputAsset {
	c.logger.Debug().Str("task key", c.ks.GetKey(taskRef)).Msg("GetComputeTaskInputAssets")
	assets, err := c.FailableGetTaskInputAssets(taskRef)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("Task input assets retrieval failed")
	}
	return assets
}

func (c *TestClient) FailableGetTaskInputAssets(taskRef string) ([]*asset.ComputeTaskInputAsset, error) {
	param := &asset.GetTaskInputAssetsParam{
		ComputeTaskKey: c.ks.GetKey(taskRef),
	}
	resp, err := c.computeTaskService.GetTaskInputAssets(c.ctx, param)
	if err != nil {
		return nil, err
	}
	return resp.Assets, nil
}

func (c *TestClient) GetDataset(dataManagerRef string) *asset.Dataset {
	param := &asset.GetDatasetParam{
		Key: c.ks.GetKey(dataManagerRef),
	}
	c.logger.Debug().Str("data manager key", c.ks.GetKey(dataManagerRef)).Msg("GetDataset")
	resp, err := c.datasetService.GetDataset(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetDataset failed")
	}
	return resp
}

func (c *TestClient) QueryFunctions(filter *asset.FunctionQueryFilter, pageToken string, pageSize int) *asset.QueryFunctionsResponse {
	resp, err := c.functionService.QueryFunctions(c.ctx, &asset.QueryFunctionsParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("QueryFunctions failed")
	}

	return resp
}

func (c *TestClient) GetFunction(functionRef string) *asset.Function {
	param := &asset.GetFunctionParam{
		Key: c.ks.GetKey(functionRef),
	}
	c.logger.Debug().Str("function key", c.ks.GetKey(functionRef)).Msg("GetFunction")
	resp, err := c.functionService.GetFunction(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetFunction failed")
	}
	return resp
}

func (c *TestClient) GetAssetCreationEvent(assetKey string) *asset.Event {
	filter := &asset.EventQueryFilter{AssetKey: assetKey, EventKind: asset.EventKind_EVENT_ASSET_CREATED}
	resp := c.QueryEvents(filter, "", 1)
	return resp.Events[0]
}

func (c *TestClient) UpdateFunction(functionRef string, name string, image *asset.Addressable) *asset.UpdateFunctionResponse {
	param := &asset.UpdateFunctionParam{
		Key:   c.ks.GetKey(functionRef),
		Name:  name,
		Image: image,
	}
	c.logger.Debug().Str("function key", c.ks.GetKey(functionRef)).Msg("UpdateFunction")
	resp, err := c.functionService.UpdateFunction(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("UpdateFunction failed")
	}
	return resp
}

func (c *TestClient) BuildFunction(keyRef string) {
	c.applyFunctionAction(keyRef, asset.FunctionAction_FUNCTION_ACTION_BUILDING)
}

func (c *TestClient) CancelFunction(keyRef string) {
	c.applyFunctionAction(keyRef, asset.FunctionAction_FUNCTION_ACTION_CANCELED)
}

func (c *TestClient) FailFunction(keyRef string) {
	c.applyFunctionAction(keyRef, asset.FunctionAction_FUNCTION_ACTION_FAILED)
}

func (c *TestClient) SetReadyFunction(keyRef string) {
	c.applyFunctionAction(keyRef, asset.FunctionAction_FUNCTION_ACTION_READY)
}

func (c *TestClient) SetReadyFromWaitingFunction(keyRef string) {
	c.BuildFunction(keyRef)
	c.SetReadyFunction(keyRef)
}

func (c *TestClient) applyFunctionAction(keyRef string, action asset.FunctionAction) {
	functionKey := c.ks.GetKey(keyRef)
	c.logger.Debug().Str("functionKey", functionKey).Str("action", action.String()).Msg("applying function action")
	_, err := c.functionService.ApplyFunctionAction(c.ctx, &asset.ApplyFunctionActionParam{
		FunctionKey: functionKey,
		Action:      action,
	})
	if err != nil {
		c.logger.Fatal().Err(err).Msgf("failed to mark function as %v", action)
	}
}

func (c *TestClient) QueryEvents(filter *asset.EventQueryFilter, pageToken string, pageSize int) *asset.QueryEventsResponse {
	resp, err := c.eventService.QueryEvents(c.ctx, &asset.QueryEventsParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("QueryEvents failed")
	}

	return resp
}

func (c *TestClient) SubscribeToEvents(startEventID string) (asset.EventService_SubscribeToEventsClient, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.ctx)

	stream, err := c.eventService.SubscribeToEvents(ctx, &asset.SubscribeToEventsParam{StartEventId: startEventID})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("SubscribeToEvents failed")
	}
	return stream, cancel
}

func (c *TestClient) QueryPlans(filter *asset.PlanQueryFilter, pageToken string, pageSize int) *asset.QueryPlansResponse {
	resp, err := c.computePlanService.QueryPlans(c.ctx, &asset.QueryPlansParam{Filter: filter, PageToken: pageToken, PageSize: uint32(pageSize)})
	if err != nil {
		c.logger.Fatal().Err(err).Msg("QueryPlans failed")
	}

	return resp
}

func (c *TestClient) RegisterTaskFailureReport(assetRef string) *asset.FailureReport {
	newFailureReport := &asset.NewFailureReport{
		AssetKey:  c.ks.GetKey(assetRef),
		AssetType: asset.FailedAssetKind_FAILED_ASSET_COMPUTE_TASK,
		ErrorType: asset.ErrorType_ERROR_TYPE_EXECUTION,
		LogsAddress: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.local/failure/" + uuid.NewString(),
		},
	}

	c.logger.Debug().Interface("failureReport", newFailureReport).Msg("registering failure report")
	failureReport, err := c.failureReportService.RegisterFailureReport(c.ctx, newFailureReport)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("RegisterFailureReport failed")
	}

	return failureReport
}

func (c *TestClient) GetFailureReport(assetRef string) *asset.FailureReport {
	param := &asset.GetFailureReportParam{
		AssetKey: c.ks.GetKey(assetRef),
	}

	c.logger.Debug().Str("asset key", param.AssetKey).Msg("getting failure report")
	failureReport, err := c.failureReportService.GetFailureReport(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("GetFailureReport failed")
	}

	return failureReport
}

func (c *TestClient) CancelComputePlan(computePlanRef string) (*asset.ApplyPlanActionResponse, error) {
	param := &asset.ApplyPlanActionParam{
		Key:    c.ks.GetKey(computePlanRef),
		Action: asset.ComputePlanAction_PLAN_ACTION_CANCELED,
	}

	c.logger.Debug().Str("compute plan key", computePlanRef).Msg("cancelling compute plan")
	return c.computePlanService.ApplyPlanAction(c.ctx, param)
}

func (c *TestClient) UpdateComputePlan(computePlanRef string, name string) *asset.UpdateComputePlanResponse {
	param := &asset.UpdateComputePlanParam{
		Key:  c.ks.GetKey(computePlanRef),
		Name: name,
	}
	c.logger.Debug().Str("compute plan key", c.ks.GetKey(computePlanRef)).Msg("UpdateComputePlan")
	resp, err := c.computePlanService.UpdatePlan(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("UpdateComputePlan failed")
	}
	return resp
}

func (c *TestClient) IsPlanRunning(computePlanRef string) *asset.IsPlanRunningResponse {
	param := &asset.IsPlanRunningParam{
		Key: c.ks.GetKey(computePlanRef),
	}

	c.logger.Debug().Str("compute plan key", computePlanRef).Msg("getting compute plan running status")

	resp, err := c.computePlanService.IsPlanRunning(c.ctx, param)
	if err != nil {
		c.logger.Fatal().Err(err).Msg("IsPlanRunning failed")
	}
	return resp
}

func (c *TestClient) makeNewModel(o *ModelOptions) *asset.NewModel {
	return &asset.NewModel{
		ComputeTaskKey:              c.ks.GetKey(o.TaskRef),
		ComputeTaskOutputIdentifier: o.TaskOutput,
		Key:                         c.ks.GetKey(o.KeyRef),
		Address: &asset.Addressable{
			Checksum:       "5e12e1a2687d81b268558217856547f8a4519f9688933351386a7f902cf1ce5d",
			StorageAddress: "http://somewhere.online/model/" + uuid.NewString(),
		},
	}
}
