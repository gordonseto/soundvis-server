import sys
if len(sys.argv) > 1:
    uid = sys.argv[1]
else:
    sys.exit(0)

import time
start_time = time.time()

# load model
from surprise import dump
file_name = 'recommendations/model'
_, loaded_algo = dump.load(file_name)

# connect to db
from pymongo import MongoClient
client = MongoClient("mongodb://notadmin:notpassword@ds117271.mlab.com:17271/soundvis")
db = client.soundvis
stations_repository = db["stations"]
listening_sessions_repository = db["listening_sessions"]

# fill dictionary with stations already listened to by user
stations_listened_to = {}
for document in listening_sessions_repository.find({"userId": uid}):
    station_id = str(document["stationId"])
    stations_listened_to[station_id] = True

# iterate through all stations and if user has not listened to it, predict rating
user_recs = []
for document in stations_repository.find({}):
    station_id = str(document["_id"])
    if station_id not in stations_listened_to:
        prediction = loaded_algo.predict(uid, station_id)
        user_recs.append(prediction)

# sort rating descending
user_recs.sort(key=lambda x: x.est, reverse=True)

from bson import ObjectId
def print_station_genre_and_rating(station_id, rating):
    station = stations_repository.find_one({"_id": ObjectId(station_id)})
    if 'genre' in station:
        print(station['genre'], rating)
    else:
        print("Genre not found for ", station_id)

end_time = time.time()

for rec in user_recs:
    print_station_genre_and_rating(rec.iid, rec.est)

print("Elapsed time: ", end_time - start_time)