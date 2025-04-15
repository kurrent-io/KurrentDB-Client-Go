package samples

import (
	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
)

func UserCertificates() {
	// region client-with-user-certificates
	settings, err := kurrentdb.ParseConnectionString("kurrentdb://admin:changeit@{endpoint}?tls=true&userCertFile={pathToCaFile}&userKeyFile={pathToKeyFile}")

	if err != nil {
		panic(err)
	}

	db, err := kurrentdb.NewClient(settings)
	// endregion client-with-user-certificates

	if err != nil {
		panic(err)
	}

	db.Close()
}
