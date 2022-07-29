package ledger

import (
	"github.com/owkin/orchestrator/lib/asset"
	"google.golang.org/protobuf/encoding/protojson"
)

// AddOrganization stores a new Organization
func (db *DB) AddOrganization(organization *asset.Organization) error {
	organizationBytes, err := marshaller.Marshal(organization)
	if err != nil {
		return err
	}
	err = db.putState(asset.OrganizationKind, organization.GetId(), organizationBytes)
	if err != nil {
		return err
	}

	return db.createIndex(allOrganizationsIndex, []string{asset.OrganizationKind, organization.Id})
}

// GetAllOrganizations returns all known Organizations
func (db *DB) GetAllOrganizations() ([]*asset.Organization, error) {
	elementKeys, err := db.getIndexKeys(allOrganizationsIndex, []string{asset.OrganizationKind})
	if err != nil {
		return nil, err
	}

	db.logger.WithField("numChildren", len(elementKeys)).Debug("GetAllOrganizations")

	organizations := []*asset.Organization{}
	for _, id := range elementKeys {
		organization, err := db.GetOrganization(id)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, organization)
	}

	return organizations, nil
}

// GetOrganization returns an organization by its ID
func (db *DB) GetOrganization(id string) (*asset.Organization, error) {
	n := asset.Organization{}

	b, err := db.getState(asset.OrganizationKind, id)
	if err != nil {
		return &n, err
	}

	err = protojson.Unmarshal(b, &n)
	return &n, err
}
