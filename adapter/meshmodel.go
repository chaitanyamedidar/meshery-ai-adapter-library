package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/layer5io/meshkit/models/meshmodel/core/types"
	"github.com/layer5io/meshkit/models/meshmodel/core/v1alpha1"
	"github.com/layer5io/meshkit/models/meshmodel/registry"
)

// MeshModelRegistrantDefinitionPath - Structure for configuring registrant paths
type MeshModelRegistrantDefinitionPath struct {
	// EntityDefinitionPath holds the path for Entity Definition file
	EntityDefintionPath string

	Type types.CapabilityType
	// Host is the address of the gRPC host capable of processing the request
	Host string
	Port int
}

// MeshModel provides utility functions for registering
// MeshModel components to a registry in a reliable way
type MeshModelRegistrant struct {
	Paths        []MeshModelRegistrantDefinitionPath
	HTTPRegistry string
}

// NewMeshModelRegistrant returns an instance of NewMeshModelRegistrant
func NewMeshModelRegistrant(paths []MeshModelRegistrantDefinitionPath, httpRegistry string) *MeshModelRegistrant {
	return &MeshModelRegistrant{
		Paths:        paths,
		HTTPRegistry: httpRegistry,
	}
}

// Register will register each capability individually to the OAM Capability registry
//
// It sends a POST request to the endpoint in the "OAMHTTPRegistry", if the request
// fails then the request is retried. It uses exponential backoff algorithm to determine
// the interval between in the retries. It will retry only for 10 mins and will stop retrying
// after that.
//
// Register function is a blocking function
func (or *MeshModelRegistrant) Register(ctxID string) error {
	for _, dpath := range or.Paths {
		if dpath.Type != types.ComponentDefinition {
			continue
		}
		if err := or.registerComponentDefinition(ctxID, dpath); err != nil {
			return err
		}
	}

	return nil
}

func (or *MeshModelRegistrant) registerComponentDefinition(ctxID string, dpath MeshModelRegistrantDefinitionPath) error {
	entity, err := readComponentDefinition(dpath.EntityDefintionPath)
	if err != nil {
		return err
	}

	mrd := registry.MeshModelRegistrantData{
		Host: registry.Host{
			Hostname: dpath.Host,
			Port:     dpath.Port,
			Metadata: ctxID,
		},
		EntityType: dpath.Type,
		Entity:     entity,
	}

	backoffOpt := backoff.NewExponentialBackOff()
	backoffOpt.MaxElapsedTime = 10 * time.Minute
	if err := backoff.Retry(func() error {
		return or.postRegistration(mrd)
	}, backoffOpt); err != nil {
		return ErrOAMRetry(err)
	}

	return nil
}

func readComponentDefinition(path string) ([]byte, error) {
	definition, err := os.Open(path)
	if err != nil {
		return nil, ErrOpenOAMDefintionFile(err)
	}
	defer func() { _ = definition.Close() }()

	var cd v1alpha1.ComponentDefinition
	if err := json.NewDecoder(definition).Decode(&cd); err != nil {
		return nil, ErrJSONMarshal(err)
	}
	enbyt, err := json.Marshal(cd)
	if err != nil {
		return nil, ErrJSONMarshal(err)
	}
	return enbyt, nil
}

func (or *MeshModelRegistrant) postRegistration(mrd registry.MeshModelRegistrantData) error {
	contentByt, err := json.Marshal(mrd)
	if err != nil {
		return backoff.Permanent(err)
	}

	// host here is given by the application itself and is trustworthy hence,
	// #nosec
	resp, err := http.Post(or.HTTPRegistry, "application/json", bytes.NewReader(contentByt))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf(
			"register process failed, host returned status: %s with status code %d",
			resp.Status,
			resp.StatusCode,
		)
	}
	return nil
}
