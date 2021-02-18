package ocistore

import (
	"context"
	"time"

	snapshottypes "github.com/replicatedhq/kots/pkg/api/snapshot/types"
)

func (s OCIStore) ListPendingScheduledSnapshots(appID string) ([]snapshottypes.ScheduledSnapshot, error) {
	logger.Debug("Listing pending scheduled snapshots",
		zap.String("appID", appID))

	query := `SELECT id, app_id, scheduled_timestamp FROM scheduled_snapshots WHERE app_id = $1 AND backup_name IS NULL;`
	rows, err := s.connection.DB.Query(query, appID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query")
	}
	defer rows.Close()

	scheduledSnapshots := []snapshottypes.ScheduledSnapshot{}
	for rows.Next() {
		s := snapshottypes.ScheduledSnapshot{}
		if err := rows.Scan(&s.ID, &s.AppID, &s.ScheduledTimestamp); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}
		scheduledSnapshots = append(scheduledSnapshots, s)
	}

	return scheduledSnapshots, nil
}

func (s OCIStore) UpdateScheduledSnapshot(snapshotID string, backupName string) error {
	logger.Debug("Updating scheduled snapshot",
		zap.String("ID", snapshotID))

	query := `UPDATE scheduled_snapshots SET backup_name = $1 WHERE id = $2`
	_, err := s.connection.DB.Exec(query, backupName, snapshotID)
	if err != nil {
		return errors.Wrap(err, "failed to exec")
	}
	if err := ocidb.Commit(context.TODO(), s.connection); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}

func (s OCIStore) DeletePendingScheduledSnapshots(appID string) error {
	logger.Debug("Deleting pending scheduled snapshots",
		zap.String("appID", appID))

	query := `DELETE FROM scheduled_snapshots WHERE app_id = $1 AND backup_name IS NULL`
	_, err := s.connection.DB.Exec(query, appID)
	if err != nil {
		return errors.Wrap(err, "failed to db exec query")
	}
	if err := ocidb.Commit(context.TODO(), s.connection); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}

func (s OCIStore) CreateScheduledSnapshot(snapshotID string, appID string, timestamp time.Time) error {
	logger.Debug("Creating scheduled snapshot",
		zap.String("appID", appID))

	query := `INSERT INTO scheduled_snapshots (id, app_id, scheduled_timestamp) VALUES ($1, $2, $3)`

	_, err := s.connection.DB.Exec(query, snapshotID, appID, timestamp)
	if err != nil {
		return errors.Wrap(err, "Failed to db exec query")
	}
	if err := ocidb.Commit(context.TODO(), s.connection); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}

func (s OCIStore) ListPendingScheduledInstanceSnapshots(clusterID string) ([]snapshottypes.ScheduledInstanceSnapshot, error) {
	logger.Debug("Listing pending scheduled instance snapshots",
		zap.String("clusterID", clusterID))

	query := `SELECT id, cluster_id, scheduled_timestamp FROM scheduled_instance_snapshots WHERE cluster_id = $1 AND backup_name IS NULL;`
	rows, err := s.connection.DB.Query(query, clusterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query")
	}
	defer rows.Close()

	scheduledSnapshots := []snapshottypes.ScheduledInstanceSnapshot{}
	for rows.Next() {
		s := snapshottypes.ScheduledInstanceSnapshot{}
		if err := rows.Scan(&s.ID, &s.ClusterID, &s.ScheduledTimestamp); err != nil {
			return nil, errors.Wrap(err, "failed to scan")
		}
		scheduledSnapshots = append(scheduledSnapshots, s)
	}

	return scheduledSnapshots, nil
}

func (s OCIStore) UpdateScheduledInstanceSnapshot(snapshotID string, backupName string) error {
	logger.Debug("Updating scheduled instance snapshot",
		zap.String("ID", snapshotID))

	query := `UPDATE scheduled_instance_snapshots SET backup_name = $1 WHERE id = $2`
	_, err := s.connection.DB.Exec(query, backupName, snapshotID)
	if err != nil {
		return errors.Wrap(err, "failed to exec")
	}
	if err := ocidb.Commit(context.TODO(), s.connection); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}

func (s OCIStore) DeletePendingScheduledInstanceSnapshots(clusterID string) error {
	logger.Debug("Deleting pending scheduled instance snapshots",
		zap.String("clusterID", clusterID))

	query := `DELETE FROM scheduled_instance_snapshots WHERE cluster_id = $1 AND backup_name IS NULL`
	_, err := s.connection.DB.Exec(query, clusterID)
	if err != nil {
		return errors.Wrap(err, "failed to db exec query")
	}
	if err := ocidb.Commit(context.TODO(), s.connection); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}

func (s OCIStore) CreateScheduledInstanceSnapshot(snapshotID string, clusterID string, timestamp time.Time) error {
	logger.Debug("Creating scheduled instance snapshot",
		zap.String("clusterID", clusterID))

	query := `INSERT INTO scheduled_instance_snapshots (id, cluster_id, scheduled_timestamp) VALUES ($1, $2, $3)`
	_, err := s.connection.DB.Exec(query, snapshotID, clusterID, timestamp)
	if err != nil {
		return errors.Wrap(err, "Failed to db exec query")
	}
	if err := ocidb.Commit(context.TODO(), s.connection); err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	return nil
}
