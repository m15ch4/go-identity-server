package main

import (
	"sync"

	"github.com/google/uuid"
)

type VMCreationTask struct {
	TaskID       string
	DeploymentID string
}

type VMDeployment struct {
	ID         string
	ResourceID string
	Status     string
}

type DeploymentUpdate struct {
	DeploymentID string
	ResourceID   string
}

type VM struct {
	ID       string
	Name     string
	NumCPUs  int
	MemoryMB int
}

type VMService interface {
	CreateVM(createVMBody *CreateVMBody) (*VMCreationTask, error)
	ListTasks() ([]VMCreationTask, error)
	ListDeployments() ([]VMDeployment, error)
	ListVMs() ([]VM, error)
	UpdateDeploymentStatus()
}

type vmService struct {
	tasks         []VMCreationTask
	deployments   sync.Map
	vms           sync.Map
	updateChannel chan DeploymentUpdate
}

func NewVMService() VMService {
	return &vmService{
		tasks:         []VMCreationTask{},
		updateChannel: make(chan DeploymentUpdate),
	}
}

func (s *vmService) CreateVM(createVMBody *CreateVMBody) (*VMCreationTask, error) {
	var task VMCreationTask
	task.TaskID = uuid.New().String()

	deployment := s.createDeployment(createVMBody)
	task.DeploymentID = deployment.ID

	s.tasks = append(s.tasks, task)
	return &task, nil
}

func (s *vmService) createDeployment(createVMBody *CreateVMBody) *VMDeployment {
	var deployment VMDeployment
	deployment.ID = uuid.New().String()
	deployment.Status = "in-progress"

	s.deployments.Store(deployment.ID, deployment)
	go s.simulateVMCreation(deployment.ID, createVMBody)

	return &deployment
}

func (s *vmService) simulateVMCreation(deploymentID string, createVMBody *CreateVMBody) {
	vm := &VM{
		ID:       uuid.New().String(),
		Name:     createVMBody.Name,
		NumCPUs:  createVMBody.NumCPUs,
		MemoryMB: createVMBody.MemoryMB,
	}

	s.vms.Store(vm.ID, vm)
	s.updateChannel <- DeploymentUpdate{DeploymentID: deploymentID, ResourceID: vm.ID}
}

func (s *vmService) UpdateDeploymentStatus() {
	for update := range s.updateChannel {
		value, ok := s.deployments.Load(update.DeploymentID)
		if !ok {
			continue
		}
		deployment := value.(VMDeployment)
		deployment.Status = "created"
		s.deployments.Store(update.DeploymentID, deployment)
	}
}

func (s *vmService) ListTasks() ([]VMCreationTask, error) {
	return s.tasks, nil
}

func (s *vmService) ListDeployments() ([]VMDeployment, error) {
	var deployments []VMDeployment
	s.deployments.Range(func(key, value interface{}) bool {
		deployment := value.(VMDeployment)
		deployments = append(deployments, deployment)
		return true
	})
	return deployments, nil
}

func (s *vmService) ListVMs() ([]VM, error) {
	var vms []VM
	s.vms.Range(func(key, value interface{}) bool {
		vm := value.(VM)
		vms = append(vms, vm)
		return true
	})
	return vms, nil
}
