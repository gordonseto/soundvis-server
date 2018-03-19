package recordings

import "github.com/gordonseto/soundvis-server/stream/models"

type RecordingTrackList struct {
	Tracklist []*RecordingSongTimestamp `bson:"tracklist"`
}

type RecordingSongTimestamp struct {
	Time int64	`bson:"time"`
	Song *stream.Song	`bson:"song"`
}