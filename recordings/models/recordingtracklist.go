package recordings

import (
	"github.com/gordonseto/soundvis-server/stream/models"
)

type RecordingTrackList struct {
	RecordingId string 	`bson:"recordingId"`
	Tracklist []*RecordingSongTimestamp `bson:"tracklist"`
}

type RecordingSongTimestamp struct {
	Time int64	`bson:"time"`
	Song *stream.Song	`bson:"song"`
}

func NewRecordingTrackList(recordingId string) *RecordingTrackList {
	return &RecordingTrackList{recordingId, make([]*RecordingSongTimestamp, 0)}
}

func NewRecordingSongTimestamp(time int64, song *stream.Song) *RecordingSongTimestamp {
	return &RecordingSongTimestamp{time, song}
}

func (rtl *RecordingTrackList) AddTimeStamp(progress int64, song *stream.Song) {
	rtl.Tracklist = append(rtl.Tracklist, NewRecordingSongTimestamp(progress, song))
}