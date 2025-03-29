package kurrentdb_test

import "testing"

func TestStreams(t *testing.T) {
	t.Run("AppendEvents", TestAppendEvents)
	t.Run("ReadStream", TestReadStream)
	t.Run("Subscription", TestSubscription)
	t.Run("Delete", TestDelete)
	t.Run("Connection", TestConnection)
	t.Run("SecureAuthentication", TestSecureAuthentication)
	t.Run("InsecureAuthentication", TestInsecureAuthentication)
	t.Run("Rebalance", TestCluster)
}

func TestPersistentSubscriptions(t *testing.T) {
	t.Run("PersistentSubscriptions", TestPersistentSub)
	t.Run("PersistentSubscriptionsReadTests", TestPersistentSubRead)
}

func TestMisc(t *testing.T) {
	t.Run("ConnectionString", TestConnectionString)
	t.Run("PositionParsing", TestPositionParsing)
	t.Run("UuidParsing", TestUUIDParsing)
}

func TestCI(t *testing.T) {
	t.Run("AppendEvents", TestAppendEvents)
	t.Run("ReadStream", TestReadStream)
	t.Run("Subscription", TestSubscription)
	t.Run("Delete", TestDelete)
	t.Run("Connection", TestConnection)
	t.Run("SecureAuthentication", TestSecureAuthentication)
	t.Run("InsecureAuthentication", TestInsecureAuthentication)
	t.Run("Rebalance", TestCluster)
	t.Run("PersistentSubscriptions", TestPersistentSub)
	t.Run("PersistentSubscriptionsReadTests", TestPersistentSubRead)
	t.Run("Projection", TestProjection)
	t.Run("ConnectionString", TestConnectionString)
	t.Run("PositionParsing", TestPositionParsing)
	t.Run("UuidParsing", TestUUIDParsing)
}
