package dbal

import (
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/owkin/orchestrator/lib/asset"
	orcerrors "github.com/owkin/orchestrator/lib/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sqlNode struct {
	ID           string
	CreationDate time.Time
}

func (s *sqlNode) toNode() *asset.Node {
	return &asset.Node{
		Id:           s.ID,
		CreationDate: timestamppb.New(s.CreationDate),
	}
}

// AddNode implements persistence.NodeDBAL
func (d *DBAL) AddNode(node *asset.Node) error {
	stmt := getStatementBuilder().
		Insert("nodes").
		Columns("id", "channel", "creation_date").
		Values(node.GetId(), d.channel, node.GetCreationDate().AsTime())

	return d.exec(stmt)
}

// NodeExists implements persistence.NodeDBAL
func (d *DBAL) NodeExists(id string) (bool, error) {
	stmt := getStatementBuilder().
		Select("COUNT(id)").
		From("nodes").
		Where(sq.Eq{"id": id, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return false, err
	}

	var count int
	err = row.Scan(&count)

	return count == 1, err
}

// GetAllNodes implements persistence.NodeDBAL
func (d *DBAL) GetAllNodes() ([]*asset.Node, error) {
	stmt := getStatementBuilder().
		Select("id", "creation_date").
		From("nodes").
		Where(sq.Eq{"channel": d.channel})

	rows, err := d.query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*asset.Node

	for rows.Next() {
		scanned := sqlNode{}

		err = rows.Scan(&scanned.ID, &scanned.CreationDate)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, scanned.toNode())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

// GetNode implements persistence.NodeDBAL
func (d *DBAL) GetNode(id string) (*asset.Node, error) {
	stmt := getStatementBuilder().
		Select("id", "creation_date").
		From("nodes").
		Where(sq.Eq{"id": id, "channel": d.channel})

	row, err := d.queryRow(stmt)
	if err != nil {
		return nil, err
	}

	scanned := sqlNode{}
	err = row.Scan(&scanned.ID, &scanned.CreationDate)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orcerrors.NewNotFound("node", id)
		}
		return nil, err
	}

	return scanned.toNode(), nil
}
