package recordings

import "github.com/gordonseto/soundvis-server/stream/models"

type RecordingTrackList struct {
	Tracklist []*RecordingSongTimestamp `bson:"tracklist"`
}

type RecordingSongTimestamp struct {
	Time int64	`bson:"time"`
	Song *stream.Song	`bson:"song"`
}

func NewRecordingTrackList() *RecordingTrackList {
	return &RecordingTrackList{make([]*RecordingSongTimestamp, 0)}
}

func NewRecordingSongTimestamp(time int64, song *stream.Song) *RecordingSongTimestamp {
	return &RecordingSongTimestamp{time, song}
}

func (rtl *RecordingTrackList) AddTimeStamp(timestamp *RecordingSongTimestamp) {
	rtl.Tracklist = append(rtl.Tracklist, timestamp)
}