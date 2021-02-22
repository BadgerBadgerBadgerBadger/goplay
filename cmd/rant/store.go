package main

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/prologic/bitcask"
)

var db *bitcask.Bitcask

func GetAuthedUser(userID string) (AuthedUser, bool, error) {

	data, err := db.Get([]byte(userID))
	if err != nil {

		if err != bitcask.ErrKeyNotFound {
			return AuthedUser{}, false, errors.Wrap(err, "failed to retrieve data")
		}

		return AuthedUser{}, false, nil
	}

	authedUser := AuthedUser{}

	err = json.Unmarshal(data, &authedUser)
	if err != nil {
		return AuthedUser{}, false, errors.Wrap(err, "failed to unmarshall stored data data")
	}

	return authedUser, true, nil
}

func StoreAuthedUser(userID string, authedUser AuthedUser) error {

	data, err := json.Marshal(authedUser)
	if err != nil {
		return errors.Wrap(err, "failed to marshall data for storage")
	}

	err = db.Put([]byte(userID), data)
	if err != nil {
		return errors.Wrap(err, "failed to put data in database")
	}

	return nil
}
