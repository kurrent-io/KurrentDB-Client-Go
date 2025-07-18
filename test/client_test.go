package test

import "testing"

func TestStreams(t *testing.T) {
	t.Run("AppendEvents", TestAppendEventsSuite)
	t.Run("MultiAppend", TestMultiAppendEventsSuite)
	t.Run("ReadStream", TestReadStreamSuite)
	t.Run("ReadAll", TestReadAllSuite)
	t.Run("Subscription", TestSubscriptionSuite)
	t.Run("Delete", TestDeleteSuite)
	t.Run("Connection", TestConnectionSuite)
	t.Run("SecureAuthentication", TestSecureAuthenticationSuite)
	t.Run("InsecureAuthentication", TestInsecureAuthenticationSuite)
	t.Run("Rebalance", TestClusterSuite)
}

func TestPersistentSubscriptions(t *testing.T) {
	t.Run("PersistentSubscriptions", TestPersistentSubscriptionSuite)
	t.Run("PersistentSubscriptionsReadTests", TestPersistentSubscriptionReadSuite)
}

func TestMisc(t *testing.T) {
	t.Run("ConnectionString", TestConnectionStringSuite)
	t.Run("PositionParsing", TestPositionParsingSuite)
	t.Run("UuidParsing", TestUUIDParsingSuite)
}
